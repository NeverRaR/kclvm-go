//go:build cgo && kclvm_service_capi
// +build cgo,kclvm_service_capi

package capicall

// #include "kclvm_service_call.h"
// #include <stdlib.h>
import "C"

func ExecSingleFile(workDir string, fileName string) string {
	cWork := C.CString(workDir)
	cFile := C.CString(fileName)
	result := C.GoString(C.kclvm_service_single_exec(cWork, cFile))
	return result
}
