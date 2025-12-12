package moduleMsgQueue

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/Mericusta/go-sgs"
	"github.com/Mericusta/go-sgs/config"
	"go.uber.org/zap"
)

type ModuleKafkaService struct {
	sgs.ModuleBase

	config *config.MsgQueueConfig

	isSync bool

	// 异步生产者
	asyncProducer sarama.AsyncProducer

	// 同步生产者
	syncProducer sarama.SyncProducer

	// kafka 配置
	producerConfig *sarama.Config

	// 同步超时时间
	syncOvertime time.Duration

	// 缓存队列
	queue      []*sarama.ProducerMessage
	bucketSize int

	// 上报 chan
	reportChan chan []*sarama.ProducerMessage

	// 上报统计
	reportTS    int64
	reportCount int64
}

var KafkaService *ModuleKafkaService

func (*ModuleKafkaService) New(mos ...sgs.ModuleOption) *ModuleKafkaService {
	mks := &ModuleKafkaService{}
	for _, mo := range mos {
		mo(mks)
	}
	return mks
}

func (*ModuleKafkaService) WithKafkaConfig(c *config.MsgQueueConfig) sgs.ModuleOption {
	return func(m sgs.Module) { m.(*ModuleKafkaService).config = c }
}

func (*ModuleKafkaService) WithAsyncProducer(overtime time.Duration) sgs.ModuleOption {
	return func(m sgs.Module) {
		m.(*ModuleKafkaService).isSync = false
		m.(*ModuleKafkaService).syncOvertime = overtime
	}
}

func (*ModuleKafkaService) WithSyncProducer(overtime time.Duration) sgs.ModuleOption {
	return func(m sgs.Module) {
		m.(*ModuleKafkaService).isSync = true
		m.(*ModuleKafkaService).syncOvertime = overtime
	}
}

func (*ModuleKafkaService) WithBucketSize(size int) sgs.ModuleOption {
	return func(m sgs.Module) { m.(*ModuleKafkaService).bucketSize = size }
}

func (m *ModuleKafkaService) Mounted() {
	m.Logger().Debug("OBSERVE: Mounted, dial kafka then subscribe subject", zap.Any("config", m.config))

	if m.config == nil {
		panic(fmt.Errorf("kafka config is nil"))
	}
	if len(m.config.URL) == 0 {
		panic(fmt.Errorf("kafka config url is empty"))
	}
	urls := strings.Split(m.config.URL, ",")
	if len(urls) == 0 {
		panic(fmt.Errorf("kafka config url is empty"))
	}

	config := sarama.NewConfig()
	// 设置缓存 buffer 尺寸
	config.ChannelBufferSize = 65535
	// 使用随机分区
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	// WaitForAll 等待确保消息写入所有副本，同步时增加延迟，异步时降低 kafka 吞吐量
	config.Producer.RequiredAcks = sarama.WaitForAll
	// 设置重试次数
	config.Producer.Retry.Max = 5

	// 创建链接实例
	if m.isSync {
		// 创建同步 producer
		producer, err := sarama.NewSyncProducer(urls, config)
		if err != nil {
			panic(fmt.Errorf("connect to kafka %v occurs error: %v", m.config.URL, err))
		}
		m.syncProducer = producer
		m.producerConfig = config
	} else {
		// 设置每 64kb 数据立刻同步
		config.Producer.Flush.Bytes = 65536
		// 设置每 128 条数据立刻同步
		config.Producer.Flush.Messages = 128
		// 设置每 100ms 立刻同步
		config.Producer.Flush.Frequency = 100 * time.Millisecond
		// 设置 不需要处理 successes chan
		config.Producer.Return.Successes = true
		// 设置 需要处理 errors chan
		config.Producer.Return.Errors = true
		// 创建 异步 producer
		producer, err := sarama.NewAsyncProducer(urls, config)
		if err != nil {
			panic(fmt.Errorf("connect to kafka %v occurs error: %v", m.config.URL, err))
		}
		m.asyncProducer = producer
		m.producerConfig = config
	}

	// 缓存队列
	m.queue = make([]*sarama.ProducerMessage, 0, m.bucketSize)
	m.reportChan = make(chan []*sarama.ProducerMessage, 8)

	m.Logger().Debug("OBSERVE: Mounted, kafka urls", zap.Any("urls", urls), zap.Any("sync", m.isSync))
}

func (m *ModuleKafkaService) HandleEvent(event *sgs.ModuleEvent) {
	switch data := event.Data().(type) {
	case IMessagePatch:
		now := time.Now().Unix()
		for {
			iMsg, has := data.Pop()
			if !has {
				break
			}
			if iMsg == nil {
				continue
			}
			msg, ok := iMsg.(*sarama.ProducerMessage)
			if msg == nil || !ok {
				continue
			}
			m.handleProducerMsg(event.From(), msg)
			m.reportCount++
			if m.reportTS == 0 || (now-m.reportTS) > 10 {
				m.Logger().Info("ModuleKafkaService.HandleEvent", zap.Any("speed pre-second", m.reportCount/10))
				m.reportCount, m.reportTS = 0, now
			}
		}
	case *sarama.ProducerMessage:
		m.handleProducerMsg(event.From(), data)
	default:
		m.Logger().Error(sgs.ErrorMsgHandleEventNonImplement)
	}
}

func (m *ModuleKafkaService) Run() {
	if m.producerConfig.Producer.Return.Successes {
		go m.handleAsyncProducerReturnSuccesses()
	}

	if m.producerConfig.Producer.Return.Errors {
		go m.handleAsyncProducerReturnErrors()
	}

	if m.isSync {
		for {
			select {
			case reportQueue := <-m.reportChan:
				beginAt := time.Now()
				failedCount := 0
				failedErrorMap := make(map[string]int)
				for _, _msg := range reportQueue {
					if _, _, err := m.syncProducer.SendMessage(_msg); err != nil {
						failedCount++
						failedErrorMap[err.Error()]++
					}
				}
				if len(failedErrorMap) > 0 {
					consumeDuration := time.Since(beginAt)
					m.Logger().Info("Run, sync producer send message done", zap.Any("count", len(reportQueue)), zap.Any("consume ms", consumeDuration.Milliseconds()), zap.Any("failedCount", failedCount), zap.Any("failedErrorMap", failedErrorMap))
				} else {
					if rand.Intn(10)%10 == 5 {
						consumeDuration := time.Since(beginAt)
						m.Logger().Info("Run, sync producer send message done", zap.Any("count", len(reportQueue)), zap.Any("consume ms", consumeDuration.Milliseconds()), zap.Any("failedCount", failedCount), zap.Any("failedErrorMap", failedErrorMap))
					}
				}
			case <-m.Ctx().Done():
				select {
				case reportQueue, ok := <-m.reportChan:
					if !ok {
						return
					}
					failedCount := 0
					failedErrorMap := make(map[string]int)
					for _, _msg := range reportQueue {
						if _, _, err := m.syncProducer.SendMessage(_msg); err != nil {
							failedCount++
							failedErrorMap[err.Error()]++
						}
					}
					m.Logger().Info("Run before return, sync producer send message done", zap.Any("count", len(reportQueue)), zap.Any("failedCount", failedCount), zap.Any("failedErrorMap", failedErrorMap))
				default:
					return
				}
				return
			}
		}
	}
}

func (m *ModuleKafkaService) handleProducerMsg(fromIdentify string, msg *sarama.ProducerMessage) {
	if m.isSync {
		if len(m.queue) < m.bucketSize {
			m.queue = append(m.queue, msg)
			return
		}
		reportQueue := m.queue
		m.reportChan <- reportQueue
		m.queue = make([]*sarama.ProducerMessage, 0, m.bucketSize)
	} else {
		t1, failed := time.Now(), false
		timer := time.NewTimer(m.syncOvertime)
		select {
		case m.asyncProducer.Input() <- msg:
		case <-timer.C:
			failed = true
		}
		syncMS := time.Since(t1).Milliseconds()
		if failed || syncMS >= 100 {
			m.Logger().Info("handleProducerMsg, async with long duration", zap.Any("failed", failed), zap.Any("syncMS", syncMS))
		}
	}
}

func (m *ModuleKafkaService) handleAsyncProducerReturnSuccesses() {
	for {
		select {
		case <-m.asyncProducer.Successes():
		case <-m.Ctx().Done():
			for range m.asyncProducer.Successes() {

			}
		}
	}
}

func (m *ModuleKafkaService) handleAsyncProducerReturnErrors() {
	for {
		select {
		case <-m.asyncProducer.Errors():

		case <-m.Ctx().Done():
			for range m.asyncProducer.Errors() {

			}
		}
	}
}
