package moduleServiceDiscovery

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Mericusta/go-sgs"
	"github.com/Mericusta/go-sgs/config"
	"github.com/redis/go-redis/v9"
)

type ModuleRedisService struct {
	sgs.ModuleBase

	// RedisDB 的配置
	config *config.DBConfig

	// 与 RedisDB 建立的链接
	client *redis.Client

	// 服务发现的 key
	serviceDiscoveryKey string
}

var Service *ModuleRedisService

func (*ModuleRedisService) New(mos ...sgs.ModuleOption) *ModuleRedisService {
	mns := &ModuleRedisService{}
	for _, mo := range mos {
		mo(mns)
	}
	return mns
}

func (*ModuleRedisService) WithRedisDBConfig(c *config.DBConfig) sgs.ModuleOption {
	return func(m sgs.Module) { m.(*ModuleRedisService).config = c }
}

func (*ModuleRedisService) WithServiceDiscoveryKey(s string) sgs.ModuleOption {
	return func(m sgs.Module) { m.(*ModuleRedisService).serviceDiscoveryKey = s }
}

func (m *ModuleRedisService) Mounted() {
	m.Logger().Debug("OBSERVE: Mounted")

	// 和 RedisDB 建立链接
	redisDBIndex, err := strconv.Atoi(m.config.Database)
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
}

func (m *ModuleRedisService) HandleEvent(event *sgs.ModuleEvent) {
	// switch data := event.Data().(type) {
	// default:
	// 	m.Logger().Error(sgs.ErrorMsgHandleEventNonImplement)
	// }
}

func (m *ModuleRedisService) Unmounted() {
	m.Logger().Debug("OBSERVE: Unmounted")
}
