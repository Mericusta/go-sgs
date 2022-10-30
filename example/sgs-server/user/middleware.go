package user

import (
	"fmt"

	"github.com/Mericusta/go-sgs/event"
	serverModel "github.com/Mericusta/go-sgs/example/model/server"
	"github.com/Mericusta/go-sgs/example/sgs-server/server"
	"github.com/Mericusta/go-sgs/middleware"
)

type UserMiddleware struct {
	server *server.Server
}

func NewMiddleware(server *server.Server) *UserMiddleware {
	return &UserMiddleware{server: server}
}

func (m *UserMiddleware) Do(ctx middleware.IContext, e *event.Event) bool {
	if handler, has := userHandlerMgr[e.ID()]; handler != nil && has {
		iUser, has := m.server.UserMgr().Load(ctx.Link().UID())
		if !has {
			fmt.Printf("Error: can not find user by uid %v", ctx.Link().UID())
			return false
		}
		user, ok := iUser.(*serverModel.User)
		if !ok {
			fmt.Printf("Error: server user manager uid %v value type is not *User\n", ctx.Link().UID())
			return false
		}
		handler(NewContext(ctx, user), e.Data())
		return false
	}
	return true
}
