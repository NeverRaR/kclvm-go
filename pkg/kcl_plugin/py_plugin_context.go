package kcl_plugin

import (
	"sync"

	"github.com/timandy/routine"
)

var ctxThreadLocal = routine.NewThreadLocal()

var ctxMutex sync.Mutex

type PyPluginContext struct {
	PathList []string
	WorkDir  string
	Target   string
}

func NewPyPluginContext() *PyPluginContext {
	ctx := new(PyPluginContext)
	return ctx
}
