package maintenance

import (
	"context"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/Hecatoncheir/pob2_engine/internal/buildinfo"
	"github.com/Hecatoncheir/pob2_engine/internal/gamedata"
	"github.com/Hecatoncheir/pob2_engine/internal/pob"
)

const DoctorCapability = "maintenance.doctor"

type Report struct {
	Status          string        `json:"status"`
	Version         string        `json:"version"`
	ProtocolVersion string        `json:"protocol_version"`
	OS              string        `json:"os"`
	Arch            string        `json:"arch"`
	GoVersion       string        `json:"go_version"`
	Executable      string        `json:"executable"`
	Commit          string        `json:"commit,omitempty"`
	Modified        bool          `json:"modified"`
	DataID          string        `json:"data_id"`
	DataSHA256      string        `json:"data_sha256"`
	TreeVersion     string        `json:"tree_version"`
	PoBCommit       string        `json:"pob_commit"`
	RestartRequired bool          `json:"restart_required"`
	Issues          []string      `json:"issues,omitempty"`
	Server          *ServerReport `json:"server,omitempty"`
	EngineUpdate    *pob.State    `json:"engine_update,omitempty"`
	PoB             *pob.State    `json:"pob,omitempty"`
}

type ServerReport struct {
	State        string `json:"state"`
	Mode         string `json:"mode"`
	StorageMode  string `json:"storage_mode"`
	StoragePath  string `json:"storage_path,omitempty"`
	StorageBytes int64  `json:"storage_bytes,omitempty"`
	Clients      uint32 `json:"clients"`
	ActiveTasks  uint32 `json:"active_tasks"`
}

func (s *Service) Doctor(ctx context.Context) (Report, error) {
	if err := ctx.Err(); err != nil {
		return Report{}, err
	}
	report := Report{
		Status: "ok", Version: s.version, ProtocolVersion: buildinfo.ProtocolVersion,
		OS: s.goos, Arch: s.goarch, GoVersion: runtime.Version(), Executable: s.path,
	}
	report.Commit, report.Modified = revision()
	s.inspectExecutable(&report)
	report.inspectData()
	s.inspectUpdates(ctx, &report)
	return report, nil
}

func (s *Service) inspectExecutable(report *Report) {
	if s.pathErr != nil {
		report.AddIssue("исполняемый файл: " + s.pathErr.Error())
		return
	}
	info, err := os.Stat(s.path)
	if err != nil {
		report.AddIssue("исполняемый файл: " + err.Error())
		return
	}
	report.RestartRequired = !os.SameFile(s.image, info)
}

func (r *Report) inspectData() {
	data, err := gamedata.Embedded()
	if err != nil {
		r.AddIssue("встроенный каталог: " + err.Error())
		return
	}
	r.DataID, r.DataSHA256 = data.ID(), data.SHA256()
	r.TreeVersion, r.PoBCommit = data.TreeVersion(), data.PoBCommit()
}

func (r *Report) AddIssue(issue string) {
	r.Status = "degraded"
	r.Issues = append(r.Issues, issue)
}

func (r *Report) InspectStorage() {
	if r.Server.StoragePath == "" {
		return
	}
	info, err := os.Lstat(r.Server.StoragePath)
	if err != nil {
		r.AddIssue("хранилище: " + err.Error())
		return
	}
	if !info.Mode().IsRegular() || info.Size() == 0 {
		r.AddIssue("хранилище: некорректный файл базы")
		return
	}
	r.Server.StorageBytes = info.Size()
}

func revision() (commit string, modified bool) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", false
	}
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			commit = setting.Value
		case "vcs.modified":
			modified = setting.Value == "true"
		}
	}
	return commit, modified
}
