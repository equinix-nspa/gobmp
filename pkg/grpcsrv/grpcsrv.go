package grpcsrv

import (
	"context"
	"fmt"
	"net"

	"github.com/golang/glog"
	"github.com/sbezverk/gobmp/pkg/api/generated"
	"github.com/sbezverk/gobmp/pkg/gobmpsrv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServer represents a gRPC server implementation
type GRPCServer struct {
	server *grpc.Server
	port   string
}

// GRPCServiceRegistrar is a function type for registering services with a gRPC server
type GRPCServiceRegistrar func(*grpc.Server, gobmpsrv.BMPServer) error

// NewGRPCServer creates a new gRPC server instance
func NewGRPCServer(
	srv gobmpsrv.BMPServer,
	registrar GRPCServiceRegistrar,
) (*GRPCServer, error) {

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// register services
	if err := registrar(grpcServer, srv); err != nil {
		return nil, fmt.Errorf("failed to register services: %w", err)
	}

	return &GRPCServer{
		server: grpcServer,
		port:   "50001", // TBD pass parameter for this and have default value
	}, nil
}

// Start() starts the gRPC server
func (s *GRPCServer) Start() error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	glog.Infof("Starting gRPC server on port %s", s.port)

	// Start server in a goroutine
	go func() {
		if err := s.server.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			glog.Errorf("gRPC server failed", "error", err)
		}
	}()

	return nil
}

// Stop stops the gRPC server
func (s *GRPCServer) Stop(ctx context.Context) error {
	glog.Info("stopping gRPC server")
	stopped := make(chan struct{})

	go func() {
		s.server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		glog.Info("gRPC server stopped gracefully")
		return nil
	case <-ctx.Done():
		glog.Warning("gRPC server shutdown timed out, forcing stop")
		s.server.Stop()
		return ctx.Err()
	}
}

type StoreContentsServer struct {
	bmpsrv gobmpsrv.BMPServer
	generated.UnimplementedStoreContentsServiceServer
}

// This function is called from the code generated from proto
func (s *StoreContentsServer) Get(context.Context, *generated.GetRequest) (*generated.GetResponse, error) {
	srvStore := s.bmpsrv.GetStore()
	if srvStore == nil {
		glog.Warning("No store present on server")
		return nil, status.Error(codes.NotFound, "No store present on server")
	}
	bgplsStore := srvStore.GetBGPLS()

	response := &generated.GetResponse{
		BgpLs: GetBGPLS(bgplsStore),
	}
	glog.Infof("Get() => %d nodes, %d links", len(response.BgpLs.Nodes), len(response.BgpLs.Links))
	return response, nil
}

func NewStoreContentsServer(bmpsrv gobmpsrv.BMPServer) *StoreContentsServer {
	return &StoreContentsServer{bmpsrv: bmpsrv}
}
