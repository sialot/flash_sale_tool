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
		resultJson = `refreshCallBack({code:-1, msg:'` + err.Error() + `'})`
		w.Write([]byte(resultJson))
		return
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
	webdriver.StopTask()
	responseStr = `{code:1}`
	w.Write([]byte(responseStr))
}