package engine

import (
	"go-crawler/collect"
	"go.uber.org/zap"
	"sync"
)

type Crawler struct {
	out chan collect.ParseResult
	// 存储请求的唯一标识
	Visited     map[string]bool
	VisitedLock sync.Mutex
	options
}

type Scheduler interface {
	Schedule()
	Push(...*collect.Request)
	Pull() *collect.Request
}

type ScheduleEngine struct {
	requestCh chan *collect.Request
	workerCh  chan *collect.Request
	reqQueue  []*collect.Request
	Logger    *zap.Logger
}

func NewEngine(opts ...Option) *Crawler {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}
	e := &Crawler{}
	e.Visited = make(map[string]bool, 100)
	out := make(chan collect.ParseResult)
	e.out = out
	e.options = options
	return e
}

func NewSchedule() *ScheduleEngine {
	s := &ScheduleEngine{}
	requestCh := make(chan *collect.Request)
	workerCh := make(chan *collect.Request)
	s.requestCh = requestCh
	s.workerCh = workerCh
	return s
}

func (e *Crawler) Run() {
	go e.Schedule()
	for i := 0; i < e.WorkCount; i++ {
		go e.CreateWork()
	}
	e.HandleResult()
}

func (s *ScheduleEngine) Push(reqs ...*collect.Request) {
	for _, req := range reqs {
		s.requestCh <- req
	}
}

func (s *ScheduleEngine) Pull() *collect.Request {
	r := <-s.workerCh
	return r
}

func (s *ScheduleEngine) Output() *collect.Request {
	r := <-s.workerCh
	return r
}

// 从请求通道s.requestCh取请求到reqQueue，送到s.workerCh等待执行
func (s *ScheduleEngine) Schedule() {
	for {
		var req *collect.Request
		var ch chan *collect.Request

		if len(s.reqQueue) > 0 {
			req = s.reqQueue[0]
			ch = s.workerCh
		}
		select {
		case r := <-s.requestCh:
			s.reqQueue = append(s.reqQueue, r)
		case ch <- req:
			s.reqQueue = s.reqQueue[1:]
		}
	}
}

// 从seeds中取得Requests，启动调度
func (e *Crawler) Schedule() {
	var reqs []*collect.Request
	for _, seed := range e.Seeds {
		seed.RootReq.Task = seed
		seed.RootReq.Url = seed.Url
		reqs = append(reqs, seed.RootReq)
	}
	go e.scheduler.Schedule()
	go e.scheduler.Push(reqs...)
}

// 从s.workerCh取请求执行，结果放到s.out
func (s *Crawler) CreateWork() {
	for {
		r := s.scheduler.Pull()
		if err := r.CheckDepth(); err != nil {
			s.Logger.Error("check  failed",
				zap.Error(err),
			)
			continue
		}
		// 判断当前请求是否已被访问
		if s.HasVisited(r) {
			s.Logger.Debug("request has visited", zap.String("url:", r.Url))
			continue
		}
		// 设置当前请求已被访问
		s.StoreVisited(r)
		body, err := r.Task.Fetcher.Get(r)
		if len(body) < 6000 {
			s.Logger.Error("can't fetch ",
				zap.Int("length", len(body)),
				zap.String("url", r.Url),
			)
			continue
		}
		if err != nil {
			s.Logger.Error("can't fetch ",
				zap.Error(err),
				zap.String("url", r.Url),
			)
			continue
		}
		result := r.ParseFunc(body, r)

		if len(result.Requests) > 0 {
			go s.scheduler.Push(result.Requests...)
		}

		s.out <- result
	}
}

func (s *Crawler) HandleResult() {
	for {
		select {
		case result := <-s.out:
			for _, item := range result.Items {
				// todo: store
				s.Logger.Sugar().Info("get result: ", item)
			}
		}
	}
}

// 判断请求是否重复
func (e *Crawler) HasVisited(r *collect.Request) bool {
	e.VisitedLock.Lock()
	defer e.VisitedLock.Unlock()
	unique := r.Unique()
	return e.Visited[unique]
}

func (e *Crawler) StoreVisited(reqs ...*collect.Request) {
	e.VisitedLock.Lock()
	defer e.VisitedLock.Unlock()

	for _, r := range reqs {
		unique := r.Unique()
		e.Visited[unique] = true
	}
}
