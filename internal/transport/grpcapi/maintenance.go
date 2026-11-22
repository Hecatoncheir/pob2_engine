package grpcapi

import (
	"context"
	"errors"

	pb "github.com/Hecatoncheir/pob2_engine/gen/go/pob2/v1"
	"github.com/Hecatoncheir/pob2_engine/internal/maintenance"
	"github.com/Hecatoncheir/pob2_engine/internal/pob"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, maintenance.UpdateTimeout)
	defer cancel()
	result, err := s.maintenance.Update(ctx, req.GetApply(), req.GetTarget())
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil, status.FromContextError(err).Err()
	}
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return &pb.UpdateResponse{Status: result.Status, CurrentVersion: result.CurrentVersion, LatestVersion: result.LatestVersion,
		ReleaseUrl: result.ReleaseURL, RestartRequired: result.RestartRequired,
		Engine: encodeComponent(result.Engine), Pob: encodeComponent(result.PoB)}, nil
}

func (s *Server) Doctor(ctx context.Context, _ *pb.DoctorRequest) (*pb.DoctorResponse, error) {
	report, err := s.maintenance.Doctor(ctx)
	if err != nil {
		return nil, rpcError(err)
	}
	stats := s.life.Stats()
	mode, path := s.storageInfo()
	report.Server = &maintenance.ServerReport{State: stats.State, Mode: s.life.Config().Mode,
		StorageMode: mode, StoragePath: path, Clients: stats.Clients, ActiveTasks: stats.Tasks}
	report.InspectStorage()
	return encodeDoctor(report), nil
}

func encodeDoctor(r maintenance.Report) *pb.DoctorResponse {
	s := r.Server
	return &pb.DoctorResponse{
		Status: r.Status, Version: r.Version, ProtocolVersion: r.ProtocolVersion, Os: r.OS, Arch: r.Arch,
		GoVersion: r.GoVersion, Executable: r.Executable, Commit: r.Commit, Modified: r.Modified,
		DataId: r.DataID, DataSha256: r.DataSHA256, TreeVersion: r.TreeVersion, PobCommit: r.PoBCommit,
		RestartRequired: r.RestartRequired, Issues: r.Issues,
		EngineUpdate: encodeComponent(r.EngineUpdate), Pob: encodeComponent(r.PoB),
		Server: &pb.DoctorServer{State: s.State, Mode: s.Mode, StorageMode: s.StorageMode, StoragePath: s.StoragePath,
			StorageBytes: s.StorageBytes, Clients: s.Clients, ActiveTasks: s.ActiveTasks},
	}
}

func (c *Client) Update(ctx context.Context, apply bool, targets ...string) (maintenance.UpdateResult, error) {
	target := ""
	if len(targets) != 0 {
		target = targets[0]
	}
	result, err := c.rpc.Update(ctx, &pb.UpdateRequest{Apply: apply, Target: target})
	if err != nil {
		return maintenance.UpdateResult{}, err
	}
	return maintenance.UpdateResult{Status: result.GetStatus(), CurrentVersion: result.GetCurrentVersion(),
		LatestVersion: result.GetLatestVersion(), ReleaseURL: result.GetReleaseUrl(), RestartRequired: result.GetRestartRequired(),
		Engine: decodeComponent(result.GetEngine()), PoB: decodeComponent(result.GetPob())}, nil
}

func (c *Client) Doctor(ctx context.Context) (maintenance.Report, error) {
	response, err := c.rpc.Doctor(ctx, &pb.DoctorRequest{})
	if err != nil {
		return maintenance.Report{}, err
	}
	report := maintenance.Report{
		Status: response.GetStatus(), Version: response.GetVersion(), ProtocolVersion: response.GetProtocolVersion(),
		OS: response.GetOs(), Arch: response.GetArch(), GoVersion: response.GetGoVersion(), Executable: response.GetExecutable(),
		Commit: response.GetCommit(), Modified: response.GetModified(), DataID: response.GetDataId(), DataSHA256: response.GetDataSha256(),
		TreeVersion: response.GetTreeVersion(), PoBCommit: response.GetPobCommit(), RestartRequired: response.GetRestartRequired(), Issues: response.GetIssues(),
		EngineUpdate: decodeComponent(response.GetEngineUpdate()), PoB: decodeComponent(response.GetPob()),
	}
	if s := response.GetServer(); s != nil {
		report.Server = &maintenance.ServerReport{State: s.GetState(), Mode: s.GetMode(), StorageMode: s.GetStorageMode(),
			StoragePath: s.GetStoragePath(), StorageBytes: s.GetStorageBytes(), Clients: s.GetClients(), ActiveTasks: s.GetActiveTasks()}
	}
	return report, nil
}

func encodeComponent(state *pob.State) *pb.ComponentVersion {
	if state == nil {
		return nil
	}
	return &pb.ComponentVersion{Status: state.Status, InstalledVersion: state.InstalledVersion,
		LatestVersion: state.LatestVersion, ReleaseUrl: state.ReleaseURL,
		RestartRequired: state.RestartRequired, Message: state.Message}
}

func decodeComponent(state *pb.ComponentVersion) *pob.State {
	if state == nil {
		return nil
	}
	return &pob.State{Status: state.GetStatus(), InstalledVersion: state.GetInstalledVersion(),
		LatestVersion: state.GetLatestVersion(), ReleaseURL: state.GetReleaseUrl(),
		RestartRequired: state.GetRestartRequired(), Message: state.GetMessage()}
}
