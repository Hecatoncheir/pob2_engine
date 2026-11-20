package grpcapi

import (
	"context"
	"crypto/rand"
	"fmt"
	"slices"
	"sync"
	"time"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/buildinfo"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Client struct {
	conn          *grpc.ClientConn
	rpc           pb.SessionServiceClient
	leaseID       string
	instanceID    string
	cancel        context.CancelFunc
	heartbeatDone chan struct{}
	lost          chan struct{}
	lostOnce      sync.Once
	closeOnce     sync.Once
	closeErr      error
}

func Dial(ctx context.Context, address, token string) (*Client, error) {
	return DialKind(ctx, address, token, "cli")
}

func DialKind(ctx context.Context, address, token, kind string) (*Client, error) {
	c, err := open(address, token)
	if err != nil {
		return nil, err
	}
	if err := c.initialize(ctx, kind); err != nil {
		c.conn.Close()
		return nil, err
	}
	return c, nil
}

// Inspect проверяет готовность без регистрации клиента и продления жизни сервера.
func Inspect(ctx context.Context, address, token string) (*pb.GetServerInfoResponse, error) {
	c, err := open(address, token)
	if err != nil {
		return nil, err
	}
	defer c.conn.Close()
	return c.Info(ctx)
}

func open(address, token string) (*Client, error) {
	if err := ValidateAddress(address); err != nil {
		return nil, err
	}
	if token == "" {
		return nil, fmt.Errorf("нужен токен сервера")
	}
	c := &Client{heartbeatDone: make(chan struct{}), lost: make(chan struct{}), cancel: func() {}}
	conn, err := grpc.NewClient("passthrough:///"+address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoke grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		if c.leaseID != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "pob2-lease", c.leaseID)
		}
		return invoke(ctx, method, req, reply, cc, opts...)
	}), grpc.WithStreamInterceptor(func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, next grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token, "pob2-lease", c.leaseID)
		return next(ctx, desc, cc, method, opts...)
	}))
	if err != nil {
		return nil, err
	}
	c.conn, c.rpc = conn, pb.NewSessionServiceClient(conn)
	return c, nil
}

func (c *Client) initialize(ctx context.Context, kind string) error {
	info, err := c.Info(ctx)
	if err != nil {
		return err
	}
	if info.GetProtocolVersion() != buildinfo.ProtocolVersion {
		return fmt.Errorf("несовместимый протокол %q; нужен %q", info.GetProtocolVersion(), buildinfo.ProtocolVersion)
	}
	if info.GetState() != "ready" {
		return status.Error(codes.Unavailable, "сервер завершает работу")
	}
	c.instanceID = info.GetInstanceId()
	if !slices.Contains(info.GetCapabilities(), "client.leases") {
		close(c.heartbeatDone)
		return nil
	}
	lease, err := c.rpc.RegisterClient(ctx, &pb.RegisterClientRequest{ClientId: rand.Text(), Kind: kind})
	if err != nil {
		return err
	}
	if lease.GetLeaseTimeoutMs() < 100 || lease.GetLeaseTimeoutMs() > uint64((24*time.Hour).Milliseconds()) {
		return fmt.Errorf("сервер вернул недопустимый срок аренды")
	}
	c.leaseID = lease.GetLeaseId()
	lifeCtx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	go c.heartbeat(lifeCtx, time.Duration(lease.GetLeaseTimeoutMs())*time.Millisecond)
	return nil
}

func (c *Client) Done() <-chan struct{} { return c.lost }
func (c *Client) InstanceID() string    { return c.instanceID }
func (c *Client) LeaseID() string       { return c.leaseID }

func (c *Client) Watch(ctx context.Context, id string, after uint64) (grpc.ServerStreamingClient[pb.WatchSessionResponse], error) {
	return c.rpc.WatchSession(ctx, &pb.WatchSessionRequest{SessionId: id, AfterRevision: after})
}

func (c *Client) heartbeat(ctx context.Context, ttl time.Duration) {
	defer close(c.heartbeatDone)
	timer := time.NewTicker(ttl / 3)
	defer timer.Stop()
	lastAck := time.Now()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			callCtx, cancel := context.WithTimeout(ctx, min(ttl/3, 2*time.Second))
			_, err := c.rpc.Heartbeat(callCtx, &pb.HeartbeatRequest{LeaseId: c.leaseID})
			cancel()
			if err == nil {
				lastAck = time.Now()
				continue
			}
			if status.Code(err) == codes.FailedPrecondition || time.Since(lastAck) >= ttl {
				c.conn.Close()
				c.lostOnce.Do(func() { close(c.lost) })
				return
			}
		}
	}
}

func (c *Client) Close() error {
	c.closeOnce.Do(func() {
		c.cancel()
		<-c.heartbeatDone
		if c.leaseID != "" {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			_, _ = c.rpc.ReleaseClient(ctx, &pb.ReleaseClientRequest{LeaseId: c.leaseID})
			cancel()
		}
		c.closeErr = c.conn.Close()
		c.lostOnce.Do(func() { close(c.lost) })
	})
	return c.closeErr
}

func (c *Client) Info(ctx context.Context) (*pb.GetServerInfoResponse, error) {
	return c.rpc.GetServerInfo(ctx, &pb.GetServerInfoRequest{})
}

func fromSnapshot(value *pb.SessionSnapshot) domain.Snapshot {
	return domain.Snapshot{SessionID: value.GetSessionId(), BuildName: value.GetBuildName(), Revision: value.GetRevision()}
}

func (c *Client) Create(ctx context.Context, cmd application.CreateSession) (domain.Snapshot, error) {
	res, err := c.rpc.CreateSession(ctx, &pb.CreateSessionRequest{SessionId: cmd.SessionID, BuildName: cmd.BuildName, RequestId: cmd.RequestID})
	if err != nil {
		return domain.Snapshot{}, err
	}
	return fromSnapshot(res.GetSnapshot()), nil
}

func (c *Client) Snapshot(ctx context.Context, id string) (domain.Snapshot, error) {
	res, err := c.rpc.GetSnapshot(ctx, &pb.GetSnapshotRequest{SessionId: id})
	if err != nil {
		return domain.Snapshot{}, err
	}
	return fromSnapshot(res.GetSnapshot()), nil
}

func (c *Client) Rename(ctx context.Context, cmd application.RenameBuild) (domain.Snapshot, error) {
	res, err := c.rpc.RenameBuild(ctx, &pb.RenameBuildRequest{SessionId: cmd.SessionID, BuildName: cmd.BuildName, ExpectedRevision: cmd.ExpectedRevision, RequestId: cmd.RequestID})
	if err != nil {
		return domain.Snapshot{}, err
	}
	return fromSnapshot(res.GetSnapshot()), nil
}

func (c *Client) Shutdown(ctx context.Context) error {
	_, err := c.rpc.Shutdown(ctx, &pb.ShutdownRequest{})
	return err
}
