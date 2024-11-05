package main

import (
	"fmt"
	"go-crawler/collect"
)

func main() {
	url := "https://m.damai.cn/shows/item.html?from=def&itemId=838630926963&sqm=dianying.h5.unknown.value&spm=a2o71.category_singconcert.floor1.item_3"
	var f collect.Fetcher = collect.BaseFetch{}
	body, err := f.Get(url)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
}
