package collect

type Request struct {
	Url       string
	Cookie    string
	ParseFunc func([]byte) ParseResult
}

type ParseResult struct {
	Requests []*Request //用于进一步获取数据
	Items    []interface{}
}
