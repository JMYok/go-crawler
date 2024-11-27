package geekbang

import (
	"go-crawler/collect"
	"regexp"
	"strings"
)

var GeekBangTask = &collect.Task{
	Property: collect.Property{
		Name:     "geekbang_article",
		WaitTime: 2,
		MaxDepth: 1,
		Cookie:   "_ga=GA1.2.1599052755.1702431092; LF_ID=5f0e727-b912600-bb65e95-1f2b820; mantis5539=cf985a2746c7457b8b1580eacfec22fb@5539; MEIQIA_TRACK_ID=2a0z6O1sHDKQDuJb2pjLySxpIPT; MEIQIA_VISIT_ID=2a0z6RcH5qksgkvdHIdmSrfYIft; _tea_utm_cache_20000743={%22utm_source%22:%22geektime_search%22%2C%22utm_medium%22:%22geektime_search%22%2C%22utm_campaign%22:%22geektime_search%22%2C%22utm_term%22:%22geektime_search%22%2C%22utm_content%22:%22geektime_search%22}; _gcl_au=1.1.417825686.1728531518; _ga_MTX5SQH9CV=GS1.2.1728531518.3.0.1728531518.0.0.0; GCID=f2ab571-0b192ea-9b4f955-c2b6e81; GRID=f2ab571-0b192ea-9b4f955-c2b6e81; _gid=GA1.2.862769862.1731403818; _ga_JW698SFNND=GS1.2.1732185714.17.1.1732185714.0.0.0; GCESS=Bg0BAQkBAQsCBgACBHUOP2cIAQMHBKpZOAUGBJTsausFBAAAAAABCBVsGwAAAAAABAQAjScADAEBAwR1Dj9nCgQAAAAA; tfstk=f0WjZj15Sq0j3vF4OcEyPGjr5gJ1CNwU1ctOxGHqXKpvCF_OrjRw7G3sCw70jojN3_T1yaTvQKW4C19cUskwurv1XwJ_8yyULijDsdUU8gwbyRppDn3t7nK-y2I3MMwULijYpvbK6JWVysjBRFp9MCKJeUxZWdK9WuTJfHctkNQO2utBfVKvDdH-eHTJWdpOWghorDt_GEjb48lyu9epOiLSdMXkDXTQKEHtOTtfVeIY9AHOFnOAQ_8GEtRV6MAh3GysEKSCwLK1Ikh65B1O3LBYyJdk61Ce1_uY0wQ12NvVelhvJs72etCIXAIWhsdMcQgbPeBc29vJaJEC2TbVmTsZXRKP8UIcHdwLxK9vkLtlQzDv5_CO3IvikPTGFgBA1goqLeNxMfwpUAtW8uZSsfjySJ6NcQe_pIKkcQr7VqCMM3xWRuZSldOvqnTLVugAN; gksskpitn=40fad00b-d850-4550-9014-8ca291cc62ec; Hm_lvt_59c4ff31a9ee6263811b23eb921a5083=1732105869,1732110956,1732185720,1732263449; HMACCOUNT=8CE909B64FF0A097; Hm_lvt_022f847c4e3acd44d4a2481d9187f1e6=1732105869,1732110956,1732185720,1732263449; __tea_cache_tokens_20000743={%22web_id%22:%227428819269044945156%22%2C%22user_unique_id%22:%221797141%22%2C%22timestamp%22:1732275462697%2C%22_type_%22:%22default%22}; Hm_lpvt_59c4ff31a9ee6263811b23eb921a5083=1732275463; Hm_lpvt_022f847c4e3acd44d4a2481d9187f1e6=1732275463; _ga_03JGDGP9Y3=GS1.2.1732275002.95.1.1732275462.0.0.0; SERVERID=3431a294a18c59fc8f5805662e2bd51e|1732276245|1732274999",
	},
	Rule: collect.RuleTree{
		Root: func() ([]*collect.Request, error) {
			roots := []*collect.Request{
				&collect.Request{
					Priority: 1,
					Url:      "https://time.geekbang.com/column/article/646801",
					Method:   "GET",
					RuleName: "文章爬取规则",
				},
			}
			return roots, nil
		},
		Trunk: map[string]*collect.Rule{
			"文章爬取规则": &collect.Rule{
				ItemFields: []string{
					"文章标题",
					"文章内容",
					"文章代码",
				},
				ParseFunc: ParseArticle,
			},
		},
	},
}

const titleReg = `<title>`
const paragraphReg = `<div[^>]*data-slate-type="paragraph"[^>]*>.*?<span[^>]*data-slate-string="true">([^<]*)</span>`
const codeReg = `<div[^>]*data-slate-type="pre"[^>]*>.*?<span[^>]*data-slate-string="true">([^<]*)</span>`

// TODO: 1. 爬取得到的是乱码 2.正则匹配失效
func ParseArticle(ctx *collect.Context) (collect.ParseResult, error) {
	titleRe := regexp.MustCompile(titleReg)
	paragraphRe := regexp.MustCompile(paragraphReg)
	codeRe := regexp.MustCompile(codeReg)

	article := map[string]interface{}{
		"文章标题": ExtraString(ctx.Body, titleRe),
		"文章内容": ExtraString(ctx.Body, paragraphRe),
		"文章代码": ExtraString(ctx.Body, codeRe),
	}

	data := ctx.Output(article)

	return collect.ParseResult{
		Items: []interface{}{data},
	}, nil
}

func ExtraString(contents []byte, re *regexp.Regexp) string {
	matches := re.FindStringSubmatch(string(contents))
	var result []string
	for _, match := range matches {
		if len(match) > 1 {
			result = append(result, string(match[1])) // 仅提取内容部分
		}
	}
	return strings.Join(result, " ")
}
