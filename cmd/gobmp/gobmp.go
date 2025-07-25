package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"net/http"
	_ "net/http/pprof"

	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/api/generated"
	"github.com/sbezverk/gobmp/pkg/dumper"
	"github.com/sbezverk/gobmp/pkg/filer"
	"github.com/sbezverk/gobmp/pkg/gobmpsrv"
	"github.com/sbezverk/gobmp/pkg/grpcsrv"
	"github.com/sbezverk/gobmp/pkg/kafka"
	"github.com/sbezverk/gobmp/pkg/nats"
	"github.com/sbezverk/gobmp/pkg/pub"
	"github.com/sbezverk/tools"
	"google.golang.org/grpc"
)

var (
	dstPort           int
	srcPort           int
	perfPort          int
	kafkaSrv          string
	kafkaTpRetnTimeMs string // Kafka topic retention time in ms
	natsSrv           string
	intercept         string
	splitAF           string
	dump              string
	file              string
	storeData         string
)

func init() {
	runtime.GOMAXPROCS(1)
	flag.IntVar(&srcPort, "source-port", 5000, "port exposed to outside")
	flag.IntVar(&dstPort, "destination-port", 5050, "port openBMP is listening")
	flag.StringVar(&kafkaSrv, "kafka-server", "", "URL to access Kafka server")
	flag.StringVar(&kafkaTpRetnTimeMs, "kafka-topic-retention-time-ms", "900000", "Kafka topic retention time in ms, default is 900000 ms i.e 15 minutes")
	flag.StringVar(&natsSrv, "nats-server", "", "URL to access NATS server")
	flag.StringVar(&intercept, "intercept", "false", "When intercept set \"true\", all incomming BMP messges will be copied to TCP port specified by destination-port, otherwise received BMP messages will be published to Kafka.")
	flag.StringVar(&splitAF, "split-af", "true", "When set \"true\" (default) ipv4 and ipv6 will be published in separate topics. if set \"false\" the same topic will be used for both address families.")
	flag.IntVar(&perfPort, "performance-port", 56767, "port used for performance debugging")
	flag.StringVar(&dump, "dump", "", "Dump resulting messages to file when \"dump=file\", to standard output when \"dump=console\" or to NATS when \"dump=nats\"")
	flag.StringVar(&file, "msg-file", "/tmp/messages.json", "Full path anf file name to store messages when \"dump=file\"")
	flag.StringVar(&storeData, "store-data", "false", "When store-data is set to \"true\", the supported (BGP-LS only for now) BMP state will be stored and accesible through API")
}

func main() {
	flag.Parse()
	_ = flag.Set("logtostderr", "true")
	// Starting performance collecting http server
	go func() {
		glog.Info(http.ListenAndServe(fmt.Sprintf(":%d", perfPort), nil))
	}()
	// Initializing publisher
	var publisher pub.Publisher
	var err error
	switch strings.ToLower(dump) {
	case "file":
		publisher, err = filer.NewFiler(file)
		if err != nil {
			glog.Errorf("failed to initialize file publisher with error: %+v", err)
			os.Exit(1)
		}
		glog.V(5).Infof("file publisher has been successfully initialized.")
	case "console":
		publisher, err = dumper.NewDumper()
		if err != nil {
			glog.Errorf("failed to initialize console publisher with error: %+v", err)
			os.Exit(1)
		}
		glog.V(5).Infof("console publisher has been successfully initialized.")
	case "nats":
		publisher, err = nats.NewPublisher(natsSrv)
		if err != nil {
			glog.Errorf("failed to initialize NATS publisher with error: %+v", err)
			os.Exit(1)
		}
		glog.V(5).Infof("NATS publisher has been successfully initialized.")
	default:
		kConfig := &kafka.Config{
			ServerAddress:        kafkaSrv,
			TopicRetentionTimeMs: kafkaTpRetnTimeMs,
		}
		publisher, err = kafka.NewKafkaPublisher(kConfig)
		if err != nil {
			glog.Errorf("failed to initialize Kafka publisher with error: %+v", err)
			os.Exit(1)
		}
		glog.V(5).Infof("Kafka publisher has been successfully initialized.")
	}

	// Initializing bmp server
	interceptFlag, err := strconv.ParseBool(intercept)
	if err != nil {
		glog.Errorf("failed to parse to bool the value of the intercept flag with error: %+v", err)
		os.Exit(1)
	}
	splitAFFlag, err := strconv.ParseBool(splitAF)
	if err != nil {
		glog.Errorf("failed to parse to bool the value of the intercept flag with error: %+v", err)
		os.Exit(1)
	}
	storeDataFlag, err := strconv.ParseBool(storeData)
	if err != nil {
		glog.Errorf("failed to parse to bool the value of the store-data flag with error: %+v", err)
		os.Exit(1)
	}
	bmpSrv, err := gobmpsrv.NewBMPServer(srcPort, dstPort, interceptFlag, publisher, splitAFFlag, storeDataFlag)
	if err != nil {
		glog.Errorf("failed to setup new gobmp server with error: %+v", err)
		os.Exit(1)
	}
	// Starting Interceptor server
	bmpSrv.Start()

	// Create gRPC server for store services
	grpcSrv, err := grpcsrv.NewGRPCServer(bmpSrv, registerGRPCStoreServices)
	if err != nil {
		glog.Errorf("failed to setup new grpc server with error: %+v", err)
		os.Exit(1)
	}
	err = grpcSrv.Start()
	if err != nil {
		glog.Errorf("failed to start grpc server with error: %+v", err)
		os.Exit(1)
	}

	stopCh := tools.SetupSignalHandler()
	<-stopCh

	bmpSrv.Stop()
	err = grpcSrv.Stop(context.Background())
	if err != nil {
		glog.Errorf("failed to stop grpc server with error: %+v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

// registerGRPCStoreServices is responsible for instantiating the gRPC store services and to register them with the gRPC server
func registerGRPCStoreServices(s *grpc.Server, bmpsrv gobmpsrv.BMPServer) error {
	// Create & register StoreContents service server
	storeContentsServer := grpcsrv.NewStoreContentsServer(bmpsrv)
	generated.RegisterStoreContentsServiceServer(s, storeContentsServer)

	return nil
}
