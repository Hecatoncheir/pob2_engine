//go:build pob_bytecode

package pob

import (
	"embed"
	"io/fs"
)

//go:embed all:bytecode/embedded
var compiledBridges embed.FS

func init() {
	bundledBytecode, _ = fs.Sub(compiledBridges, "bytecode/embedded")
}
