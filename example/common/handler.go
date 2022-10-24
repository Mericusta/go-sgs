package common

import (
	"context"
	"fmt"
	"time"

	"github.com/Mericusta/go-sgs/protocol"
)

// TODO: 不要 tickerFunc，用 event 或者什么代替，数据驱动
func HandleLogic[LINKTYPE Client | User](ctx context.Context, link *LINKTYPE, tickerFunc func(*LINKTYPE), callbackMap map[protocol.ProtocolID]func(*LINKTYPE, protocol.Protocol)) {
	ticker := time.NewTicker(time.Second)
LOOP:
	for {
		select {
		case <-ticker.C: // 主动发送 TODO: event chan
		priority:
			for {
				select {
				case <-ctx.Done(): // 主动结束，必须保证本端先断开 tcp 套接字，再 cancel
					ticker.Stop()
					fmt.Printf("Note: link %v stop ticker\n", link.UID())
					goto LOOP
				default:
					break priority
				}
			}
			// 发送逻辑
			fmt.Printf("Note: link %v handle send logic\n", link.UID())
			if tickerFunc != nil {
				tickerFunc(link)
			}
		case e, ok := <-link.Recv(): // 被动接收，对端断开 tcp 套接字
			if !ok { // 被动结束
				fmt.Printf("Note: link %v receive channel closed\n", link.UID())
				ticker.Stop()
				break LOOP
			}
			// 接收逻辑
			fmt.Printf("Note: link %v handle recv logic\n", link.UID())
			callback := callbackMap[e.ID()]
			if callback == nil {
				fmt.Printf("Error: event ID %v callback is nil\n", e.ID())
				continue
			}
			callback(link, e.Data())
		}
	}
}
