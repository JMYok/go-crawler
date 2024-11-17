package collect

import (
	"errors"
	"time"
)

type Request struct {
	Url       string
	Cookie    string
	Depth     int
	MaxDepth  int
	WaitTime  time.Duration
	ParseFunc func([]byte, *Request) ParseResult
}

type ParseResult struct {
	Requests []*Request    //当前url请求中，包含的新的请求
	Items    []interface{} //获取到的数据
}

func (r *Request) Check() error {
	if r.Depth > r.MaxDepth {
		return errors.New("Max depth limit reached")
	}
	return nil
}
