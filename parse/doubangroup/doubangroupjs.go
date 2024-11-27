package doubangroup

import (
	"go-crawler/collect"
)

var DoubangroupJSTask = &collect.TaskModle{
	Property: collect.Property{
		Name:     "js_find_douban_sun_room",
		WaitTime: 2,
		MaxDepth: 5,
		Cookie:   "bid=TuWqA7AM2_4; viewed=\"1500149_26883690\"; _pk_id.100001.8cb4=939420b529734caa.1731404400.; __yadk_uid=eaEj42WRDNHKvtY84a4E4ic8v9KGYszX; __utmc=30149280; dbcl2=\"150361748:Qt1XBWORkVA\"; ck=gzyI; push_noty_num=0; push_doumail_num=0; __utmv=30149280.15036; __utmz=30149280.1731789209.5.4.utmcsr=accounts.douban.com|utmccn=(referral)|utmcmd=referral|utmcct=/; _pk_ref.100001.8cb4=%5B%22%22%2C%22%22%2C1731929749%2C%22https%3A%2F%2Faccounts.douban.com%2F%22%5D; _pk_ses.100001.8cb4=1; __utma=30149280.1825385725.1728983070.1731918463.1731929749.11; __utmt=1; __utmb=30149280.14.5.1731929910346",
	},
	Root: `
    var arr = new Array();
     for (var i = 25; i <= 25; i+=25) {
      var obj = {
         Url: "https://www.douban.com/group/szsh/discussion?start=" + i,
         Priority: 1,
         RuleName: "解析网站URL",
         Method: "GET",
       };
      arr.push(obj);
    };
    console.log("console.log:"+arr[0].Url);
    AddJsReq(arr);
      `,
	Rules: []collect.RuleModle{
		{
			Name: "解析网站URL",
			ParseFunc: `
      ctx.ParseJSReg("解析阳台房","(https://www.douban.com/group/topic/[0-9a-z]+/)\"[^>]*>([^<]+)</a>");
      `,
		},
		{
			Name: "解析阳台房",
			ParseFunc: `
      //console.log("parse output");
      ctx.OutputJS("<div class=\"topic-content\">[\\s\\S]*?阳台[\\s\\S]*?<div class=\"aside\">");
      `,
		},
	},
}
