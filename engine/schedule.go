package engine

import (
	"go-crawler/collect"
	"go.uber.org/zap"
)

type ScheduleEngine struct {
	requestCh chan *collect.Request
	workerCh  chan *collect.Request
	out       chan collect.ParseResult
	options
}

func NewSchedule(opts ...Option) *ScheduleEngine {
	// 默认设置
	options := defaultOptions

	//自定义设置加入到options中
	for _, opt := range opts {
		opt(&options)
	}
	s := &ScheduleEngine{}
	s.options = options
	return s
}

func (s *ScheduleEngine) Run() {
	requestCh := make(chan *collect.Request)
	workerCh := make(chan *collect.Request)
	out := make(chan collect.ParseResult)
	s.requestCh = requestCh
	s.workerCh = workerCh
	s.out = out
	go s.Schedule()
	// 实现多worker并行处理。（串行的从s.workerCh取任务）
	for i := 0; i < s.WorkCount; i++ {
		go s.CreateWork()
	}
	s.HandleResult()
}

// 从请求通道s.requestCh取请求到reqQueue，送到s.workerCh等待执行
func (s *ScheduleEngine) Schedule() {
	var reqQueue = s.Seeds
	go func() {
		for {
			var req *collect.Request
			var ch chan *collect.Request

			if len(reqQueue) > 0 {
				req = reqQueue[0]
				ch = s.workerCh
			}
			select {
			// 收到新请求，阻塞终止，下一个循环就能取到req，从而执行ch <- req
			case r := <-s.requestCh:
				reqQueue = append(reqQueue, r)
			// 若req=nil，往 nil 通道中写入数据会陷入到堵塞的状态，直到接收到新的请求才会被唤醒。
			case ch <- req:
				reqQueue = reqQueue[1:]
			}
		}
	}()
}

// 从s.workerCh取请求执行，结果放到s.out
func (s *ScheduleEngine) CreateWork() {
	for {
		r := <-s.workerCh
		if err := r.Check(); err != nil {
			s.Logger.Error("check failed", zap.Error(err))
		}
		body, err := s.Fetcher.Get(r)
		if err != nil {
			s.Logger.Error("can't fetch ",
				zap.Error(err),
			)
			continue
		}
		result := r.ParseFunc(body, r)
		s.out <- result
	}
}

func (s *ScheduleEngine) HandleResult() {
	for {
		select {
		case result := <-s.out:
			for _, req := range result.Requests {
				s.requestCh <- req
			}
			for _, item := range result.Items {
				// todo: store
				s.Logger.Sugar().Info("get result ", item)
			}
		}
	}
}
