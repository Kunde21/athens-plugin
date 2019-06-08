package backend

import (
	"context"
	"log"
	"net"
	"os"
	"os/exec"
	"time"

	"google.golang.org/grpc"

	stpb "github.com/Kunde21/athens-plugin/pb/v1/storage"
	"github.com/pkg/errors"
)

type Plugin struct {
	c    stpb.StorageBackendServiceClient
	conn *grpc.ClientConn
	canc context.CancelFunc
	cmd  *exec.Cmd
}

// NewPlugin storage backend
func NewPlugin(ctx context.Context, plugin, unixSock, config string) (Plugin, error) {
	bin, err := exec.LookPath(plugin)
	if err != nil {
		return Plugin{}, err
	}
	ctx, canc := context.WithCancel(ctx)
	cFunc := func() *exec.Cmd {
		cmd := exec.CommandContext(ctx, bin, []string{"-sock", unixSock, "-config", config}...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd
	}
	cmd := cFunc()
	if err := cmd.Start(); err != nil {
		log.Println(errors.Wrap(err, "plugin failed to initialize"))
		canc()
		return Plugin{}, err
	}
	conn, err := grpc.DialContext(ctx, unixSock,
		grpc.WithDialer(dialer),
		grpc.WithInsecure(),
		grpc.WithBackoffMaxDelay(1*time.Second),
		grpc.WithBlock(),
	)
	if err != nil {
		canc()
		return Plugin{}, err
	}
	p := Plugin{
		c:    stpb.NewStorageBackendServiceClient(conn),
		conn: conn,
		canc: canc,
		cmd:  cmd,
	}
	go p.watchPlugin(ctx, cFunc)
	return p, nil
}

func dialer(addr string, timeout time.Duration) (net.Conn, error) {
	return net.Dial("unix", addr)
}

// Close the connection to the plugin
func (p Plugin) Close() error {
	p.canc()
	return p.conn.Close()
}

// watchPlugin ensures that plugin is restarted if it crashes.
func (p Plugin) watchPlugin(ctx context.Context, f func() *exec.Cmd) {
	for {
		if err := p.cmd.Wait(); err != nil {
			log.Println(errors.Wrap(err, "plugin exited"))
		}
		select {
		case <-ctx.Done():
			return
		default:
		}
		p.cmd = f()
		if err := p.cmd.Start(); err != nil {
			log.Println(errors.Wrap(err, "plugin failed to start"))
		}
	}
}
