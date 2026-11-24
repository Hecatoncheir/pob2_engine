package cli

import (
	"context"
	"fmt"
	"io"
	"time"

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
	session, installation, err := openWorkbench(ctx)
	if err != nil {
		return err
	}
	defer session.Close()
	if *useGRPC {
		return r.serveWorkbench(ctx, session, installation)
	}
	server, err := mcpapi.NewWorkbench(session, installation)
	if err != nil {
		return err
	}
	return server.Run(ctx, &mcp.IOTransport{Reader: r.in, Writer: writer{r.out}})
}

func (r runner) serveWorkbench(ctx context.Context, session *pob.Workbench, installation pob.Installation) error {
	service, err := workbench.New(session, installation)
	if err != nil {
		return err
	}
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

func openWorkbench(ctx context.Context) (*pob.Workbench, pob.Installation, error) {
	manager, err := pob.Local()
	if err != nil {
		return nil, pob.Installation{}, err
	}
	installCtx, cancel := context.WithTimeout(ctx, pob.InstallTimeout)
	defer cancel()
	installation, err := manager.Ensure(installCtx)
	if err != nil {
		return nil, installation, err
	}
	ready, cancelReady := context.WithTimeout(ctx, 60*time.Second)
	defer cancelReady()
	session, err := pob.NewWorkbench(ready, installation)
	return session, installation, err
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
