package main

import (
	"log"
	"net/http"
	config "./config"
	webdriver "./driver"
	server "./server"
	script "./script"
	"time"
)

// 构造方法
func init() {
	config.InitConfig()
}

// 入口
func main() {
	log.Println("==========================")
	log.Println("= 大福酱酱的抢单神器 v1.0 ")
	log.Println("==========================")
	log.Println("")

	// 启动本地API服务
	port, _ := config.SysConfig.Get("server.port")
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			resp, err := http.Get("http://localhost:" + port + "/")
			if err != nil {
				continue
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				continue
			}
			break
		}

		// 加载脚本
		script.LoadScripts()

		// 初始化webdriver
		err := webdriver.Init()
		if err != nil {
			panic(err)
		}

		log.Println("API SERVER > 启动成功!")
		log.Println("")
	}()

	server.StartServer(port)
}

