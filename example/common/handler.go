package common

import (
	"context"
	"fmt"
	"time"

	"github.com/Mericusta/go-sgs/protocol"
)

func HandleLogic[LINKTYPE Client | User](ctx context.Context, link *LINKTYPE, callbackMap map[protocol.ProtocolID]func(*LINKTYPE, protocol.Protocol)) {
	ticker := time.NewTicker(time.Second)
LOOP:
	for {
		select {
		case <-ticker.C: // 主动发送
		priority:
			for {
				select {
				case <-ctx.Done(): // 主动结束
					ticker.Stop()
					fmt.Printf("Note: link %v stop ticker\n", link.UID())
					goto LOOP
				default:
					break priority
				}
			}
			// 发送逻辑
		case e, ok := <-link.Recv(): // 被动接收
			if e == nil || !ok { // 被动结束
				fmt.Printf("Note: link %v receive channel closed\n", link.UID())
				ticker.Stop()
				break LOOP
			}
		}
	}
}
