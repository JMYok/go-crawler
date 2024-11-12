package doubangroup

import (
	"go-crawler/collect"
	"regexp"
)

// 获取所有帖子的 URL
const urlListRe = `(https://www.douban.com/group/topic/[0-9a-z]+/)"[^>]*>([^<]+)</a>`

func ParseURL(contents []byte, req *collect.Request) collect.ParseResult {
	re := regexp.MustCompile(urlListRe)

	matches := re.FindAllSubmatch(contents, -1)
	result := collect.ParseResult{}

	for _, m := range matches {
		u := string(m[1])
		// 组装到一个新的 Request 中，用作下一步的爬取。
		result.Requests = append(
			result.Requests,
			&collect.Request{
				Url:    u,
				Cookie: req.Cookie,
				ParseFunc: func(c []byte, request *collect.Request) collect.ParseResult {
					return GetContent(c, u)
				},
			})
	}
	return result
}

const ContentRe = `<div class="topic-content">[\s\S]*?阳台[\s\S]*?<div`

func GetContent(contents []byte, url string) collect.ParseResult {
	re := regexp.MustCompile(ContentRe)

	ok := re.Match(contents)
	if !ok {
		return collect.ParseResult{
			Items: []interface{}{},
		}
	}

	// 发现正文中有对应的文字，就将当前帖子的 URL 写入到 Items 当中
	result := collect.ParseResult{
		Items: []interface{}{url},
	}

	return result
}
