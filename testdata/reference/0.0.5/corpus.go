// Package multigroup хранит исходные XML для проверки установки PoB.
package multigroup

import "embed"

// Builds содержит по одной сборке каждого типа из приёмки 0.0.5.
//
//go:embed caster.xml attacker.xml minion.xml
var Builds embed.FS
