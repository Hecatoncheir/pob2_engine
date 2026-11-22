package grpcapi

import (
	"context"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/calculation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) catalogCommands() (application.CatalogCommands, error) {
	commands, ok := s.commands.(application.CatalogCommands)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "сервис не поддерживает каталог и предпросмотр")
	}
	return commands, nil
}

func (s *Server) GetCatalog(ctx context.Context, req *pb.GetCatalogRequest) (*pb.GetCatalogResponse, error) {
	commands, err := s.catalogCommands()
	if err != nil {
		return nil, err
	}
	page, err := commands.Catalog(ctx, application.CatalogQuery{Kind: req.GetKind(), Query: req.GetQuery(), Offset: int(req.GetOffset()), Limit: int(req.GetLimit())})
	if err != nil {
		return nil, rpcError(err)
	}
	result := &pb.GetCatalogResponse{DataId: page.DataID, DataSha256: page.DataSHA256, PobCommit: page.PoBCommit, TreeVersion: page.TreeVersion,
		Kind: page.Kind, Query: page.Query, Offset: int32(page.Offset), Limit: int32(page.Limit), Total: int32(page.Total)}
	for _, entry := range page.Entries {
		result.Entries = append(result.Entries, wireCatalogEntry(entry))
	}
	return result, nil
}

func (c *Client) Catalog(ctx context.Context, query application.CatalogQuery) (application.CatalogPage, error) {
	// Проверка до сужения int: большие значения не должны превращаться в допустимые.
	if query.Offset < 0 || query.Offset > 10000 || query.Limit < 0 || query.Limit > 100 {
		return application.CatalogPage{}, status.Error(codes.InvalidArgument, "неверные границы страницы каталога")
	}
	value, err := c.rpc.GetCatalog(ctx, &pb.GetCatalogRequest{Kind: query.Kind, Query: query.Query, Offset: int32(query.Offset), Limit: int32(query.Limit)})
	if err != nil {
		return application.CatalogPage{}, err
	}
	page := application.CatalogPage{DataID: value.GetDataId(), DataSHA256: value.GetDataSha256(), PoBCommit: value.GetPobCommit(), TreeVersion: value.GetTreeVersion(),
		Kind: value.GetKind(), Query: value.GetQuery(), Offset: int(value.GetOffset()), Limit: int(value.GetLimit()), Total: int(value.GetTotal()), Entries: []application.CatalogEntry{}}
	for _, entry := range value.GetEntries() {
		page.Entries = append(page.Entries, localCatalogEntry(entry))
	}
	return page, nil
}

func (s *Server) PreviewSupport(ctx context.Context, req *pb.PreviewSupportRequest) (*pb.PreviewSupportResponse, error) {
	commands, err := s.catalogCommands()
	if err != nil {
		return nil, err
	}
	value, err := commands.PreviewSupport(ctx, application.PreviewSupport{SessionID: req.GetSessionId(), ExpectedRevision: req.GetExpectedRevision(), SupportID: req.GetSupportId(), Level: int(req.GetLevel()), Quality: int(req.GetQuality())})
	if err != nil {
		return nil, rpcError(err)
	}
	return &pb.PreviewSupportResponse{SessionId: value.SessionID, Revision: value.Revision, GroupId: int32(value.GroupID), Support: wireCatalogEntry(value.Support),
		Compatible: value.Compatible, Status: value.Status, Comparison: wireComparison(value.Comparison)}, nil
}

func (c *Client) PreviewSupport(ctx context.Context, cmd application.PreviewSupport) (application.SupportPreview, error) {
	value, err := c.rpc.PreviewSupport(ctx, &pb.PreviewSupportRequest{SessionId: cmd.SessionID, ExpectedRevision: cmd.ExpectedRevision, SupportId: cmd.SupportID, Level: int32(cmd.Level), Quality: int32(cmd.Quality)})
	if err != nil {
		return application.SupportPreview{}, err
	}
	comparison, err := localComparison(value.GetComparison())
	if err != nil {
		return application.SupportPreview{}, err
	}
	return application.SupportPreview{SessionID: value.GetSessionId(), Revision: value.GetRevision(), GroupID: int(value.GetGroupId()), Support: localCatalogEntry(value.GetSupport()),
		Compatible: value.GetCompatible(), Status: value.GetStatus(), Comparison: comparison}, nil
}

func wireCatalogEntry(entry application.CatalogEntry) *pb.CatalogEntry {
	return &pb.CatalogEntry{Kind: entry.Kind, Id: entry.ID, Name: entry.Name, Aliases: entry.Aliases, Levels: wireInts(entry.Levels), Qualities: wireInts(entry.Qualities),
		Modes: entry.Modes, Support: entry.Support, Family: entry.Family, Slots: entry.Slots, ClassIds: wireInts(entry.ClassIDs), Coverage: entry.Coverage}
}

func localCatalogEntry(entry *pb.CatalogEntry) application.CatalogEntry {
	return application.CatalogEntry{Kind: entry.GetKind(), ID: entry.GetId(), Name: entry.GetName(), Aliases: append([]string{}, entry.GetAliases()...), Levels: localInts(entry.GetLevels()), Qualities: localInts(entry.GetQualities()),
		Modes: append([]string{}, entry.GetModes()...), Support: entry.GetSupport(), Family: entry.GetFamily(), Slots: append([]string{}, entry.GetSlots()...), ClassIDs: localInts(entry.GetClassIds()), Coverage: append([]string{}, entry.GetCoverage()...)}
}

func wireInts(values []int) []int32 {
	result := make([]int32, len(values))
	for i, value := range values {
		result[i] = int32(value)
	}
	return result
}

func localInts(values []int32) []int {
	result := make([]int, len(values))
	for i, value := range values {
		result[i] = int(value)
	}
	return result
}

func wireComparison(value calculation.Comparison) *pb.BuildComparison {
	result := &pb.BuildComparison{Before: wireResult(value.Before), After: wireResult(value.After)}
	for _, delta := range value.Deltas {
		result.Deltas = append(result.Deltas, &pb.StatDelta{Stat: delta.Stat, Before: delta.Before, After: delta.After, Difference: delta.Difference})
	}
	return result
}

func localComparison(value *pb.BuildComparison) (calculation.Comparison, error) {
	before, err := localResult(value.GetBefore())
	if err != nil {
		return calculation.Comparison{}, err
	}
	after, err := localResult(value.GetAfter())
	if err != nil {
		return calculation.Comparison{}, err
	}
	result := calculation.Comparison{Before: before, After: after, Deltas: []calculation.Delta{}}
	for _, delta := range value.GetDeltas() {
		result.Deltas = append(result.Deltas, calculation.Delta{Stat: delta.GetStat(), Before: delta.GetBefore(), After: delta.GetAfter(), Difference: delta.GetDifference()})
	}
	return result, nil
}
