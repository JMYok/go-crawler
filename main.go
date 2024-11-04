package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// tag v0.0.3
func main() {
	url := "https://www.thepaper.cn/"
	resp, err := http.Get(url)

	if err != nil {
		fmt.Println("fetch url error:%v", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error status code:%v", resp.StatusCode)
	}

	// 读取为字节流
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("read content failed:%v\n", err)
		return
	}

	numLinks := strings.Count(string(body), "<a")
	fmt.Printf("homepage has %d links!\n", numLinks)

	isExist := strings.Contains(string(body), "AI")
	fmt.Printf("是否存在AI:%v\n", isExist)
}
