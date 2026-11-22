//go:build linux || darwin

package maintenance

import "github.com/Hecatoncheir/pob2_engine/internal/privatefs"

func replaceExecutable(source, target string) error { return privatefs.Replace(source, target) }
