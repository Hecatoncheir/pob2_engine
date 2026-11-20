package grpcapi

import (
	"context"
	"fmt"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/buildinfo"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	conn *grpc.ClientConn
	rpc  pb.SessionServiceClient
}

func Dial(ctx context.Context, address, token string) (*Client, error) {
	if err := validateAddress(address); err != nil {
		return nil, err
	}
	if token == "" {
		return nil, fmt.Errorf("нужен токен сервера")
	}
	conn, err := grpc.NewClient("passthrough:///"+address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoke grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		return invoke(ctx, method, req, reply, cc, opts...)
	}))
	if err != nil {
		return nil, err
	}
	c := &Client{conn: conn, rpc: pb.NewSessionServiceClient(conn)}
	info, err := c.Info(ctx)
	if err != nil {
		conn.Close()
		return nil, err
	}
	if info.GetProtocolVersion() != buildinfo.ProtocolVersion {
		conn.Close()
		return nil, fmt.Errorf("несовместимый протокол %q; нужен %q", info.GetProtocolVersion(), buildinfo.ProtocolVersion)
	}
	return c, nil
}

func (c *Client) Close() error { return c.conn.Close() }

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
