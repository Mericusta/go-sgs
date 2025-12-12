package sgs

import (
	"context"
)

type moduleController struct {
	module   Module
	canceler context.CancelFunc
}

func (mc *moduleController) Module() Module {
	return mc.module
}

func (mc *moduleController) Cancel() {
	mc.canceler()
}
