package grpcapi

import (
	"context"
	"errors"

	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/lifecycle"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func lifecycleError(err error) error {
	code := codes.InvalidArgument
	switch {
	case errors.Is(err, lifecycle.ErrDraining):
		code = codes.Unavailable
	case errors.Is(err, lifecycle.ErrLeaseExpired):
		code = codes.FailedPrecondition
	case errors.Is(err, lifecycle.ErrClientConflict):
		code = codes.AlreadyExists
	case errors.Is(err, lifecycle.ErrClientLimit):
		code = codes.ResourceExhausted
	}
	return status.Error(code, err.Error())
}

func rpcError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return status.FromContextError(err).Err()
	}
	var app *application.Error
	if !errors.As(err, &app) {
		return status.Error(codes.Internal, "внутренняя ошибка сервиса")
	}
	code := codes.Internal
	switch app.Code {
	case application.InvalidArgument:
		code = codes.InvalidArgument
	case application.NotFound:
		code = codes.NotFound
	case application.AlreadyExists:
		code = codes.AlreadyExists
	case application.RevisionConflict:
		code = codes.Aborted
	case application.RequestConflict:
		code = codes.FailedPrecondition
	case application.ResourceExhausted:
		code = codes.ResourceExhausted
	}
	return status.Error(code, app.Error())
}
