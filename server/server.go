package server

import (
	"net/http"
	"time"
	"os"
	"os/signal"
	"log"

	webdriver "../driver"
)

//
// Server Start
//
func StartServer(port string){

	log.Println("API SERVER > 启动中...")

	// 声明路由
	mux := http.NewServeMux()

	// UI
	mux.Handle("/", http.FileServer(http.Dir("static")))

	// 欢迎测试页
	mux.HandleFunc("/api/welcome", welcome)

	// 抓取上下文
	mux.HandleFunc("/api/geWsUrl", geWsUrl)

	// 点击按钮测试
	mux.HandleFunc("/api/clickBtnByText", clickBtnByText)
	

	// 启动服务
	svr := http.Server{
		Addr:         ":" + port,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 0,
		Handler:      mux,
	}

	// 监听退出信号
	quitChan := make(chan os.Signal)
	signal.Notify(quitChan, os.Interrupt, os.Kill)

	// 启动协程，等待信号
	go func() {
		<-quitChan
		svr.Close()
		log.Println("API SERVER > 已关闭!")
	}()

	svr.ListenAndServe()
}

//
// Router Start
//

// 欢迎
func welcome(w http.ResponseWriter, r *http.Request) {
	var resultJson string
	resultJson += "Welcome!"
	w.Write([]byte(resultJson))
}

// 抓取调试地址
func geWsUrl(w http.ResponseWriter, r *http.Request) {
	var responseStr string
	values := r.URL.Query()
	goodUrl :=values.Get("goodUrl")
	callback :=values.Get("callback")

	if goodUrl == "" {
		responseStr = `{code:-1}`

		if callback != "" {
			responseStr = callback + "(" + responseStr + ")"
		}
		w.Write([]byte(responseStr))
		return
	}

	log.Println("抓取调试地址 > " + goodUrl)
	wsUrl, err := webdriver.GetWsUrl(goodUrl)
	if err != nil {
		responseStr = `{code:-1, msg:'` + err.Error() + `'}`

		if callback != "" {
			responseStr = callback + "(" + responseStr + ")"
		}
		w.Write([]byte(responseStr))
		return
	}

	responseStr = `{code:1, wsUrl:'`+wsUrl+`'}`

	if callback != "" {
		responseStr = callback + "(" + responseStr + ")"
	}

	w.Write([]byte(responseStr))
}

// 测试按钮点击
func clickBtnByText(w http.ResponseWriter, r *http.Request) {


}