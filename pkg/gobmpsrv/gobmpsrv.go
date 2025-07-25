package gobmpsrv

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/bmp"
	"github.com/sbezverk/gobmp/pkg/message"
	"github.com/sbezverk/gobmp/pkg/parser"
	"github.com/sbezverk/gobmp/pkg/pub"
	"github.com/sbezverk/gobmp/pkg/store"
)

// BMPServer defines methods to manage BMP Server
type BMPServer interface {
	Start()
	Stop()
	GetStore() *store.Store
}

// Per-client info
type clientInfo struct {
	store *store.Store
}

func newClientInfo() *clientInfo {
	return &clientInfo{
		store: store.NewStore(),
	}
}

type clientsInfo struct {
	mutex sync.RWMutex
	info  map[string]clientInfo
}

func (c *clientsInfo) Add(clientRemoteAddr string, info clientInfo) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	glog.Infof("Adding client %s", clientRemoteAddr)
	if val, ok := c.info[clientRemoteAddr]; ok {
		return fmt.Errorf("%+v already present with %+v", clientRemoteAddr, val)
	}
	c.info[clientRemoteAddr] = info
	return nil
}

func (c *clientsInfo) Del(clientRemoteAddr string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	glog.Infof("Removing client %s", clientRemoteAddr)
	if _, ok := c.info[clientRemoteAddr]; !ok {
		return fmt.Errorf("%+v not present", clientRemoteAddr)
	}
	delete(c.info, clientRemoteAddr)
	return nil
}

func newClientsInfo() *clientsInfo {
	return &clientsInfo{
		info: make(map[string]clientInfo),
	}
}

type bmpServer struct {
	splitAF         bool
	intercept       bool
	storeData       bool
	publisher       pub.Publisher
	sourcePort      int
	destinationPort int
	incoming        net.Listener
	stop            chan struct{}
	clientsInfo     *clientsInfo
}

func (srv *bmpServer) Start() {
	// Starting bmp server server
	glog.Infof("Starting gobmp server on %s, intercept mode: %t, store-data: %t\n", srv.incoming.Addr().String(), srv.intercept, srv.storeData)
	go srv.server()
}

func (srv *bmpServer) Stop() {
	glog.Infof("Stopping gobmp server\n")
	if srv.publisher != nil {
		srv.publisher.Stop()
	}
	close(srv.stop)
}

func (srv *bmpServer) server() {
	for {
		client, err := srv.incoming.Accept()
		if err != nil {
			glog.Errorf("fail to accept client connection with error: %+v", err)
			continue
		}
		glog.V(5).Infof("client %+v accepted, calling bmpWorker", client.RemoteAddr())
		go srv.bmpWorker(client)
	}
}

func (srv *bmpServer) GetStore() *store.Store {
	if srv.clientsInfo == nil {
		return nil
	}
	// Pick the first client
	for _, client := range srv.clientsInfo.info {
		return client.store
	}
	return nil
}

func (srv *bmpServer) bmpWorker(client net.Conn) {
	defer func() {
		_ = client.Close()
	}()
	// Create new client info (keyed by client remote address)
	newClientInfo := newClientInfo()
	if err := srv.clientsInfo.Add(client.RemoteAddr().String(), *newClientInfo); err != nil {
		glog.Errorf("Failed to add client (already added) %s, %+v: %+v", client.RemoteAddr().String(), *newClientInfo, err)
	}

	var msgQueue chan interface{}
	var storeStop chan struct{}
	if srv.storeData {
		// We need a message queue to be able to store messages generated by the producer
		msgQueue = make(chan interface{})
		storeStop = make(chan struct{})
		// Start a goroutine to handle the messages from producer and store them
		go newClientInfo.store.Store(msgQueue, storeStop)
	}

	var server net.Conn
	var err error
	if srv.intercept {
		server, err = net.Dial("tcp", ":"+fmt.Sprintf("%d", srv.destinationPort))
		if err != nil {
			glog.Errorf("failed to connect to destination with error: %+v", err)
			return
		}
		defer func() { _ = server.Close() }()
		glog.V(5).Infof("connection to destination server %v established, start intercepting", server.RemoteAddr())
	}
	var producerQueue chan bmp.Message
	prod := message.NewProducer(srv.publisher, srv.splitAF, msgQueue)
	prodStop := make(chan struct{})
	producerQueue = make(chan bmp.Message)
	// Starting messages producer per client with dedicated work queue
	go prod.Producer(producerQueue, prodStop)

	parserQueue := make(chan []byte)
	parsStop := make(chan struct{})
	// Starting parser per client with dedicated work queue
	go parser.Parser(parserQueue, producerQueue, parsStop)
	defer func() {
		glog.V(5).Infof("all done with client %+v", client.RemoteAddr())
		close(parsStop)
		close(prodStop)
		if storeStop != nil {
			close(storeStop)
		}
		if err := srv.clientsInfo.Del(client.RemoteAddr().String()); err != nil {
			glog.Errorf("Failed to del client %s, %+v: %+v", client.RemoteAddr().String(), *newClientInfo, err)
		}

	}()
	for {
		headerMsg := make([]byte, bmp.CommonHeaderLength)
		if _, err := io.ReadAtLeast(client, headerMsg, bmp.CommonHeaderLength); err != nil {
			glog.Errorf("fail to read from client %+v with error: %+v", client.RemoteAddr(), err)
			return
		}
		// Recovering common header first
		header, err := bmp.UnmarshalCommonHeader(headerMsg[:bmp.CommonHeaderLength])
		if err != nil {
			glog.Errorf("fail to recover BMP message Common Header with error: %+v", err)
			continue
		}
		// Allocating space for the message body
		msg := make([]byte, int(header.MessageLength)-bmp.CommonHeaderLength)
		if _, err := io.ReadFull(client, msg); err != nil {
			glog.Errorf("fail to read from client %+v with error: %+v", client.RemoteAddr(), err)
			return
		}

		fullMsg := make([]byte, int(header.MessageLength))
		copy(fullMsg, headerMsg)
		copy(fullMsg[bmp.CommonHeaderLength:], msg)
		// Sending information to the server only in intercept mode
		if srv.intercept {
			if _, err := server.Write(fullMsg); err != nil {
				glog.Errorf("fail to write to server %+v with error: %+v", server.RemoteAddr(), err)
				return
			}
		}
		parserQueue <- fullMsg
	}
}

// NewBMPServer instantiates a new instance of BMP Server
func NewBMPServer(sPort, dPort int, intercept bool, p pub.Publisher, splitAF bool, storeData bool) (BMPServer, error) {
	incoming, err := net.Listen("tcp", fmt.Sprintf(":%d", sPort))
	if err != nil {
		glog.Errorf("fail to setup listener on port %d with error: %+v", sPort, err)
		return nil, err
	}
	bmp := bmpServer{
		stop:            make(chan struct{}),
		sourcePort:      sPort,
		destinationPort: dPort,
		intercept:       intercept,
		publisher:       p,
		incoming:        incoming,
		splitAF:         splitAF,
		storeData:       storeData,
		clientsInfo:     newClientsInfo(),
	}

	return &bmp, nil
}
