// Copyright 2022 The KCL Authors. All rights reserved.

//go:build cgo
// +build cgo

package kcl_plugin

/*

typedef struct PyObject PyObject;

*/
import "C"

import (
	_ "embed"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"kusionstack.io/kclvm-go/pkg/3rdparty/dlopen"
	"kusionstack.io/kclvm-go/pkg/kclvm_runtime"
)

const KCLVM_PY_PLUGIN_MODULE_NAME = "__kclvm_plugin_py_in_go__"

//go:embed kcl_plugin_py/kclvm_plugin_py.py
var kclvmPyPluginModuleSrc string

var kclvmPyPluginModule *C.PyObject

func InitKclvmPyPlugin() {
	pyLib = loadKclvmPyLib()
	PySetPythonHome(pythonHome)
	PyInitialize()
	addPyImportPath(getKclvmPyPackagePath())
	addPyImportPath(getKclvmPyPackagePath() + "/site-packages")
	pluginModFileName := KCLVM_PY_PLUGIN_MODULE_NAME + ".py"
	pluginModPath, err := filepath.Abs(pluginModFileName)
	if err != nil {
		panic(fmt.Errorf("fail to get abs path of file : %s", pluginModFileName))
	}
	_, err = os.Stat(pluginModPath)
	var kclvmPyPluginFile *os.File
	if os.IsNotExist(err) {
		kclvmPyPluginFile, err = os.Create(pluginModPath)
		if err != nil {
			panic(fmt.Errorf("fail to create file : %s", pluginModPath))
		}
	} else {
		kclvmPyPluginFile, err = os.OpenFile(pluginModPath, os.O_RDWR, 0666)
		if err != nil {
			panic(fmt.Errorf("fail to open file : %s", pluginModPath))
		}
	}
	_, err = io.WriteString(kclvmPyPluginFile, kclvmPyPluginModuleSrc)
	if err != nil {
		panic(fmt.Errorf("fail to write file : %s", pluginModPath))
	}
	kclvmPyPluginModule = ImportPyModule(pluginModPath[:strings.LastIndex(pluginModPath, "/")], KCLVM_PY_PLUGIN_MODULE_NAME)
}

func loadKclvmPyLib() *dlopen.LibHandle {

	pyLibPath := findPy3LibPath()

	pyLibDir, err := ioutil.ReadDir(pyLibPath)
	if err != nil {
		return nil
	}

	sysType := runtime.GOOS
	libSuffix := ".so"

	if sysType == "darwin" {
		libSuffix = ".dylib"
	} else if sysType == "windows" {
		libSuffix = ".dll"
	}
	pyLib := ""
	for _, fi := range pyLibDir {
		if !fi.IsDir() {
			if strings.HasPrefix(fi.Name(), "libpython") && strings.HasSuffix(fi.Name(), libSuffix) {
				pyLib = pyLibPath + "/" + fi.Name()
				break
			}
		}
	}

	h, err := dlopen.GetHandle([]string{pyLib})

	if err != nil {
		return nil
	}

	runtime.SetFinalizer(h, func(x *dlopen.LibHandle) {
		x.Close()
	})

	return h
}

func findPy3LibPath() string {
	pythonHome = os.Getenv("KCLVM_PYTHON3_HOME")
	if len(pythonHome) == 0 {
		panic("can not get KCLVM_PYTHON3_HOME! please set KCLVM_PYTHON3_HOME")
	}
	return pythonHome + "/lib"
}

func getKclvmPyPackagePath() string {
	kclvmLibPath := kclvm_runtime.MustGetKclvmRoot() + "/lib"
	kclvmLibDir, err := ioutil.ReadDir(kclvmLibPath)
	if err != nil {
		panic(fmt.Errorf("can not find kclvmLibDir:%s", kclvmLibPath))
	}
	kclvmPyPackagePath := ""
	for _, fi := range kclvmLibDir {
		if fi.IsDir() {
			if strings.HasPrefix(fi.Name(), "python") {
				kclvmPyPackagePath = strings.Join([]string{kclvmLibPath, fi.Name()}, "/")
			}

		}
	}
	return kclvmPyPackagePath
}

func addPyImportPath(importPath string) {
	sys := PyImportImportModule("sys")
	defer PyDecRef(sys)
	path := PyObjectGetAttrString(sys, "path")
	defer PyDecRef(path)
	pImportPath := PyUnicodeFromString(importPath)
	defer PyDecRef(pImportPath)
	PyListAppend(path, pImportPath)
}

func ImportPyModule(dir, name string) *C.PyObject {
	sys := PyImportImportModule("sys")
	defer PyDecRef(sys)
	path := PyObjectGetAttrString(sys, "path")
	defer PyDecRef(path)
	PyListInsert(path, 0, PyUnicodeFromString(dir))
	return PyImportImportModule(name)
}

// DecRef PyObject inside func CallPyFunc instead of outside to prevent multiple DecRef
// PyTuple_SetItem “steals” a reference to o and discards a reference to an item already in the tuple at the affected position,
// so we just need DecRef the tuple
// PyDict_SetItem does not "steal" a reference to val, so we need DecRef both every val and dict itself
func CallPyFunc(module *C.PyObject, funcName string, kwargs map[string]*C.PyObject, args ...*C.PyObject) *C.PyObject {
	funcObj := PyObjectGetAttrString(module, funcName)
	argsLen := len(args)
	pyArgs := PyTupleNew(argsLen)

	defer PyDecRef(pyArgs)

	for i, obj := range args {
		PyTupleSetItem(pyArgs, i, obj)
	}

	pyKwargs := PyDictNew()

	defer PyDecRef(pyKwargs)

	for k, v := range kwargs {
		kObj := PyUnicodeFromString(k)
		defer PyDecRef(kObj)
		defer PyDecRef(v)
		PyDictSetItem(pyKwargs, kObj, v)
	}

	return PyObjectCall(funcObj, pyArgs, pyKwargs)
}

func PyObjToString(obj *C.PyObject) (string, error) {
	s := PyObjectRepr(obj)
	if s == nil {
		PyErrClear()
		return "", fmt.Errorf("failed to call Repr object method")
	}
	defer PyDecRef(s)

	return PyUnicodeAsUTF8(s), nil
}
