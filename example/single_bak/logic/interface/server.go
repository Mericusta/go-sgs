package inObj

import (
	"github.com/Mericusta/go-sgs"
)

type IServer interface {
	// basic
	Identify() string
	Logger() *sgs.Logger
}
