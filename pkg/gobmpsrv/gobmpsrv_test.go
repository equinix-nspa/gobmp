package gobmpsrv

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mocking net.Conn
type mockedNetConn struct {
	mock.Mock
}

func (mnc *mockedNetConn) Close() error {
	args := mnc.Called()
	return args.Error(0)
}

func (mnc *mockedNetConn) LocalAddr() net.Addr {
	args := mnc.Called()
	return args.Get(0).(net.Addr)
}

func (mnc *mockedNetConn) Read(buf []byte) (int, error) {
	args := mnc.Called(buf)
	return args.Get(0).(int), args.Error(1)
}

func (mnc *mockedNetConn) RemoteAddr() net.Addr {
	args := mnc.Called()
	return args.Get(0).(net.Addr)
}

func (mnc *mockedNetConn) SetDeadline(time time.Time) error {
	args := mnc.Called(time)
	return args.Get(0).(error)
}

func (mnc *mockedNetConn) SetReadDeadline(time time.Time) error {
	args := mnc.Called(time)
	return args.Get(0).(error)
}

func (mnc *mockedNetConn) SetWriteDeadline(time time.Time) error {
	args := mnc.Called(time)
	return args.Get(0).(error)
}

func (mnc *mockedNetConn) Write(buf []byte) (int, error) {
	args := mnc.Called(buf)
	return args.Get(0).(int), args.Get(1).(error)
}

// Mocking net.Addr
type mockedNetAddr struct {
	mock.Mock
}

func (mna *mockedNetAddr) Network() string {
	args := mna.Called()
	return args.Get(0).(string)
}

func (mna *mockedNetAddr) String() string {
	args := mna.Called()
	return args.Get(0).(string)
}

func TestClientInfoAddDelIPv6(t *testing.T) {
	clientsInfo := newClientsInfo()

	mnc := mockedNetConn{}
	mna := mockedNetAddr{}

	// Mocked calls (once for Update and once for Delete)
	mnc.On("RemoteAddr").Return(&mna).Twice()
	mna.On("String").Return("[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:4").Twice()

	// Update(add)
	clientInfo, err := clientsInfo.Update(&mnc)
	require.Nil(t, err)
	require.NotNil(t, clientInfo)
	require.Equal(t, 1, len(clientsInfo.info))
	require.Equal(t, 1, len(clientInfo.ports))
	_, found := clientInfo.ports["4"]
	require.True(t, found)

	// Delete
	err = clientsInfo.Delete(&mnc)
	require.Equal(t, 0, len(clientsInfo.info))
	require.Nil(t, err)
}

func TestClientInfoAddDelIPv4(t *testing.T) {
	clientsInfo := newClientsInfo()

	mnc := mockedNetConn{}
	mna := mockedNetAddr{}

	// Mocked calls (once for Update and once for Delete)
	mnc.On("RemoteAddr").Return(&mna).Twice()
	mna.On("String").Return("192.100.10.1:4").Twice()

	// Update(add)
	clientInfo, err := clientsInfo.Update(&mnc)
	require.Nil(t, err)
	require.NotNil(t, clientInfo)
	require.Equal(t, 1, len(clientsInfo.info))
	require.Equal(t, 1, len(clientInfo.ports))
	_, found := clientInfo.ports["4"]
	require.True(t, found)

	// Delete
	err = clientsInfo.Delete(&mnc)
	require.Equal(t, 0, len(clientsInfo.info))
	require.Nil(t, err)
}

// Add/del multiple clients using same IP address (faking mutiple TCP conns from same BGP client)
func TestClientInfoAddDelMultIPv4SameAddr(t *testing.T) {
	clientsInfo := newClientsInfo()

	mnc1 := mockedNetConn{}
	mnc2 := mockedNetConn{}
	mna1 := mockedNetAddr{}
	mna2 := mockedNetAddr{}

	// Mocked calls (once for Update and once for Delete)
	mnc1.On("RemoteAddr").Return(&mna1).Twice()
	mnc2.On("RemoteAddr").Return(&mna2).Twice()
	mna1.On("String").Return("192.100.10.1:4").Twice()
	mna2.On("String").Return("192.100.10.1:5").Twice()

	// Update with 1st conn
	clientInfo, err := clientsInfo.Update(&mnc1)
	require.Nil(t, err)
	require.NotNil(t, clientInfo)
	require.Equal(t, 1, len(clientsInfo.info))
	require.Equal(t, 1, len(clientInfo.ports))
	_, found := clientInfo.ports["4"]
	require.True(t, found)

	// Update with 2nd conn
	clientInfo, err = clientsInfo.Update(&mnc2)
	require.Nil(t, err)
	require.NotNil(t, clientInfo)
	require.Equal(t, 1, len(clientsInfo.info))
	require.Equal(t, 2, len(clientInfo.ports))
	_, found = clientInfo.ports["5"]
	require.True(t, found)

	// Delete 1st conn
	err = clientsInfo.Delete(&mnc1)
	require.Equal(t, 1, len(clientsInfo.info))
	require.Equal(t, 1, len(clientInfo.ports))
	require.Nil(t, err)
	// 2nd conn still there
	_, found = clientInfo.ports["5"]
	require.True(t, found)
	// Delete 2nd conn
	err = clientsInfo.Delete(&mnc2)
	require.Equal(t, 0, len(clientsInfo.info))
	require.Nil(t, err)
}

// Add/del multiple clients using different IP address but same port (faking TCP conns from different BGP clients)
func TestClientInfoAddDelMultIPv4DiffAddrSamePort(t *testing.T) {
	clientsInfo := newClientsInfo()

	mnc1 := mockedNetConn{}
	mnc2 := mockedNetConn{}
	mna1 := mockedNetAddr{}
	mna2 := mockedNetAddr{}

	// Mocked calls (once for Update and once for Delete)
	mnc1.On("RemoteAddr").Return(&mna1).Twice()
	mnc2.On("RemoteAddr").Return(&mna2).Twice()
	mna1.On("String").Return("192.100.10.1:4").Twice()
	mna2.On("String").Return("192.100.10.2:4").Twice()

	// Update with 1st conn
	clientInfo, err := clientsInfo.Update(&mnc1)
	require.Nil(t, err)
	require.NotNil(t, clientInfo)
	require.Equal(t, 1, len(clientsInfo.info))
	require.Equal(t, 1, len(clientInfo.ports))
	_, found := clientInfo.ports["4"]
	require.True(t, found)

	// Update with 2nd conn
	clientInfo, err = clientsInfo.Update(&mnc2)
	require.Nil(t, err)
	require.NotNil(t, clientInfo)
	require.Equal(t, 2, len(clientsInfo.info))
	require.Equal(t, 1, len(clientInfo.ports))
	_, found = clientInfo.ports["4"]
	require.True(t, found)

	// Delete 1st conn
	err = clientsInfo.Delete(&mnc1)
	require.Equal(t, 1, len(clientsInfo.info))
	require.Nil(t, err)
	// 2nd conn still there
	_, found = clientInfo.ports["4"]
	require.True(t, found)
	// Delete 2nd conn
	err = clientsInfo.Delete(&mnc2)
	require.Equal(t, 0, len(clientsInfo.info))
	require.Nil(t, err)
}

// Add/del multiple clients using different IP address and different port (faking TCP conns from different BGP clients)
func TestClientInfoAddDelMultIPv4DiffAddrDiffPort(t *testing.T) {
	clientsInfo := newClientsInfo()

	mnc1 := mockedNetConn{}
	mnc2 := mockedNetConn{}
	mna1 := mockedNetAddr{}
	mna2 := mockedNetAddr{}

	// Mocked calls (once for Update and once for Delete)
	mnc1.On("RemoteAddr").Return(&mna1).Twice()
	mnc2.On("RemoteAddr").Return(&mna2).Twice()
	mna1.On("String").Return("192.100.10.1:4").Twice()
	mna2.On("String").Return("192.100.10.2:5").Twice()

	// Update with 1st conn
	clientInfo, err := clientsInfo.Update(&mnc1)
	require.Nil(t, err)
	require.NotNil(t, clientInfo)
	require.Equal(t, 1, len(clientsInfo.info))
	require.Equal(t, 1, len(clientInfo.ports))
	_, found := clientInfo.ports["4"]
	require.True(t, found)

	// Update with 2nd conn
	clientInfo, err = clientsInfo.Update(&mnc2)
	require.Nil(t, err)
	require.NotNil(t, clientInfo)
	require.Equal(t, 2, len(clientsInfo.info))
	require.Equal(t, 1, len(clientInfo.ports))
	_, found = clientInfo.ports["5"]
	require.True(t, found)

	// Delete 1st conn
	err = clientsInfo.Delete(&mnc1)
	require.Equal(t, 1, len(clientsInfo.info))
	require.Nil(t, err)
	// 2nd conn still there
	_, found = clientInfo.ports["5"]
	require.True(t, found)
	// Delete 2nd conn
	err = clientsInfo.Delete(&mnc2)
	require.Equal(t, 0, len(clientsInfo.info))
	require.Nil(t, err)
}

func TestClientInfoAddError(t *testing.T) {
	clientsInfo := newClientsInfo()

	mnc := mockedNetConn{}
	mna := mockedNetAddr{}

	// Mocked calls
	mnc.On("RemoteAddr").Return(&mna).Once()
	// This causes missing port error
	mna.On("String").Return("booboo").Once()

	// Update(add)
	clientInfo, err := clientsInfo.Update(&mnc)
	require.ErrorContains(t, err, "booboo: missing port in address")
	require.Nil(t, clientInfo)
	require.Equal(t, 0, len(clientsInfo.info))
}

func TestClientInfoADelError(t *testing.T) {
	clientsInfo := newClientsInfo()

	mnc1 := mockedNetConn{}
	mna1 := mockedNetAddr{}
	mna2 := mockedNetAddr{}

	// Mocked calls: we try to delete a different address
	mnc1.On("RemoteAddr").Return(&mna1).Once()
	mnc1.On("RemoteAddr").Return(&mna2).Once()
	mna1.On("String").Return("192.100.10.1:4").Once()
	mna2.On("String").Return("192.100.10.2:4").Once()

	// Update with 1st conn
	clientInfo, err := clientsInfo.Update(&mnc1)
	require.Nil(t, err)
	require.NotNil(t, clientInfo)
	require.Equal(t, 1, len(clientsInfo.info))
	require.Equal(t, 1, len(clientInfo.ports))
	_, found := clientInfo.ports["4"]
	require.True(t, found)

	// Delete 1st conn with wrong address (error)
	err = clientsInfo.Delete(&mnc1)
	require.ErrorContains(t, err, "192.100.10.2:4 not present")
	require.Equal(t, 1, len(clientsInfo.info))
}

// Mocking Publisher
type mockedPublisher struct {
	mock.Mock
}

func (mp *mockedPublisher) PublishMessage(msgType int, msgHash []byte, msg []byte) error {
	args := mp.Called(msgType, msgHash, msg)
	return args.Error(0)
}

func (mp *mockedPublisher) Stop() {
	mp.Called()
}
func TestBmpWorkerError(t *testing.T) {

	mnc := mockedNetConn{}
	mna := mockedNetAddr{}

	mp := mockedPublisher{}

	// Mocked calls (Read error)
	mnc.On("RemoteAddr").Return(&mna)
	mna.On("String").Return("192.100.10.1:4")
	mnc.On("Read", mock.Anything).Return(0, fmt.Errorf("No data from mock")).Once()
	mnc.On("Close").Return(nil).Once()

	bmpSrv := bmpServer{
		stop:            make(chan struct{}),
		sourcePort:      5000,
		destinationPort: 5001,
		intercept:       false,
		publisher:       &mp,
		splitAF:         true,
		storeData:       false,
		clientsInfo:     newClientsInfo(),
	}
	bmpSrv.bmpWorker(&mnc)

	require.Equal(t, 0, len(bmpSrv.clientsInfo.info))
}
