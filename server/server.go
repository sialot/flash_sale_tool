package server

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	webdriver "../driver"
	script "../script"
)

//
// Server Start
//
func StartServer(port string) {

	log.Println("API SERVER > 启动中...")

	// 声明路由
	mux := http.NewServeMux()

	// UI
	mux.Handle("/", http.FileServer(http.Dir("static")))

	// 欢迎测试页
	mux.HandleFunc("/api/welcome", welcome)

	// 获取任务列表
	mux.HandleFunc("/api/getTaskList", getTaskList)

	// 刷新上下文
	mux.HandleFunc("/api/refreshCtx", refreshCtx)
	
	// 打开商品页
	mux.HandleFunc("/api/openPage", openPage)

	// 打开商品页
	mux.HandleFunc("/api/execTask", execTask)
	mux.HandleFunc("/api/cancelExec", cancelExec)
	
	// 自动秒杀测试
	mux.HandleFunc("/api/autoBuyTest", autoBuyTest)

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

	taskListJson, err := script.GetTaskListJson()
	if err != nil {
		taskListJson = "[]"
	}

	resultJson += "initCallBack({code:1, wsUrl:'" + webdriver.RemoteDebugUrl + 
		"',data:" + taskListJson + "})"
	w.Write([]byte(resultJson))
}

func refreshCtx(w http.ResponseWriter, r *http.Request) {
	var resultJson string

	err := webdriver.InitContext()
	if err != nil {
		resultJson = `{code:-1}`
	}

	resultJson += "refreshCallBack({code:1, wsUrl:'" + webdriver.RemoteDebugUrl + "'})"
	w.Write([]byte(resultJson))
}

// 获取任务列表
func getTaskList(w http.ResponseWriter, r *http.Request) {
	var responseStr string
	values := r.URL.Query()
	callback := values.Get("callback")

	taskListJson, err := script.GetTaskListJson()
	if err != nil {
		responseStr = `{code:-1, msg:'` + err.Error() + `'}`

		if callback != "" {
			responseStr = callback + "(" + responseStr + ")"
		}
		w.Write([]byte(responseStr))
		return
	}
	
	responseStr = `{code:1,data:` + taskListJson + `}`
	if callback != "" {
		responseStr = callback + "(" + responseStr + ")"
	}

	w.Write([]byte(responseStr))
}

// 打开商品页
func openPage(w http.ResponseWriter, r *http.Request) {
	var responseStr string
	values := r.URL.Query()
	goodUrl := values.Get("goodUrl")
	callback := values.Get("callback")

	if goodUrl == "" {
		responseStr = `{code:-1}`

		if callback != "" {
			responseStr = callback + "(" + responseStr + ")"
		}
		w.Write([]byte(responseStr))
		return
	}

	log.Println("OPEN_PAGE > " + goodUrl)
	err := webdriver.OpenPage(goodUrl)
	if err != nil {
		responseStr = `{code:-1, msg:'` + err.Error() + `'}`

		if callback != "" {
			responseStr = callback + "(" + responseStr + ")"
		}
		w.Write([]byte(responseStr))
		return
	}

	responseStr = `{code:1}`

	if callback != "" {
		responseStr = callback + "(" + responseStr + ")"
	}

	w.Write([]byte(responseStr))
}

// 自定义脚本
func execTask(w http.ResponseWriter, r *http.Request) {

	var responseStr string
	values := r.URL.Query()
	taskJson := values.Get("taskJson")
	callback := values.Get("callback")

	if taskJson == "" {
		responseStr = `{code:-1}`

		if callback != "" {
			responseStr = callback + "(" + responseStr + ")"
		}
		w.Write([]byte(responseStr))
		return
	}

	log.Println("EXECTASK > " + taskJson)

	err := webdriver.ExecTask(taskJson)

	if err != nil {
		responseStr = `{code:-1, msg:'` + err.Error() + `'}`

		if callback != "" {
			responseStr = callback + "(" + responseStr + ")"
		}
		w.Write([]byte(responseStr))
		return
	}

	responseStr = `{code:1}`

	if callback != "" {
		responseStr = callback + "(" + responseStr + ")"
	}

	w.Write([]byte(responseStr))
}

// 取消
func cancelExec(w http.ResponseWriter, r *http.Request) {
	var responseStr string
	webdriver.StopSignal = true
	responseStr = `{code:1}`
	w.Write([]byte(responseStr))
}

// 自动购买测试 DEMO
func autoBuyTest(w http.ResponseWriter, r *http.Request) {

	// 返回值
	var responseStr string

	values := r.URL.Query()
	buyText := values.Get("buyText")
	orderText := values.Get("orderText")
	pwText := values.Get("pwText")
	payText := values.Get("payText")
	callback := values.Get("callback")

	if buyText == "" || orderText == "" || pwText == "" || payText == "" {
		responseStr = `{code:-1}`

		if callback != "" {
			responseStr = callback + "(" + responseStr + ")"
		}
		w.Write([]byte(responseStr))
		return
	}

	log.Println("自动购买测试")
	err := webdriver.AutoBuyTaobaoV2(buyText, orderText, pwText, payText)
	if err != nil {
		responseStr = `{code:-1, msg:'` + err.Error() + `'}`

		if callback != "" {
			responseStr = callback + "(" + responseStr + ")"
		}
		w.Write([]byte(responseStr))
		return
	}

	responseStr = `{code:1}`

	if callback != "" {
		responseStr = callback + "(" + responseStr + ")"
	}

	w.Write([]byte(responseStr))
}

// 按名称点击按钮 DEMO
func clickBtnByText(w http.ResponseWriter, r *http.Request) {

	// 返回值
	var responseStr string

	values := r.URL.Query()
	text := values.Get("text")
	callback := values.Get("callback")

	if text == "" {
		responseStr = `{code:-1}`

		if callback != "" {
			responseStr = callback + "(" + responseStr + ")"
		}
		w.Write([]byte(responseStr))
		return
	}

	log.Println("按包含文本点击 a 标签 > " + text)
	err := webdriver.ClickBtnByText(text)
	if err != nil {
		responseStr = `{code:-1, msg:'` + err.Error() + `'}`

		if callback != "" {
			responseStr = callback + "(" + responseStr + ")"
		}
		w.Write([]byte(responseStr))
		return
	}

	responseStr = `{code:1}`

	if callback != "" {
		responseStr = callback + "(" + responseStr + ")"
	}

	w.Write([]byte(responseStr))
}
