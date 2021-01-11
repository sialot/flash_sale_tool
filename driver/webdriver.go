package webdriver
import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"fmt"
//	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"os/exec"
	"errors"
	jsoniter "github.com/json-iterator/go"
	config "../config"
)

var remoteDebugPort string
var remoteDebugUrl string
var globalAllocCtx context.Context = nil
var globalTaskCtx context.Context = nil

// chrome 调试信息对象
type Page struct {
    Description  string
    DevtoolsFrontendUrl   string
    Id string
    Title   string
    PageType string `json:"type"`
	Url  string
	WebSocketDebuggerUrl string
}

// 初始化
func Init() error{
	var err error
	
	// 启动浏览器
	err = _startChrome()
	if err != nil {
		return err
	}
	return nil
}

// 启动浏览器
func _startChrome() error{

	log.Println("CHROME 浏览器 > 启动中...")
	var err error

	// 抢单UI网址
	port, err := config.SysConfig.Get("server.port")
	if err != nil {
		log.Println("配置加载失败\"chrome.remote_debugging_port\"")
		return err
	}
	UIUrl := "http://localhost:" + port + "/"

	// 获取远程调试端口
	remoteDebugPort, err = config.SysConfig.Get("chrome.remote_debugging_port")
	if err != nil {
		log.Println("配置加载失败\"chrome.remote_debugging_port\"")
		return err
	}

	log.Println("       远程调试端口：" + remoteDebugPort)

	// 判断当前操作系统
	switch os := runtime.GOOS; os {

	// OS X	
    case "darwin":
		log.Println("       当前系统：mac os x")
		chromePath, err:= config.SysConfig.Get("chrome.path.macos")
		if err != nil {
			log.Println("配置加载失败\"chrome.path.macos\"")
			return err
		}

		log.Println("       可执行文件位置：" + chromePath)

		// 拼接启动命令
		// /Applications/Google\ Chrome.app/Contents/MacOS/Google\ Chrome  ~/Desktop/test.mp4 --remote-debugging-port=9222 http://localhost:9222/json
		cmd := exec.Command( chromePath, "--remote-debugging-port=" + remoteDebugPort, UIUrl)
		err = cmd.Run()
		if err != nil {
	
			// 命令执行失败
			log.Println("CHROME 浏览器 > 启动失败！")
			return err
		}

	// Windows
	case "windows":
		log.Println("       当前系统：windows")
		chromePath, err:= config.SysConfig.Get("chrome.path.windows")
		if err != nil {
			log.Println("配置加载失败\"chrome.path.windows\"")
			return err
		}
		log.Println("       可执行文件位置：" + chromePath)
		
		// 拼接启动命令
		// cmd.exe /c start "" "C:\Program Files (x86)\Google\Chrome\Application\chrome.exe" --new-window --remote-debugging-port=9222 http://localhost:9222/json
		cmd := exec.Command("cmd.exe", "/c", "start","", chromePath, "--remote-debugging-port=" + remoteDebugPort, UIUrl)
		err = cmd.Run()
		if err != nil {
	
			// 命令执行失败
			log.Println("CHROME 浏览器 > 启动失败！")
			return err
		}

    default:
        fmt.Println("不支持当前操作系统")
		err := errors.New("不支持当前操作系统")
		return err
    }

	log.Println("CHROME 浏览器 > 启动成功！")
	log.Println("")
	return nil
}

// 打开商品页
func OpenPage(goodUrl string) (string, error){

	log.Println("CHROME 远程调试地址 > 获取中...")
	log.Println("       请求地址：" + "http://localhost:" + remoteDebugPort + "/json")
	log.Println("       商品页地址：" + goodUrl)

	remoteDebugUrl = ""

	// 抓取json数据
	resp, err := http.Get("http://localhost:" + remoteDebugPort + "/json")
	if err != nil {

		// 命令执行失败
		log.Println("CHROME 远程调试地址 > 获取失败！")
		log.Println("       请关闭所有正在运行的chrome浏览器,然后重新启动秒杀神器！")
		return "", err
	}
	defer resp.Body.Close()

	// 返回成功
	if resp.StatusCode == http.StatusOK {

		robots, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("CHROME 远程调试地址 > 获取失败！")
			log.Println("       请关闭所有正在运行的chrome浏览器,然后重新启动秒杀神器！")
			return "", err
		}

		// 解析json
		var pageArr []Page
		err = jsoniter.Unmarshal(robots, &pageArr)
		if err != nil {
			return "", err
		}

		if len(pageArr) > 0 {

			// 遍历找到调试信息
			for i:=0;i<len(pageArr);i++ {
				
				if pageArr[i].Url == UIUrl {
					remoteDebugUrl = pageArr[i].WebSocketDebuggerUrl
					break
				}
			}
		}
	}

	if remoteDebugUrl != "" {
		log.Println("       调试地址：" + remoteDebugUrl)
		log.Println("CHROME 远程调试地址 > 获取成功！")

		log.Println("刷新远程调试上下文 > 开始")
		globalAllocCtx, _ = chromedp.NewRemoteAllocator(context.Background(), remoteDebugUrl)
		globalTaskCtx, _ = chromedp.NewContext(globalAllocCtx, chromedp.WithLogf(log.Printf))
		log.Println("刷新远程调试上下文 > 结束")

		log.Println("打开商品页 > 开始")
		err := chromedp.Run(globalTaskCtx,
			chromedp.Navigate(goodUrl),	
		)
		if err!= nil {
			log.Println(err)
			log.Println("打开商品页 > 失败")
			return "",err
		}
		log.Println("打开商品页 > 结束")

		return remoteDebugUrl, nil
	} else {
		err := errors.New("远程调试地址 > 获取失败！请确认网址输入正确")
		log.Println("CHROME 远程调试地址 > 获取失败！")
		log.Println("       请确认网址输入正确")
		return "", err
	}
}

// 淘宝自动秒杀 TEST DEMO
func AutoBuyTaobaoV1(buyText string, orderText string, pwText string, payText string)  error{

	log.Println("自动购买  > 购买按钮：" + buyText)
	log.Println("           提交按钮：" + orderText)
	log.Println("           支付密码：" + pwText)
	log.Println("           付款按钮：" + payText)

	// 拼接 xpath 表达式，搜索包含指定文本的a标签
	buySel := fmt.Sprintf(`//a[text()[contains(., '%s')]]`, buyText)
	orderSel := fmt.Sprintf(`//a[text()[contains(., '%s')]]`, orderText)
	paySel := fmt.Sprintf(`//input[@value='%s']`, payText)

	err := chromedp.Run(globalTaskCtx,
		chromedp.WaitVisible(buySel),
		chromedp.Click(buySel),
		chromedp.WaitVisible(orderSel),
		chromedp.Click(orderSel),
		chromedp.WaitVisible(`input[id=payPassword_rsainput]`),
		chromedp.SendKeys(`input[id=payPassword_rsainput]`, pwText),
		chromedp.WaitVisible(paySel),
		chromedp.Click(paySel),
	)
	if err!= nil {
		log.Println(err)
		return err
	}

	log.Println("自动购买  > 完成")
	return nil
}

// 按包含文本点击按钮 TEST DEMO
func ClickBtnByText(text string)  error{

	log.Println("按包含文本搜索并点击A标签  > 文本：" + text)

	// 拼接 xpath 表达式，搜索包含指定文本的a标签
	sel := fmt.Sprintf(`//a[text()[contains(., '%s')]]`, text)

	err := chromedp.Run(globalTaskCtx,	
		chromedp.Click(sel),
	)
	if err!= nil {
		log.Println(err)
		return err
	}

	log.Println("按包含文本搜索并点击A标签  > 成功")
	return nil
}

func Demo()  {

	dir, err := ioutil.TempDir("", "chromedp-example")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	// 使用普通模式打开浏览器
 	// options := []chromedp.ExecAllocatorOption{
    //     chromedp.Flag("headless", false),
    //     chromedp.Flag("hide-scrollbars", false),
    //     chromedp.Flag("mute-audio", false),
    //     chromedp.UserAgent(`Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36`),
    // }
 
	// options = append(chromedp.DefaultExecAllocatorOptions[:],options...)
	
//	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)

	allocCtx, _ := chromedp.NewRemoteAllocator(context.Background(), "ws://localhost:9021/devtools/page/68036593BF20DDADACFD11E584ACA592")

	//defer cancel()
	
	// also set up a custom logger
	taskCtx, _ := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	//defer cancel()

	chromedp.Run(taskCtx,
		//打开网页
		chromedp.Navigate(`http://www.baidu.com`), 

		//chromedp.Sleep(100*time.Second),

		)
	
		
}
