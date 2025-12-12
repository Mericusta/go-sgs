package moduleMsgQueue

import (
	"fmt"
	"net/http"

	"github.com/Mericusta/go-sgs"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	producer "github.com/kinesis-producer-go/kinesis-producer"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 封装 zap -> kinesis 的 WriteSyncer

type kinesisWriter struct {
	pr *producer.Producer
}

// 实现 producer.Producer 的接口
func (kw *kinesisWriter) Write(p []byte) (int, error) {
	// 同步调用 producer.Put，内部会走聚合、重试
	if err := kw.pr.Put(p); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (kw *kinesisWriter) Sync() error {
	// producer 自带 Flush/Stop，这里留空即可
	return nil
}

// 封装 zap 消息

type KinesisMsg struct {
	appName string
	logType string
	ts      int64
	content string
	fields  []zap.Field
}

func NewKinesisMsg(appName, logType string, ts int64, content string, fields ...zap.Field) *KinesisMsg {
	return &KinesisMsg{appName: appName, logType: logType, ts: ts, content: content, fields: fields}
}

// 目标 s3 桶地址: [region][stream]
// 目标 s3 桶中具体路径: [appName]/[logType]/[ts -> Year]/[ts -> Month]/[ts -> Day]/[ts -> Hour]

type ModuleKinesisService struct {
	sgs.ModuleBase

	region     string
	stream     string
	bucketSize int

	producer *producer.Producer
	logger   *zap.Logger
}

var KinesisService *ModuleKinesisService

func (*ModuleKinesisService) New(mos ...sgs.ModuleOption) *ModuleKinesisService {
	mks := &ModuleKinesisService{}
	for _, mo := range mos {
		mo(mks)
	}
	return mks
}

func (*ModuleKinesisService) WithRegion(region string) sgs.ModuleOption {
	return func(m sgs.Module) { m.(*ModuleKinesisService).region = region }
}

func (*ModuleKinesisService) WithStream(stream string) sgs.ModuleOption {
	return func(m sgs.Module) { m.(*ModuleKinesisService).stream = stream }
}

func (*ModuleKinesisService) WithBucketSize(size int) sgs.ModuleOption {
	return func(m sgs.Module) { m.(*ModuleKinesisService).bucketSize = size }
}

func (*ModuleKinesisService) WithLogger(logger *zap.Logger) sgs.ModuleOption {
	return func(m sgs.Module) { m.(*ModuleKinesisService).logger = logger }
}

func (m *ModuleKinesisService) Mounted() {
	m.Logger().Debug("OBSERVE: Mounted")

	// AWS SDK v2 config
	cfg, err := config.LoadDefaultConfig(m.Ctx(), config.WithRegion(m.region))
	if err != nil {
		panic(err)
	}

	// 构造 aws 客户端
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = 500
	transport.MaxIdleConnsPerHost = 500
	cfg.HTTPClient = &http.Client{Transport: transport}
	awsClient := kinesis.NewFromConfig(cfg)

	// 生成 kinesis-producer 配置
	if m.bucketSize <= 0 {
		m.bucketSize = 1024
	}
	producerConfig := &producer.Config{
		StreamName:         aws.String(m.stream),
		BacklogCount:       m.bucketSize, // 默认 1024 条缓冲
		Client:             awsClient,
		BatchCount:         64,
		BatchSize:          1 << 20,
		AggregateBatchSize: 200 * (1 << 10),
	}
	m.Logger().Info("ModuleKinesisService.Mounted", zap.Any("producerConfig", producerConfig))

	// 构造 kinesis-producer 实例
	m.producer = producer.New(producerConfig)
	m.producer.Start() // 后台 goroutine 开始聚合 & 发送

	if m.logger == nil {
		// 构造 kinesis 专用的 zap core
		encCfg := zap.NewProductionEncoderConfig()
		encCfg.TimeKey = "ts"
		encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		m.logger = zap.New(zapcore.NewCore(
			zapcore.NewJSONEncoder(encCfg),
			zapcore.AddSync(&kinesisWriter{pr: m.producer}),
			zapcore.InfoLevel,
		), zap.AddCaller())
	}
}

func (m *ModuleKinesisService) HandleEvent(event *sgs.ModuleEvent) {
	switch data := event.Data().(type) {
	case *KinesisMsg:
		m.handleKinesisMsg(data)
	default:
		m.Logger().Error(sgs.ErrorMsgHandleEventNonImplement)
	}
}

func (m *ModuleKinesisService) handleKinesisMsg(msg *KinesisMsg) {
	if msg == nil {
		return
	}

	fields := make([]zap.Field, 0, len(msg.fields)+3)
	fields = append(fields,
		zap.String("app_name", msg.appName),
		zap.String("log_type", msg.logType),
		zap.Int64("time_unix", msg.ts),
	)
	fields = append(fields, msg.fields...)
	m.logger.Info(msg.content, fields...)
	m.Logger().Debug(fmt.Sprintf("handleKinesisMsg %v", msg.content), fields...)
}

func (m *ModuleKinesisService) Unmounted() {
	m.Logger().Debug("OBSERVE: Unmounted")

	err := m.logger.Sync() // 同步 zap.logger 中所有数据
	if err != nil {
		m.Logger().Error("Unmounted, logger.sync occurs error", zap.Error(err))
	}
	m.producer.Stop() // 停止并且 flush 剩余数据
}
