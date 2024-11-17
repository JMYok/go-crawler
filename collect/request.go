package collect

import "time"

type Request struct {
	Url       string
	Cookie    string
	WaitTime  time.Duration
	ParseFunc func([]byte, *Request) ParseResult
}

type ParseResult struct {
	Requests []*Request    //当前url请求中，包含的新的请求
	Items    []interface{} //获取到的数据
}
