package grpcapi

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"net"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/pob"
	"github.com/Hecatoncheir/pob2_engine/internal/workbench"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type WorkbenchReady struct {
	Address string `json:"address"`
	Token   string `json:"token"`
}

type workbenchServer struct {
	pb.UnimplementedWorkbenchServiceServer
	service *workbench.Service
}

func ServeWorkbench(ctx context.Context, service *workbench.Service, ready func(WorkbenchReady) error) error {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer listener.Close()
	token := rand.Text()
	server := grpc.NewServer(grpc.MaxRecvMsgSize(maxMessageSize), grpc.MaxSendMsgSize(maxMessageSize),
		grpc.UnaryInterceptor(workbenchAuthorization(token)))
	pb.RegisterWorkbenchServiceServer(server, &workbenchServer{service: service})
	stop := context.AfterFunc(ctx, server.Stop)
	defer stop()
	if err := ready(WorkbenchReady{Address: listener.Addr().String(), Token: token}); err != nil {
		return err
	}
	err = server.Serve(listener)
	if errors.Is(err, grpc.ErrServerStopped) {
		return nil
	}
	return err
}

func workbenchAuthorization(token string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, request any, _ *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		values := metadata.ValueFromIncomingContext(ctx, "authorization")
		if len(values) != 1 || subtle.ConstantTimeCompare([]byte(values[0]), []byte("Bearer "+token)) != 1 {
			return nil, status.Error(codes.Unauthenticated, "неверный токен рабочей копии")
		}
		return next(ctx, request)
	}
}

func (s *workbenchServer) ListWorkbenchTools(context.Context, *pb.ListWorkbenchToolsRequest) (*pb.ListWorkbenchToolsResponse, error) {
	tools := s.service.Tools()
	names := make([]string, len(tools))
	for i, tool := range tools {
		names[i] = tool.Name
	}
	return &pb.ListWorkbenchToolsResponse{Names: names}, nil
}

func (s *workbenchServer) CallWorkbenchTool(ctx context.Context, request *pb.CallWorkbenchToolRequest) (*pb.CallWorkbenchToolResponse, error) {
	response, err := s.service.Invoke(ctx, request.GetName(), request.GetArgumentsJson())
	if err == nil {
		return &pb.CallWorkbenchToolResponse{Revision: uint64(response.Result.Revision), ResultJson: response.Result.Result, Png: response.PNG}, nil
	}
	var failure *pob.WorkbenchError
	if errors.As(err, &failure) {
		return &pb.CallWorkbenchToolResponse{Error: &pb.WorkbenchToolError{Code: failure.Code, Message: failure.Message}}, nil
	}
	return nil, status.Error(codes.Internal, err.Error())
}
