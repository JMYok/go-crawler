package collect

import (
	"errors"
	"time"
)

type Task struct {
	Url      string
	Cookie   string
	WaitTime time.Duration
	MaxDepth int
	RootReq  *Request
	Fetcher  Fetcher
}

// 单个请求
type Request struct {
	Task      *Task
	Url       string
	Depth     int
	ParseFunc func([]byte, *Request) ParseResult
}

type ParseResult struct {
	Requests []*Request    //当前url请求中，包含的新的请求
	Items    []interface{} //获取到的数据
}

func (r *Request) Check() error {
	if r.Depth > r.Task.MaxDepth {
		return errors.New("Max depth limit reached")
	}
	return nil
}
