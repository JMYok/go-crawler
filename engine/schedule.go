package engine

import (
	"github.com/robertkrimen/otto"
	"go-crawler/collect"
	"go-crawler/parse/doubanbook"
	"go-crawler/parse/doubangroup"
	"go.uber.org/zap"
	"sync"
)

// 注册任务
func init() {
	Store.Add(doubanbook.DoubanBookTask)
	Store.Add(doubangroup.DoubangroupTask)
	Store.AddJSTask(doubangroup.DoubangroupJSTask)
}

// 全局爬虫任务实例
var Store = &CrawlerStore{
	list: []*collect.Task{},
	hash: map[string]*collect.Task{},
}

type CrawlerStore struct {
	list []*collect.Task
	hash map[string]*collect.Task
}

func (c *CrawlerStore) Add(task *collect.Task) {
	c.hash[task.Name] = task
	c.list = append(c.list, task)
}

// 用于动态规则添加请求。
func AddJsReqs(jreqs []map[string]interface{}) []*collect.Request {
	reqs := make([]*collect.Request, 0)

	for _, jreq := range jreqs {
		req := &collect.Request{}
		u, ok := jreq["Url"].(string)
		if !ok {
			return nil
		}
		req.Url = u
		req.RuleName, _ = jreq["RuleName"].(string)
		req.Method, _ = jreq["Method"].(string)
		req.Priority, _ = jreq["Priority"].(int64)
		reqs = append(reqs, req)
	}
	return reqs
}

// 用于动态规则添加请求。
func AddJsReq(jreq map[string]interface{}) []*collect.Request {
	reqs := make([]*collect.Request, 0)
	req := &collect.Request{}
	u, ok := jreq["Url"].(string)
	if !ok {
		return nil
	}
	req.Url = u
	req.RuleName, _ = jreq["RuleName"].(string)
	req.Method, _ = jreq["Method"].(string)
	req.Priority, _ = jreq["Priority"].(int64)
	reqs = append(reqs, req)
	return reqs
}

func (c *CrawlerStore) AddJSTask(m *collect.TaskModle) {
	task := &collect.Task{
		Property: m.Property,
	}

	task.Rule.Root = func() ([]*collect.Request, error) {
		vm := otto.New()

		// 将 Go 原生函数注册到 JS 虚拟机中
		vm.Set("AddJsReq", AddJsReqs)

		// 执行js脚本,返回了生成的请求数组
		v, err := vm.Eval(m.Root)
		if err != nil {
			return nil, err
		}

		//  转换为go数据结构
		e, err := v.Export()
		if err != nil {
			return nil, err
		}
		return e.([]*collect.Request), nil
	}

	for _, r := range m.Rules {
		//使用了闭包，方便后续执行 parse 脚本并最后返回解析后的 collect.ParseResult 结果
		parseFunc := func(parse string) func(ctx *collect.Context) (collect.ParseResult, error) {
			return func(ctx *collect.Context) (collect.ParseResult, error) {
				vm := otto.New()
				vm.Set("ctx", ctx)
				v, err := vm.Eval(parse)
				if err != nil {
					return collect.ParseResult{}, err
				}
				e, err := v.Export()
				if err != nil {
					return collect.ParseResult{}, err
				}
				if e == nil {
					return collect.ParseResult{}, err
				}
				return e.(collect.ParseResult), err
			}
		}(r.ParseFunc)
		if task.Rule.Trunk == nil {
			task.Rule.Trunk = make(map[string]*collect.Rule, 0)
		}
		task.Rule.Trunk[r.Name] = &collect.Rule{
			ParseFunc: parseFunc,
		}
	}

	c.hash[task.Name] = task
	c.list = append(c.list, task)
}

type Crawler struct {
	out chan collect.ParseResult
	// 存储请求的唯一标识
	Visited     map[string]bool
	VisitedLock sync.Mutex
	failures    map[string]*collect.Request // 失败请求id -> 失败请求
	failureLock sync.Mutex
	options
}

type Scheduler interface {
	Schedule()
	Push(...*collect.Request)
	Pull() *collect.Request
}

type ScheduleEngine struct {
	requestCh   chan *collect.Request
	workerCh    chan *collect.Request
	reqQueue    []*collect.Request
	priReqQueue []*collect.Request
	Logger      *zap.Logger
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

// 从请求通道s.requestCh和优先队列s.priReqQueue取请求到reqQueue，送到s.workerCh等待执行
func (s *ScheduleEngine) Schedule() {
	// 放在请求外部，防止取到req后，没有走case ch <- req导致请求丢失的情况
	var req *collect.Request
	var ch chan *collect.Request
	for {
		if req == nil && len(s.priReqQueue) > 0 {
			req = s.priReqQueue[0]
			s.priReqQueue = s.priReqQueue[1:]
			ch = s.workerCh
		}
		if req == nil && len(s.reqQueue) > 0 {
			req = s.reqQueue[0]
			s.reqQueue = s.reqQueue[1:]
			ch = s.workerCh
		}
		select {
		case r := <-s.requestCh:
			if r.Priority > 0 {
				s.priReqQueue = append(s.priReqQueue, r)
			} else {
				s.reqQueue = append(s.reqQueue, r)
			}
		case ch <- req:
			req = nil
			ch = nil
		}
	}
}

// 从seeds中取得Requests，启动调度
func (e *Crawler) Schedule() {
	var reqs []*collect.Request
	for _, seed := range e.Seeds {
		// 根据seed.Name，从Store中取得具体task
		task := Store.hash[seed.Name]
		task.Fetcher = seed.Fetcher
		// 获取初始任务的 种子请求（初始url）
		rootreqs, err := task.Rule.Root()
		if err != nil {
			e.Logger.Error("get root failed",
				zap.Error(err),
			)
			continue
		}
		// 将请求和任务关联
		for _, req := range rootreqs {
			req.Task = task
		}
		reqs = append(reqs, rootreqs...)
	}
	go e.scheduler.Schedule()
	go e.scheduler.Push(reqs...)
}

// 从s.workerCh取请求执行，结果放到s.out
func (e *Crawler) CreateWork() {
	for {
		r := e.scheduler.Pull()
		if err := r.CheckDepth(); err != nil {
			e.Logger.Error("check  failed",
				zap.Error(err),
			)
			continue
		}
		// 判断当前请求是否已被访问
		if e.HasVisited(r) {
			e.Logger.Debug("request has visited", zap.String("url:", r.Url))
			continue
		}
		// 设置当前请求已被访问
		e.StoreVisited(r)
		body, err := r.Task.Fetcher.Get(r)
		if len(body) < 6000 {
			e.Logger.Error("can't fetch ",
				zap.Int("length", len(body)),
				zap.String("url", r.Url),
			)
			continue
		}
		if err != nil {
			e.Logger.Error("can't fetch ",
				zap.Error(err),
				zap.String("url", r.Url),
			)
			continue
		}

		// 当前请求解析规则
		rule := r.Task.Rule.Trunk[r.RuleName]

		result, err := rule.ParseFunc(&collect.Context{
			body,
			r,
		})

		if err != nil {
			e.Logger.Error("ParseFunc failed ",
				zap.Error(err),
				zap.String("url", r.Url),
			)
		}

		if len(result.Requests) > 0 {
			go e.scheduler.Push(result.Requests...)
		}

		e.out <- result
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

func (e *Crawler) SetFailure(req *collect.Request) {
	if !req.Task.Reload {
		e.VisitedLock.Lock()
		unique := req.Unique()
		delete(e.Visited, unique)
		e.VisitedLock.Unlock()
	}
	e.failureLock.Lock()
	defer e.failureLock.Unlock()
	if _, ok := e.failures[req.Unique()]; !ok {
		// 首次失败时，再重新执行一次
		e.failures[req.Unique()] = req
		e.scheduler.Push(req)
	}
	// todo: 失败2次，加载到失败队列中
}
