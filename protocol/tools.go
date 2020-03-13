package protocol

import (
	"math"
	"os"
	"sync/atomic"
	"time"
)

func SplitFile(size int64) []int64 {
	// 分割文件成小块编号
	max := int64(math.Ceil(float64(size) / float64(SplitFileSize)))
	res := make([]int64, max)
	for i := int64(0); i < max; i++ {
		res[i] = i
	}
	return res
}

func SavePathIsValid(p string) bool {
	_, statErr := os.Stat(p)
	if statErr != nil {
		return !os.IsExist(statErr)
	}
	return false
}

func SleepAfterSendPackage(n int64, sendSpeed int64) func() {
	// n: 多少次作为一批发送的数据
	// sendSpeed: 需要控制的速率，单位 kb/s
	count := int64(0)
	sleepTime := time.Duration(n * 1000000000 / sendSpeed)
	return func() {
		count++
		if count == n {
			time.Sleep(sleepTime)
			count = 0
		}
	}
}

type speedCalculator struct {
	flow  int64
	speed int64
	t     *time.Ticker
	over  chan struct{}
}

func (s *speedCalculator) GetSpeed() int64 {
	return atomic.LoadInt64(&s.speed)
}

func (s *speedCalculator) AddFlow(n int64) {
	atomic.AddInt64(&s.flow, n)
}

func (s *speedCalculator) Close() {
	s.t.Stop()
	s.over <- struct{}{}
}

func NewSpeedCalculator(t time.Duration) *speedCalculator {
	delay := time.NewTicker(t)
	s := speedCalculator{0, 0, delay, make(chan struct{})}
	go func() {
		for {
			select {
			case <-delay.C:
				f := atomic.LoadInt64(&s.flow)
				atomic.StoreInt64(&s.speed, f/int64(t.Seconds()))
				atomic.StoreInt64(&s.flow, 0)
			case <-s.over:
				return
			}
		}
	}()
	return &s
}
