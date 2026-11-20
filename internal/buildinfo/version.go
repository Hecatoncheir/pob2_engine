package buildinfo

// Version задаётся при сборке выпуска через -ldflags.
var Version = "dev"

const ProtocolVersion = "pob2.v1"

func Capabilities() []string {
	return []string{"session.create", "session.snapshot", "build.rename"}
}
