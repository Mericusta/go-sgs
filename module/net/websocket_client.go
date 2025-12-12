package moduleNet

import (
	"net/http"

	"github.com/Mericusta/go-sgs"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type IDialBehavior interface {
	URL() string
	Header() http.Header
	ToSubject() string
	Failed() error
}

// 发起 dial，建立 socket 链接
type ModuleWebsocketClient struct {
	sgs.ModuleBase

	// dialURLs    []string
	dialer *websocket.Dialer
	// dialCounter atomic.Int64
}

var WebsocketClient *ModuleWebsocketClient

func (*ModuleWebsocketClient) New(mos ...sgs.ModuleOption) *ModuleWebsocketClient {
	mwc := &ModuleWebsocketClient{}
	for _, mo := range mos {
		mo(mwc)
	}
	return mwc
}

func (*ModuleWebsocketClient) WithDialer(d *websocket.Dialer) sgs.ModuleOption {
	return func(m sgs.Module) { m.(*ModuleWebsocketClient).dialer = d }
}

// func (*ModuleWebsocketClient) WithDialURLs(urls []string) sgs.ModuleOption {
// 	return func(m sgs.Module) { m.(*ModuleWebsocketClient).dialURLs = urls }
// }

func (mwc *ModuleWebsocketClient) Mounted() {
	if mwc.dialer == nil {
		// 未设置 dialer 不提供服务
		panic("dialer is nil")
	}

	// if len(mwc.dialURLs) == 0 {
	// 	// 未提供 dial 地址不提供服务
	// 	panic("dialURLs is empty")
	// }
}

func (mwc *ModuleWebsocketClient) HandleEvent(event *sgs.ModuleEvent) {
	switch data := event.Data().(type) {
	case IDialBehavior:
		go mwc.dial(data)
	default:
		mwc.Logger().Error(sgs.ErrorMsgHandleEventNonImplement, zap.Any("data", data))
	}
}

func (mwc *ModuleWebsocketClient) dial(dialBehavior IDialBehavior) {
	// counter := mwc.dialCounter.Add(1)
	// dialURL := mwc.dialURLs[counter%int64(len(mwc.dialURLs))]
	wsc, _, err := websocket.DefaultDialer.DialContext(mwc.Ctx(), dialBehavior.URL(), dialBehavior.Header())
	if err != nil {
		mwc.Logger().Error("dial, DialContext occurs error", zap.Error(err), zap.Any("dialURL", dialBehavior.URL()), zap.Any("toSubject", dialBehavior.ToSubject()), zap.Any("header", dialBehavior.Header()))
		err = mwc.SendEvent(sgs.NewModuleEvent(dialBehavior.ToSubject(), dialBehavior.Failed()))
		if err != nil {
			mwc.Logger().Error("dial, DialContext failed, SendEvent occurs error", zap.Error(err), zap.Any("toSubject", dialBehavior.ToSubject()), zap.Any("failed", dialBehavior.Failed()))
		}
		return
	}

	err = mwc.SendEvent(sgs.NewModuleEvent(dialBehavior.ToSubject(), wsc))
	if err != nil {
		mwc.Logger().Error("dial, SendEvent occurs error", zap.Error(err), zap.Any("toSubject", dialBehavior.ToSubject()))
		return
	}
}
