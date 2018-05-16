// Copyright © 2017 JB Ribeiro <self@vredens.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package api

import (
	"net"

	"gitlab.com/mandalore/go-app/app"
	log "gitlab.com/vredens/go-logger"

	"google.golang.org/grpc"
)

// GRPC controller.
type GRPC struct {
	Server  *grpc.Server
	address string
	running bool
}

// NewGRPC creates a new gRPC controller and respective API.
// TODO: convert parameters to a Configuration struct
func NewGRPC(address string) *GRPC {
	opts := make([]grpc.ServerOption, 1)
	opts[0] = grpc.MaxConcurrentStreams(10)

	// TODO: add support for TLS
	// creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
	// if err != nil {
	// 	grpclog.Fatalf("Failed to generate credentials %v", err)
	// }
	// opts[1] = grpc.Creds(creds)

	controller := &GRPC{
		Server:  grpc.NewServer(opts...),
		address: address,
		running: false,
	}

	return controller
}

// Start initializes the gRPC server. This is a blocking operation.
func (rpc *GRPC) Start() error {
	bind, err := net.Listen("tcp", rpc.address)
	if err != nil {
		return app.NewError(app.ErrorUnexpected, "failed to start gRPC server", err)
	}

	l := app.Logger.Spawn(log.WithFields(map[string]interface{}{"addr": rpc.address}))

	l.Info("starting gRPC server")

	rpc.running = true
	if err := rpc.Server.Serve(bind); err != nil {
		rpc.running = false
		return err
	}
	rpc.running = false
	l.Info("stopped gRPC server")

	return nil
}

// Stop attempts to shutdown gracefully.
func (rpc *GRPC) Stop() error {
	if rpc.running != true {
		return app.NewError(app.ErrorDevPoo, "rpc server is already stopped", nil)
	}

	rpc.Server.GracefulStop()
	rpc.running = false

	return nil
}

// IsRunning returns true if the gRPC service is running and listening for connections.
func (rpc *GRPC) IsRunning() bool {
	return rpc.running
}
