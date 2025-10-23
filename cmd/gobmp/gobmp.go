package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"

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
	dstPort            int
	srcPort            int
	perfPort           int
	kafkaSrv           string
	kafkaTpRetnTimeMs  string // Kafka topic retention time in ms
	natsSrv            string
	intercept          string
	splitAF            string
	dump               string
	file               string
	storeData          string
	kafkaSASLEnable    bool
	kafkaSASLMechanism string
	kafkaSASLUsername  string
	kafkaSASLPassword  string
)

func init() {
	runtime.GOMAXPROCS(1)
	// Set defaults
	srcPort = 5000
	dstPort = 5050
	kafkaTpRetnTimeMs = "900000"
	intercept = "false"
	splitAF = "true"
	perfPort = 56767
	file = "/tmp/messages.json"
	kafkaSASLEnable = false
	storeData = "false"

	applyEnvOverrides()

	// Flags (CLI overrides env/defaults)
	flag.IntVar(&srcPort, "source-port", srcPort, "port exposed to outside")
	flag.IntVar(&dstPort, "destination-port", dstPort, "port openBMP is listening")
	flag.StringVar(&kafkaSrv, "kafka-server", kafkaSrv, "URL to access Kafka server")
	flag.StringVar(&kafkaTpRetnTimeMs, "kafka-topic-retention-time-ms", kafkaTpRetnTimeMs, "Kafka topic retention time in ms, default is 900000 ms i.e 15 minutes")
	flag.StringVar(&natsSrv, "nats-server", natsSrv, "URL to access NATS server")
	flag.StringVar(&intercept, "intercept", intercept, "When intercept set \"true\", all incomming BMP messges will be copied to TCP port specified by destination-port, otherwise received BMP messages will be published to Kafka.")
	flag.StringVar(&splitAF, "split-af", splitAF, "When set \"true\" (default) ipv4 and ipv6 will be published in separate topics. if set \"false\" the same topic will be used for both address families.")
	flag.IntVar(&perfPort, "performance-port", perfPort, "port used for performance debugging")
	flag.StringVar(&dump, "dump", dump, "Dump resulting messages to file when \"dump=file\", to standard output when \"dump=console\" or to NATS when \"dump=nats\"")
	flag.StringVar(&file, "msg-file", file, "Full path anf file name to store messages when \"dump=file\"")
	flag.BoolVar(&kafkaSASLEnable, "kafka-sasl-enable", kafkaSASLEnable, "Enable SASL authentication for Kafka producer")
	flag.StringVar(&kafkaSASLMechanism, "kafka-sasl-mechanism", kafkaSASLMechanism, "SASL mechanism for Kafka producer (e.g., PLAIN, SCRAM-SHA-256, SCRAM-SHA-512)")
	flag.StringVar(&kafkaSASLUsername, "kafka-sasl-username", kafkaSASLUsername, "SASL username for Kafka producer")
	flag.StringVar(&kafkaSASLPassword, "kafka-sasl-password", kafkaSASLPassword, "SASL password for Kafka producer")
	flag.StringVar(&storeData, "store-data", storeData, "When store-data is set to \"true\", the supported (BGP-LS only for now) BMP state will be stored and accesible through API")
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
			SASLEnable:           kafkaSASLEnable,
			SASLMechanism:        kafkaSASLMechanism,
			SASLUsername:         kafkaSASLUsername,
			SASLPassword:         kafkaSASLPassword,
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

// applyEnvOverrides sets global config variables from environment variables if present.
func applyEnvOverrides() {
	if val := os.Getenv("GOBMP_SOURCE_PORT"); val != "" {
		if v, err := strconv.Atoi(val); err == nil {
			srcPort = v
		}
	}
	if val := os.Getenv("GOBMP_DESTINATION_PORT"); val != "" {
		if v, err := strconv.Atoi(val); err == nil {
			dstPort = v
		}
	}
	if val := os.Getenv("GOBMP_KAFKA_SERVER"); val != "" {
		kafkaSrv = val
	}
	if val := os.Getenv("GOBMP_KAFKA_TOPIC_RETENTION_TIME_MS"); val != "" {
		kafkaTpRetnTimeMs = val
	}
	if val := os.Getenv("GOBMP_NATS_SERVER"); val != "" {
		natsSrv = val
	}
	if val := os.Getenv("GOBMP_INTERCEPT"); val != "" {
		intercept = val
	}
	if val := os.Getenv("GOBMP_SPLIT_AF"); val != "" {
		splitAF = val
	}
	if val := os.Getenv("GOBMP_PERFORMANCE_PORT"); val != "" {
		if v, err := strconv.Atoi(val); err == nil {
			perfPort = v
		}
	}
	if val := os.Getenv("GOBMP_DUMP"); val != "" {
		dump = val
	}
	if val := os.Getenv("GOBMP_MSG_FILE"); val != "" {
		file = val
	}
	if val := os.Getenv("GOBMP_KAFKA_SASL_ENABLE"); val != "" {
		kafkaSASLEnable = val == "true" || val == "1"
	}
	if val := os.Getenv("GOBMP_KAFKA_SASL_MECHANISM"); val != "" {
		kafkaSASLMechanism = val
	}
	if val := os.Getenv("GOBMP_KAFKA_SASL_USERNAME"); val != "" {
		kafkaSASLUsername = val
	}
	if val := os.Getenv("GOBMP_KAFKA_SASL_PASSWORD"); val != "" {
		kafkaSASLPassword = val
	}
}
