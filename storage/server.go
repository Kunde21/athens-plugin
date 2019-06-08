package storage

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/gomods/athens/pkg/storage"
	"google.golang.org/grpc"

	stpb "github.com/Kunde21/athens-plugin/pb/v1/storage"
)

type Plugin struct {
	srv *grpc.Server
	lis net.Listener
}

var socket string

func init() {
	flag.StringVar(&socket, "sock", "/tmp/storage.sock", "plugin unix socket name")
}

func NewPlugin(back storage.Backend, opts ...Option) (Plugin, error) {
	flag.Parse()

	bk := backend{b: back}
	socket, err := filepath.Abs(socket)
	if err != nil {
		return Plugin{}, err
	}

	srv := grpc.NewServer()
	stpb.RegisterStorageBackendServiceServer(srv, bk)

	_ = os.Remove(socket) // don't fail if not found, net.Listen will check
	lis, err := net.Listen("unix", socket)
	if err != nil {
		return Plugin{}, err
	}
	return Plugin{srv: srv, lis: lis}, nil
}

// Close the server
func (p Plugin) Close() error {
	fmt.Println("closing plugin")
	p.srv.Stop()
	return p.lis.Close()
}

// Serve the storage plugin interface
func (p Plugin) Serve() error {
	return p.srv.Serve(p.lis)
}

type Option func(*config)

type config struct {
	srv  *grpc.Server
	opts []grpc.ServerOption
}

func WithServerOptions(opts ...grpc.ServerOption) Option {
	return func(cfg *config) {
		cfg.opts = append(cfg.opts, opts...)
	}
}
