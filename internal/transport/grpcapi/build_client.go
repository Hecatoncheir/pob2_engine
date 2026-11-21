package grpcapi

import (
	"context"
	"fmt"
	"io"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/application"
	"github.com/Hecatoncheir/pob2_engine/internal/domain"
)

func (c *Client) Apply(ctx context.Context, cmd application.ApplyBuild) (domain.Snapshot, error) {
	if (cmd.XML == "") == (cmd.Edit == nil) {
		return domain.Snapshot{}, fmt.Errorf("нужен XML либо одна правка")
	}
	edit, err := wireEdit(cmd.Edit)
	if err != nil {
		return domain.Snapshot{}, err
	}
	req := &pb.ApplyBuildRequest{SessionId: cmd.SessionID, ExpectedRevision: cmd.ExpectedRevision, RequestId: cmd.RequestID}
	if edit != nil {
		req.Input = &pb.ApplyBuildRequest_Edit{Edit: edit}
	} else {
		req.Input = &pb.ApplyBuildRequest_Xml{Xml: cmd.XML}
	}
	answer, err := c.rpc.ApplyBuild(ctx, req)
	if err != nil {
		return domain.Snapshot{}, err
	}
	return fromSnapshot(answer.GetSnapshot())
}

func (c *Client) Calculate(ctx context.Context, cmd application.CalculateBuild) (domain.Snapshot, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stream, err := c.rpc.CalculateBuild(ctx, &pb.CalculateBuildRequest{SessionId: cmd.SessionID, ExpectedRevision: cmd.ExpectedRevision, RequestId: cmd.RequestID})
	if err != nil {
		return domain.Snapshot{}, err
	}
	for {
		answer, err := stream.Recv()
		if err == io.EOF {
			return domain.Snapshot{}, fmt.Errorf("расчёт завершился без итогового снимка")
		}
		if err != nil {
			return domain.Snapshot{}, err
		}
		application.ReportProgress(ctx, answer.GetStage())
		if answer.GetSnapshot() != nil {
			return fromSnapshot(answer.GetSnapshot())
		}
	}
}

func (c *Client) Export(ctx context.Context, id string) (application.ExportedBuild, error) {
	value, err := c.rpc.ExportBuild(ctx, &pb.ExportBuildRequest{SessionId: id})
	if err != nil {
		return application.ExportedBuild{}, err
	}
	return application.ExportedBuild{SessionID: value.GetSessionId(), Revision: value.GetRevision(), XML: value.GetXml(), InputSHA256: value.GetInputSha256()}, nil
}

func (c *Client) Compare(ctx context.Context, cmd application.CompareBuild) (application.BuildComparison, error) {
	if (cmd.XML == "") == (cmd.Edit == nil) {
		return application.BuildComparison{}, fmt.Errorf("нужен XML варианта либо одна правка")
	}
	edit, err := wireEdit(cmd.Edit)
	if err != nil {
		return application.BuildComparison{}, err
	}
	req := &pb.CompareBuildRequest{SessionId: cmd.SessionID, ExpectedRevision: cmd.ExpectedRevision}
	if edit != nil {
		req.Input = &pb.CompareBuildRequest_Edit{Edit: edit}
	} else {
		req.Input = &pb.CompareBuildRequest_Xml{Xml: cmd.XML}
	}
	value, err := c.rpc.CompareBuild(ctx, req)
	if err != nil {
		return application.BuildComparison{}, err
	}
	comparison, err := localComparison(value.GetComparison())
	if err != nil {
		return application.BuildComparison{}, err
	}
	return application.BuildComparison{SessionID: value.GetSessionId(), Revision: value.GetRevision(), Comparison: comparison}, nil
}
