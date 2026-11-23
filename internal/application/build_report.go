package application

import (
	"context"
	"encoding/json"

	"github.com/Hecatoncheir/pob2_engine/internal/build"
)

type BuildReportView struct {
	Revision uint64          `json:"revision"`
	Report   json.RawMessage `json:"report"`
}

type BuildReportCommands interface {
	BuildReport(context.Context, string, uint64) (BuildReportView, error)
}

type buildReportReader interface {
	ReadBuildReport(context.Context, *build.Document, string) (json.RawMessage, error)
}

func (s *Service) SupportsBuildReport() bool {
	_, ok := s.calculator.(buildReportReader)
	return ok
}

func (s *Service) BuildReport(ctx context.Context, id string, expected uint64) (BuildReportView, error) {
	doc, snapshot, err := s.buildDocumentAtRevision(ctx, id, expected)
	if err != nil {
		return BuildReportView{}, err
	}
	reader, ok := s.calculator.(buildReportReader)
	if !ok {
		return BuildReportView{}, failure(InvalidArgument, "подробный отчёт доступен с исходным Lua PoB")
	}
	report, err := reader.ReadBuildReport(ctx, doc, snapshot.BuildName)
	if err != nil {
		return BuildReportView{}, failure(CalculationFailure, err.Error())
	}
	return BuildReportView{Revision: snapshot.Revision, Report: report}, nil
}
