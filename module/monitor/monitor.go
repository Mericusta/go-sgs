package moduleMonitor

import (
	"strconv"

	"github.com/Mericusta/go-sgs"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// ModuleMonitorService 监控服务模块
// - 监控
type ModuleMonitorService struct {
	// 组合基础模块
	sgs.ModuleBase

	// 上报慢请求的阈值，单位毫秒
	reportSlowMS int64

	// prometheus 自定义收集器
	collectors []prometheus.Collector
}

var Service *ModuleMonitorService

func (*ModuleMonitorService) New(mos ...sgs.ModuleOption) *ModuleMonitorService {
	mns := &ModuleMonitorService{}
	for _, mo := range mos {
		mo(mns)
	}
	return mns
}

func (*ModuleMonitorService) WithReportSlowMS(ms int64) sgs.ModuleOption {
	return func(m sgs.Module) { m.(*ModuleMonitorService).reportSlowMS = ms }
}

func (*ModuleMonitorService) WithPrometheusCollector(c prometheus.Collector) sgs.ModuleOption {
	return func(m sgs.Module) {
		m.(*ModuleMonitorService).collectors = append(m.(*ModuleMonitorService).collectors, c)
	}
}

func (m *ModuleMonitorService) Mounted() {
	m.Logger().Debug("OBSERVE: Mounted", zap.Any("reportSlowMS", m.reportSlowMS), zap.Any("collectorsCount", len(m.collectors)))

	// 注册收集器
	for _, collector := range m.collectors {
		prometheus.MustRegister(collector)
	}
}

func (m *ModuleMonitorService) HandleEvent(event *sgs.ModuleEvent) {
	switch data := event.Data().(type) {
	case IHandleReportBehavior:
		m.handleRequestReportBehavior(data)
	default:
		m.Logger().Error(sgs.ErrorMsgHandleEventNonImplement)
	}
}

func (m *ModuleMonitorService) Run() {
}

func (m *ModuleMonitorService) handleRequestReportBehavior(requestReportBehavior IHandleReportBehavior) {
	if m.reportSlowMS <= 0 {
		return
	}

	reporterIdentify := requestReportBehavior.ReporterIdentify()
	messageType := requestReportBehavior.MessageType()
	gateReceiveRequestMSStr := requestReportBehavior.GateReceiveRequestMS()
	gameReceiveRequestMSStr := requestReportBehavior.GameReceiveRequestMS()
	gameHandleRequestMSStr := requestReportBehavior.GameHandleRequestMS()
	gateReceiveResponseMSStr := requestReportBehavior.GateReceiveResponseMS()

	if len(reporterIdentify) == 0 || len(messageType) == 0 || len(gateReceiveRequestMSStr) == 0 || len(gameReceiveRequestMSStr) == 0 || len(gameHandleRequestMSStr) == 0 || len(gateReceiveResponseMSStr) == 0 {
		m.Logger().Error("handleRequestReportBehavior, invalid report", zap.Any("reporterIdentify", reporterIdentify), zap.Any("messageType", messageType), zap.Any("gateReceiveRequestMSStr", gateReceiveRequestMSStr), zap.Any("gameReceiveRequestMSStr", gameReceiveRequestMSStr), zap.Any("gameHandleRequestMSStr", gameHandleRequestMSStr), zap.Any("gateReceiveResponseMSStr", gateReceiveResponseMSStr))
		return
	}

	s, err := strconv.ParseInt(gateReceiveRequestMSStr, 10, 64)
	if err != nil {
		m.Logger().Error("handleRequestReportBehavior, ParseInt failed", zap.Error(err), zap.Any("reporterIdentify", reporterIdentify), zap.Any("gateReceiveRequestMSStr", gateReceiveRequestMSStr))
		return
	}
	r, err := strconv.ParseInt(gameReceiveRequestMSStr, 10, 64)
	if err != nil {
		m.Logger().Error("handleRequestReportBehavior, ParseInt failed", zap.Error(err), zap.Any("reporterIdentify", reporterIdentify), zap.Any("gameReceiveRequestMSStr", gameReceiveRequestMSStr))
		return
	}
	h, err := strconv.ParseInt(gameHandleRequestMSStr, 10, 64)
	if err != nil {
		m.Logger().Error("handleRequestReportBehavior, ParseInt failed", zap.Error(err), zap.Any("reporterIdentify", reporterIdentify), zap.Any("gameHandleRequestMSStr", gameHandleRequestMSStr))
		return
	}
	f, err := strconv.ParseInt(gateReceiveResponseMSStr, 10, 64)
	if err != nil {
		m.Logger().Error("handleRequestReportBehavior, ParseInt failed", zap.Error(err), zap.Any("reporterIdentify", reporterIdentify), zap.Any("gateReceiveResponseMSStr", gateReceiveResponseMSStr))
		return
	}

	d1, d2, d3 := r-s, h-r, f-h
	t := d1 + d2 + d3
	if t >= m.reportSlowMS {
		m.Logger().Info("handleRequestReportBehavior, slow request report", zap.Any("reporterIdentify", reporterIdentify), zap.Any("message", messageType), zap.Any("total", t), zap.Any("gate -> nats -> game", d1), zap.Any("game -> handler -> game", d2), zap.Any("game -> nats -> gate", d3))
	}
}
