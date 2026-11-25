//go:build (linux || darwin) && cgo

package pob

/*
#include <stdint.h>
extern uint32_t pob2SolveTreePaths(uint32_t *, uint32_t, uint32_t *, uint32_t);
*/
import "C"
import "unsafe"

func treePathCallback() uintptr {
	return uintptr(C.pob2SolveTreePaths)
}

//export pob2SolveTreePaths
func pob2SolveTreePaths(input *C.uint32_t, inputSize C.uint32_t, output *C.uint32_t, outputSize C.uint32_t) C.uint32_t {
	return C.uint32_t(solvePathWords(
		(*uint32)(unsafe.Pointer(input)), uint32(inputSize),
		(*uint32)(unsafe.Pointer(output)), uint32(outputSize),
	))
}
