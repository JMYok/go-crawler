package collect

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"go-crawler/collector"
	"regexp"
	"sync"
	"time"
)

type Property struct {
	Name     string        `json:"name"` // 任务名称，应保证唯一性
	Url      string        `json:"url"`
	Cookie   string        `json:"cookie"`
	WaitTime time.Duration `json:"wait_time"`
	Reload   bool          `json:"reload"` // 网站是否可以重复爬取
	MaxDepth int64         `json:"max_depth"`
}

// 一个任务实例，
type Task struct {
	Property
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

func (c *Context) ParseJSReg(name string, reg string) ParseResult {
	re := regexp.MustCompile(reg)

	matches := re.FindAllSubmatch(c.Body, -1)
	result := ParseResult{}

	for _, m := range matches {
		u := string(m[1])
		result.Requests = append(
			result.Requests, &Request{
				Method:   "GET",
				Task:     c.Req.Task,
				Url:      u,
				Depth:    c.Req.Depth + 1,
				RuleName: name,
			})
	}
	return result
}

func (c *Context) OutputJS(reg string) ParseResult {
	re := regexp.MustCompile(reg)
	ok := re.Match(c.Body)
	if !ok {
		return ParseResult{
			Items: []interface{}{},
		}
	}
	result := ParseResult{
		Items: []interface{}{c.Req.Url},
	}
	return result
}

func (c *Context) GetRule(ruleName string) *Rule {
	return c.Req.Task.Rule.Trunk[ruleName]
}

// 将数据封装为collector.OutputData,用于之后存储
func (c *Context) Output(data interface{}) *collector.OutputData {
	res := &collector.OutputData{}
	res.Data = make(map[string]interface{})
	res.Data["Rule"] = c.Req.RuleName
	res.Data["Data"] = data
	res.Data["Url"] = c.Req.Url
	res.Data["Time"] = time.Now().Format("2006-01-02 15:04:05")
	return res
}

// 单个请求
type Request struct {
	unique    string
	Task      *Task
	Url       string
	Method    string
	Depth     int64
	Priority  int64
	RuleName  string // 当前请采取的解析规则
	ParseFunc func([]byte, *Request) ParseResult
	TmpData   *Temp
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
