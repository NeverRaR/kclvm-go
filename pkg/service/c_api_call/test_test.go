//go:build cgo && kclvm_service_capi
// +build cgo,kclvm_service_capi

package capicall

import (
	"testing"
)

func TestExecSingleFile(t *testing.T) {
	files := getFiles(ERROR_DATA_PATH, ".k", true)
	for _, file := range files {
		result := execSingle(ERROR_DATA_PATH, file)
		println(result)
	}
}

func execSingle(workDir string, fileName string) string {
	return ExecSingleFile(workDir, fileName)
}
