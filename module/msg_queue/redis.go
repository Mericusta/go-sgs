package moduleMsgQueue

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Mericusta/go-sgs"
	"github.com/Mericusta/go-sgs/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type ModuleRedisService struct {
	sgs.ModuleBase

	// RedisDB 的配置
	config *config.MsgQueueConfig

	// 与 RedisDB 建立的链接
	client *redis.Client

	// 订阅的 redis 的 subject
	redisSubjects []string

	// 订阅的 redis 的 subject 的实例
	subscription *redis.PubSub

	// redis 消息的回调
	redisMsgHandler func(sgs.Module, *redis.Message)
}

var RedisService *ModuleRedisService

func (*ModuleRedisService) New(mos ...sgs.ModuleOption) *ModuleRedisService {
	mns := &ModuleRedisService{}
	for _, mo := range mos {
		mo(mns)
	}
	return mns
}

func (*ModuleRedisService) WithRedisDBConfig(c *config.MsgQueueConfig) sgs.ModuleOption {
	return func(m sgs.Module) { m.(*ModuleRedisService).config = c }
}

func (*ModuleRedisService) WithRedisSubjects(s ...string) sgs.ModuleOption {
	return func(m sgs.Module) {
		m.(*ModuleRedisService).redisSubjects = append(m.(*ModuleRedisService).redisSubjects, s...)
	}
}

func (*ModuleRedisService) WithRedisMsgHandler(h func(sgs.Module, *redis.Message)) sgs.ModuleOption {
	return func(m sgs.Module) { m.(*ModuleRedisService).redisMsgHandler = h }
}

func (m *ModuleRedisService) Mounted() {
	m.Logger().Debug("OBSERVE: Mounted")

	// 检查消息回调
	if m.redisMsgHandler == nil {
		// 无回调不提供服务
		panic("redis msg handler not set")
	}

	// 和 RedisDB 建立链接
	redisDBIndex, err := strconv.Atoi(m.config.Option["db"])
	if err != nil {
		// 配置错误不提供服务
		panic(err)
	}
	redisDBClient := redis.NewClient(&redis.Options{
		Addr: m.config.URL,
		DB:   redisDBIndex,
	})

	// ping 一下检测链接
	result, err := redisDBClient.Ping(m.Ctx()).Result()
	if strings.ToUpper(result) != "PONG" || err != nil {
		// 链接不通不提供服务
		panic(fmt.Sprintf("connect redis %v failed, result %v, err %v", m.config.URL, result, err))
	}

	// 保存 RedisDB 链接的客户端
	m.client = redisDBClient

	// 开启 redis 订阅
	m.subscription = m.client.Subscribe(m.Ctx(), m.redisSubjects...)
	if m.subscription == nil {
		// 订阅失败不提供服务
		panic(fmt.Sprintf("subscribe subject %v failed", m.redisSubjects))
	}
}

func (m *ModuleRedisService) Run() {
	for redisMsg := range m.subscription.Channel() {
		if redisMsg == nil {
			m.Logger().Error("redis subscription receive nil redis message")
			continue
		}
		m.redisMsgHandler(m, redisMsg)
	}
}

func (m *ModuleRedisService) Unmounted() {
	m.Logger().Debug("OBSERVE: Unmounted")

	// 取消订阅
	if m.subscription != nil {
		err := m.subscription.Close()
		if err != nil {
			m.Logger().Error("redis subscription close occurs error", zap.Error(err), zap.Any("subjects", m.redisSubjects))
		}
	}
}
