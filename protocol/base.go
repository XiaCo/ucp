package protocol

type SpeedCalculator interface {
	GetSpeed() int64 // 获取当前速度
	AddFlow(int64)   // 增加流量
	Close()          // 关闭
}
