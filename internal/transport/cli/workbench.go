package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/Hecatoncheir/pob2_engine/internal/pob"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/grpcapi"
	"github.com/Hecatoncheir/pob2_engine/internal/transport/mcpapi"
	"github.com/Hecatoncheir/pob2_engine/internal/workbench"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (r runner) workbench(ctx context.Context, args []string) error {
	flags := r.flags("workbench")
	useGRPC := flags.Bool("grpc", false, "использовать локальный gRPC вместо MCP stdio")
	if err := parse(flags, args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return &UsageError{Err: fmt.Errorf("workbench не принимает позиционные параметры")}
	}
	installation, err := workbenchInstallation(ctx)
	if err != nil {
		return err
	}
	service, err := workbench.New(installation)
	if err != nil {
		return err
	}
	defer service.Close()
	if *useGRPC {
		return r.serveWorkbench(ctx, service)
	}
	if err := service.Prepare(ctx); err != nil {
		return err
	}
	server := mcpapi.NewWorkbench(service)
	return server.Run(ctx, &mcp.IOTransport{Reader: r.in, Writer: writer{r.out}})
}

func (r runner) serveWorkbench(ctx context.Context, service *workbench.Service) error {
	ctx, stop := context.WithCancel(ctx)
	defer stop()
	go func() {
		_, _ = io.Copy(io.Discard, r.in)
		stop()
	}()
	return grpcapi.ServeWorkbench(ctx, service, func(ready grpcapi.WorkbenchReady) error {
		return r.write(ready)
	})
}

func workbenchInstallation(ctx context.Context) (pob.Installation, error) {
	manager, err := pob.Local()
	if err != nil {
		return pob.Installation{}, err
	}
	installCtx, cancel := context.WithTimeout(ctx, pob.InstallTimeout)
	defer cancel()
	return manager.Ensure(installCtx)
}

func (r runner) localEditor(ctx context.Context, args []string) error {
	flags := r.flags("editor")
	endpoint := flags.String("endpoint", "", "файл подключения к открытому редактору PoB II")
	if err := parse(flags, args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *endpoint == "" {
		return &UsageError{Err: fmt.Errorf("укажите editor --endpoint ФАЙЛ")}
	}
	server, err := mcpapi.NewLocalEditor(*endpoint)
	if err != nil {
		return err
	}
	return server.Run(ctx, &mcp.IOTransport{Reader: r.in, Writer: writer{r.out}})
}
