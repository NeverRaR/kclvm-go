//go:build cgo
// +build cgo

package capicall

import (
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"kusionstack.io/kclvm-go/pkg/spec/gpyrpc"
)

func TestPing(t *testing.T) {
	client := PROTOCAPI_NewKclvmServiceClient()
	out, err := client.Ping(nil)
	assert.Nil(t, err)
	out, err = client.Ping(&gpyrpc.Ping_Args{Value: "hello"})
	assert.Nil(t, err)
	assert.Equal(t, "hello", out.Value)
}

func TestExecProgram(t *testing.T) {
	for j := 0; j < 10; j++ {
		wg :=
			sync.WaitGroup{}
		n := 10
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func() {

				workdir, _ := filepath.Abs(CORRECT_DATA_PATH)
				args := &gpyrpc.ExecProgram_Args{
					WorkDir:       workdir,
					KFilenameList: []string{"complex.k"},
					Args: []*gpyrpc.CmdArgSpec{
						{Name: "__kcl_test_run", Value: "___test_schema_@@@__"},
						{Name: "__kcl_test_debug", Value: "true"},
					},
					Overrides:         []*gpyrpc.CmdOverrideSpec{},
					DisableYamlResult: false,
				}
				client := PROTOCAPI_NewKclvmServiceClient()
				client.ExecProgram(args)
				println("ok")
				//println(out.JsonResult)
				wg.Done()
			}()
		}
		wg.Wait()
	}

}
