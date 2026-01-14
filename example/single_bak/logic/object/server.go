package obj

import (
	"github.com/Mericusta/go-sgs"
)

type ModuleSingleServer struct {
	sgs.ModuleBase
}

var SingleServer *ModuleSingleServer

func (m *ModuleSingleServer) New(mos ...sgs.ModuleOption) *ModuleSingleServer {
	mgs := &ModuleSingleServer{}
	for _, mo := range mos {
		mo(mgs)
	}
	return mgs
}
