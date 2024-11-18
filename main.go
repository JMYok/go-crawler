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

	var f collect.Fetcher = &collect.BrowserFetch{
		Timeout: 3000 * time.Millisecond,
		Logger:  logger,
		Proxy:   p,
	}

	// douban cookie
	cookie := "bid=TuWqA7AM2_4; viewed=\"1500149_26883690\"; _pk_id.100001.8cb4=939420b529734caa.1731404400.; __yadk_uid=eaEj42WRDNHKvtY84a4E4ic8v9KGYszX; __utmc=30149280; dbcl2=\"150361748:Qt1XBWORkVA\"; ck=gzyI; push_noty_num=0; push_doumail_num=0; __utmv=30149280.15036; __utmz=30149280.1731789209.5.4.utmcsr=accounts.douban.com|utmccn=(referral)|utmcmd=referral|utmcct=/; _pk_ref.100001.8cb4=%5B%22%22%2C%22%22%2C1731929749%2C%22https%3A%2F%2Faccounts.douban.com%2F%22%5D; _pk_ses.100001.8cb4=1; __utma=30149280.1825385725.1728983070.1731918463.1731929749.11; __utmt=1; __utmb=30149280.14.5.1731929910346"
	var seeds []*collect.Task
	for i := 0; i <= 100; i += 25 {
		str := fmt.Sprintf("https://www.douban.com/group/szsh/discussion?start=%d", i)
		seeds = append(seeds, &collect.Task{
			Url:      str,
			WaitTime: 2 * time.Second,
			Cookie:   cookie,
			MaxDepth: 1,
			Fetcher:  f,
			RootReq: &collect.Request{
				Method:    "Get",
				ParseFunc: doubangroup.ParseURL,
			},
		})
	}

	s := engine.NewEngine(
		engine.WithFetcher(f),
		engine.WithLogger(logger),
		engine.WithWorkCount(3),
		engine.WithSeeds(seeds),
		engine.WithScheduler(engine.NewSchedule()),
	)
	s.Run()
}
