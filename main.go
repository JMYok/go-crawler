package main

import (
	"fmt"
	"go-crawler/collect"
	"go-crawler/engine"
	"go-crawler/log"
	"go-crawler/parse/doubangroup"
	"go-crawler/proxy"
	"go.uber.org/zap/zapcore"
	"time"
)

func main() {
	// log
	plugin := log.NewStdoutPlugin(zapcore.InfoLevel)
	logger := log.NewLogger(plugin)
	logger.Info("log init end")

	// proxy
	proxyURLs := []string{"http://127.0.0.1:7890"}
	p, err := proxy.RoundRobinProxySwitcher(proxyURLs...)
	if err != nil {
		logger.Error("RoundRobinProxySwitcher failed")
	}

	// douban cookie
	cookie := "bid=TuWqA7AM2_4; viewed=\"1500149_26883690\"; _pk_id.100001.8cb4=939420b529734caa.1731404400.; __yadk_uid=eaEj42WRDNHKvtY84a4E4ic8v9KGYszX; __utmz=30149280.1731404404.3.3.utmcsr=time.geekbang.com|utmccn=(referral)|utmcmd=referral|utmcct=/column/article/612328; _pk_ref.100001.8cb4=%5B%22%22%2C%22%22%2C1731676163%2C%22https%3A%2F%2Ftime.geekbang.com%2Fcolumn%2Farticle%2F612328%22%5D; _pk_ses.100001.8cb4=1; ap_v=0,6.0; __utma=30149280.1825385725.1728983070.1731404404.1731676164.4; __utmc=30149280; dbcl2=\"150361748:Qt1XBWORkVA\"; ck=gzyI; push_noty_num=0; push_doumail_num=0; __utmv=30149280.15036; __utmb=30149280.13.5.1731676205938"
	var seeds []*collect.Request
	for i := 0; i <= 100; i += 25 {
		str := fmt.Sprintf("https://www.douban.com/group/szsh/discussion?start=%d", i)
		seeds = append(seeds, &collect.Request{
			Url:       str,
			WaitTime:  1 * time.Second,
			Cookie:    cookie,
			ParseFunc: doubangroup.ParseURL,
		})
	}

	var f collect.Fetcher = &collect.BrowserFetch{
		Timeout: 3000 * time.Millisecond,
		Logger:  logger,
		Proxy:   p,
	}

	s := engine.ScheduleEngine{
		WorkCount: 5,
		Logger:    logger,
		Fetcher:   f,
		Seeds:     seeds,
	}
	s.Run()
}
