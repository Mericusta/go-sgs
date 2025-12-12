package moduleDB

import (
	"context"
	"fmt"
	"time"

	"github.com/Mericusta/go-sgs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// ModuleMongoService mongo 服务模块
type ModuleMongoService struct {
	sgs.ModuleBase

	// MongoDB 的地址
	url string

	// MongoDB 请求的超时时间
	dialTimeout time.Duration

	readPref *readpref.ReadPref

	// 与 MongoDB 建立的链接
	client *mongo.Client
}

var Service *ModuleMongoService

func (*ModuleMongoService) New(mos ...sgs.ModuleOption) *ModuleMongoService {
	mms := &ModuleMongoService{}
	for _, mo := range mos {
		mo(mms)
	}
	return mms
}

func (*ModuleMongoService) WithURL(a string) sgs.ModuleOption {
	return func(m sgs.Module) { m.(*ModuleMongoService).url = a }
}

func (*ModuleMongoService) WithDialTimeout(dt time.Duration) sgs.ModuleOption {
	return func(m sgs.Module) { m.(*ModuleMongoService).dialTimeout = dt }
}

func (*ModuleMongoService) WithReadPref(rp *readpref.ReadPref) sgs.ModuleOption {
	return func(m sgs.Module) { m.(*ModuleMongoService).readPref = rp }
}

func (m *ModuleMongoService) Mounted() {
	// 和 MongoDB 建立链接
	ctx, canceler := context.WithTimeout(m.Ctx(), m.dialTimeout)
	bsonOptions := &options.BSONOptions{
		UseJSONStructTags: true,
		NilMapAsEmpty:     true,
		NilSliceAsEmpty:   true,
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(m.url).SetReadPreference(m.readPref).SetBSONOptions(bsonOptions))
	canceler()

	if err != nil {
		// 无法建立链接不提供服务
		panic(fmt.Errorf("connect to mongo %v occurs error: %v", m.url, err))
	}

	// 保存 MongoDB 链接客户端
	m.client = client

	// 给名为 userprofiles 的 collection 创建索引
	collection := m.client.Database("MansionStory").Collection("userprofiles")
	ctx, cancel := context.WithTimeout(context.Background(), m.dialTimeout)
	_, err = collection.Indexes().CreateMany(ctx, append([]mongo.IndexModel{}, mongo.IndexModel{
		Keys:    bson.M{"player_name": 1}, // 以 player_name 字段为索引，1表示升序
		Options: options.Index().SetUnique(true),
	}))
	cancel()

	if err != nil {
		// 无法建立链接不提供服务
		panic(err)
	}
}

func (p *ModuleMongoService) HandleEvent(e *sgs.ModuleEvent) {
	p.Logger().Error(sgs.ErrorMsgHandleEventNonImplement)
}
