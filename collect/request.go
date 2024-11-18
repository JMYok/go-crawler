package collect

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

// 一个任务实例，
type Task struct {
	Name        string
	Url         string
	Cookie      string
	WaitTime    time.Duration
	Reload      bool // 网站是否可以重复爬取
	MaxDepth    int
	Visited     map[string]bool
	VisitedLock sync.Mutex
	RootReq     *Request
	Fetcher     Fetcher
	Rule        RuleTree
}

type Context struct {
	Body []byte
	Req  *Request
}

// 单个请求
type Request struct {
	unique    string
	Task      *Task
	Url       string
	Method    string
	Depth     int
	Priority  int
	RuleName  string // 当前请采取的解析规则
	ParseFunc func([]byte, *Request) ParseResult
}

type ParseResult struct {
	Requests []*Request    //当前url请求中，包含的新的请求
	Items    []interface{} //获取到的数据
}

func (r *Request) CheckDepth() error {
	if r.Depth > r.Task.MaxDepth {
		return errors.New("Max depth limit reached")
	}
	return nil
}

// 请求的唯一识别码
func (r *Request) Unique() string {
	block := md5.Sum([]byte(r.Url + r.Method))
	return hex.EncodeToString(block[:])
}
