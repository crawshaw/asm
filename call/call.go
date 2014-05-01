package call

import "C"
import "unsafe"

func Call(fn, arg unsafe.Pointer, n uint32)
