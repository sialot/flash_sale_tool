package webdriver

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"time"
	"sync"
	//	"github.com/chromedp/cdproto/cdp"
	"errors"
	"os/exec"

	config "../config"
	script "../script"
	"github.com/chromedp/chromedp"
	jsoniter "github.com/json-iterator/go"
)

// 远程调试端口
var RemoteDebugPort string = ""

// 远程调试地址
var RemoteDebugUrl string = ""

// 全局上下文
var GlobalTaskCtx context.Context = nil

// 停止标志
var StopSignal bool = false

// 操作锁
var mutex sync.Mutex
var taskProcessFlag bool = false

// 第一次打开
var firstOpen = true

// chrome 调试信息对象
type Page struct {
	Description          string
	DevtoolsFrontendUrl  string
	Id                   string
	Title                string
	PageType             string `json:"type"`
	Url                  string
	WebSocketDebuggerUrl string
}

// 初始化
func Init() error {
	var err error

	// 启动浏览器
	err = _startChrome()
	if err != nil {
		return err
	}

	// 刷新上下文
	err = InitContext()
	if err != nil {
		return err
	}

	return nil
}

// 启动浏览器
func _startChrome() error {

	log.Println("CHROME > 浏览器启动中...")
	var err error

	// 抢单UI网址
	port, err := config.SysConfig.Get("server.port")
	if err != nil {
		log.Println("	配置加载失败\"chrome.remote_debugging_port\"")
		return err
	}
	UIUrl := "http://localhost:" + port + "/"

	// 获取远程调试端口
	RemoteDebugPort, err = config.SysConfig.Get("chrome.remote_debugging_port")
	if err != nil {
		log.Println("	配置加载失败\"chrome.remote_debugging_port\"")
		return err
	}

	log.Println("	远程调试端口：" + RemoteDebugPort)

	// 判断当前操作系统
	switch os := runtime.GOOS; os {

	// OS X
	case "darwin":
		log.Println("	当前系统：mac os x")

		// 拼接启动命令
		// /usr/bin/open -a Google\ Chrome --args --remote-debugging-port=9222
		cmd := exec.Command("/usr/bin/open", "-a", "Google Chrome", "--args", "--remote-debugging-port="+RemoteDebugPort, UIUrl)
		err = cmd.Run()
		if err != nil {

			// 命令执行失败
			log.Println("	CHROME > 启动失败！")
			return err
		}

	// Windows
	case "windows":
		log.Println("	当前系统：windows")
		chromePath, err := config.SysConfig.Get("chrome.path.windows")
		if err != nil {
			log.Println("	配置加载失败\"chrome.path.windows\"")
			return err
		}
		log.Println("	可执行文件位置：" + chromePath)

		// 拼接启动命令
		// cmd.exe /c start "" "C:\Program Files (x86)\Google\Chrome\Application\chrome.exe" --new-window --remote-debugging-port=9222 http://localhost:9222/json
		cmd := exec.Command("cmd.exe", "/c", "start", "", chromePath, "--remote-debugging-port="+RemoteDebugPort, UIUrl)
		err = cmd.Run()
		if err != nil {

			// 命令执行失败
			log.Println("	启动失败！")
			return err
		}

	default:
		fmt.Println("	不支持当前操作系统")
		err := errors.New("不支持当前操作系统")
		return err
	}

	log.Println("CHROME > 启动成功！")
	
	return nil
}

// 初始化上下文
func InitContext() error {

	log.Println("INIT CONTEXT > 获取中...")

	// 抢单中无法操作
	if _getTaskProcessFlag() {
		log.Println("当前已有任务")
		var err = errors.New("当前已有任务执行！")
		return err
	}

	log.Println("	请求地址：" + "http://localhost:" + RemoteDebugPort + "/json")

	RemoteDebugUrl = ""

	// 抓取json数据
	resp, err := http.Get("http://localhost:" + RemoteDebugPort + "/json")
	if err != nil {

		// 命令执行失败
		log.Println("INIT CONTEXT > 获取失败！ 请关闭所有正在运行的chrome浏览器,然后重新启动秒杀神器！")
		return err
	} 
	
	defer resp.Body.Close()

	// 返回成功
	if resp.StatusCode == http.StatusOK {

		jsonStr, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("INIT CONTEXT > 获取失败！ 请关闭所有正在运行的chrome浏览器,然后重新启动秒杀神器！")
			return err
		}

		// 解析json
		var pageArr []Page
		err = jsoniter.Unmarshal(jsonStr, &pageArr)
		if err != nil {
			return err
		}

		if len(pageArr) > 0 {
			RemoteDebugUrl = pageArr[0].WebSocketDebuggerUrl
			log.Println("	调试地址：" + RemoteDebugUrl)
		} else {
			var err = errors.New("INIT CONTEXT > 获取失败")
			log.Println("INIT CONTEXT > 获取失败！ 请关闭所有正在运行的chrome浏览器,然后重新启动秒杀神器！")
			return err
		}
	}
	allocCtx, _ := chromedp.NewRemoteAllocator(context.Background(), RemoteDebugUrl)
	GlobalTaskCtx, _ = chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	firstOpen = true
	log.Println("INIT CONTEXT >获取成功！")
	
	return nil
}

// 打开商品页
func OpenPage(goodUrl string) error {

	log.Println("打开商品页 > " + goodUrl)
	
	// 抢单中无法操作
	if _getTaskProcessFlag() {
		log.Println("当前已有任务")
		var err = errors.New("当前已有任务执行！")
		return err
	}

	var res string
	err := chromedp.Run(GlobalTaskCtx,
		chromedp.Navigate(goodUrl),
		chromedp.Evaluate(`document.title = '【秒杀神器】' + document.title`, &res),
	)
	if err != nil {
		log.Println(err)
		log.Println("打开商品页 > 失败")
		return err
	}
	firstOpen = false
	log.Println("打开商品页 > 成功")
	
	return nil
}

// 终止任务
func StopTask() {
	mutex.Lock()
	defer mutex.Unlock()
	StopSignal = true
}

// 设置执行标识
func _setTaskProcessFlag(flag bool){
	mutex.Lock()
	defer mutex.Unlock()
	taskProcessFlag = flag
}

// 获取执行标识
func _getTaskProcessFlag() bool {
	mutex.Lock()
	defer mutex.Unlock()
	return taskProcessFlag
}

// 执行任务
func ExecTask(taskJson string) error {

	log.Println("EXECTASK > " + taskJson)
	if firstOpen {
		var err = errors.New("尚未打开秒杀商品页面！")
		return err
	}

	// 防止连点
	if _getTaskProcessFlag() {
		log.Println("EXECTASK > 当前已有任务")
		var err = errors.New("当前已有任务执行！")
		return err
	} else {
		_setTaskProcessFlag(true)
	}

	// 解析json
	var task script.Task
	var err error
	task, err = script.ReadJson(taskJson)
	if err != nil {
		log.Println(err)
		_setTaskProcessFlag(false)
		return err
	}

	// 解析处理时间 yyyyMMddHHmmss
	format := "20060102150405"
	var timeLayoutStr = "2006-01-02 15:04:05"

	//targetTime, err := time.Parse(format, task.Time)
	targetTime, _ := time.ParseInLocation(format, task.Time, time.Local)

	if err != nil {
		errTime := errors.New("非法时间表达式！")
		_setTaskProcessFlag(false)
		return errTime
	}
	log.Println("目标时间：" + targetTime.Format(timeLayoutStr))

	StopSignal = false

	ticker := time.NewTicker(time.Millisecond * 1)
	log.Println("开始计时...")

	var lastPrintStr string = ""
	go func() {
		for {
			<-ticker.C

			// 取消
			if StopSignal {
				fmt.Printf("\n")
				log.Println("取消")
				ticker.Stop()
				_setTaskProcessFlag(false)
				break
			}

			// 当前时间
			now := time.Now()

			// 打印等待状态
			var printStr = now.Format(timeLayoutStr)
			if lastPrintStr != printStr {
				lastPrintStr = printStr
				fmt.Printf("\r	[等待中] 当前时间：%s，目标时间：%s",  printStr, targetTime.Format(timeLayoutStr))

				//根据返回类型定义res
				var res string
				ctx, cancel := context.WithTimeout(GlobalTaskCtx, 100*time.Millisecond)
				defer cancel()
				_ = chromedp.Run(ctx, chromedp.Tasks{
					chromedp.Evaluate(`document.title = '抢单中...[` + now.Format("15:04:05") + `]'`, &res),
				})
			}

			// 到达预定时间
			if now.After(targetTime) {
				fmt.Printf("\n")
				ticker.Stop() //停止定时器
				err = _runScript(task)
				_setTaskProcessFlag(false)
				break
			}
		}
	}()

	return nil
}

// 脚本执行
func _runScript(task script.Task) error {

	log.Println("执行脚本开始！")

	// 遍历处理Action
	var i int
	for i = 0; i < len(task.Actions); i++ {

		// 获得action
		var actionInfo = task.Actions[i]
	
		ac, err := _getChromedpAction(actionInfo)
		if err != nil {
			log.Println(err)
			continue
		}

		// 超时
		var timeout time.Duration

		// 等待处理
		if actionInfo.Action == script.WaitVisible {

			// 暴力执行每个步骤
			for {
				if StopSignal {
					log.Println("取消操作")
					break
				}

				// 输入需要更多时间
				if task.Actions[i].Action == script.SendKey {
					timeout =  5000*time.Millisecond
				} else {
					timeout =  10*time.Millisecond
				}
				ctx, cancel := context.WithTimeout(GlobalTaskCtx, timeout)
				err = chromedp.Run(ctx, ac)
				cancel()
				if err != nil {
					continue
				} else {
					break
				}
			}
		} else {
			if StopSignal {
				log.Println("取消操作")
				break
			}

			timeout =  5000*time.Millisecond
			ctx, cancel := context.WithTimeout(GlobalTaskCtx, timeout)

			err = chromedp.Run(ctx, ac)
			cancel()
			if err != nil {
				log.Println(err)
				log.Println("步骤失败")
				return err
			}
		}

		log.Println("完成")
	}

	log.Println("执行脚本完成！")
	return nil
}

// 拼装动作
func _getChromedpAction(action script.Action) (chromedp.Action, error) {

	// 选择表达式
	var sel string
	switch action.LocateBy {
	case script.ByText:

		switch action.Tag {
		case script.A:
			sel = fmt.Sprintf(`//a[text()[contains(., '%s')]]`, action.LocateParam)

		case script.Input:
			sel = fmt.Sprintf(`//input[@value='%s']`, action.LocateParam)
		}

	case script.ByID:
		switch action.Tag {
		case script.A:
			sel = fmt.Sprintf(`//a[@id='%s']`, action.LocateParam)

		case script.Input:
			sel = fmt.Sprintf(`//input[@id='%s']`, action.LocateParam)
		}
	}

	switch action.Action {
	case script.WaitVisible:
		log.Println("WaitVisible:" + sel)
		return chromedp.WaitVisible(sel), nil
	case script.Click:
		log.Println("Click:" + sel)
		return chromedp.Click(sel), nil
	case script.SendKey:
		log.Println("SendKeys:" + sel + "  > " + action.Param)
		return chromedp.SendKeys(sel, action.Param), nil
	}
	var err = errors.New("未知动作")
	return nil, err
}