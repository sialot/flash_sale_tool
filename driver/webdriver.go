package webdriver
import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

	chromePath, _:= config.SysConfig.Get("chrome.path")
	remoteDebugPort, _ = config.SysConfig.Get("chrome.remote_debugging_port")
	port, _ := config.SysConfig.Get("server.port")

	log.Println("CHROME 浏览器 > 启动中...")
	log.Println("       可执行文件位置：" + chromePath)
	log.Println("       远程调试端口：" + remoteDebugPort)

	// 拼接启动命令
	// cmd.exe /c start "" "C:\Program Files (x86)\Google\Chrome\Application\chrome.exe" --new-window --remote-debugging-port=9222 http://localhost:9222/json
	cmd := exec.Command("cmd.exe", "/c", "start","", chromePath, "--remote-debugging-port=" + remoteDebugPort, "http://localhost:" + port)
	err := cmd.Run()
	if err != nil {

		// 命令执行失败
		log.Println("CHROME 浏览器 > 启动失败！")
		return err
	}

	log.Println("CHROME 浏览器 > 启动成功！")
	log.Println("")
	return nil
}

func GetWsUrl(goodUrl string) (string, error){

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
				
				if pageArr[i].Url == goodUrl {
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

		return remoteDebugUrl, nil
	} else {
		err := errors.New("远程调试地址 > 获取失败！请确认网址输入正确")
		log.Println("CHROME 远程调试地址 > 获取失败！")
		log.Println("       请确认网址输入正确")
		return "", err
	}
}

// // 跳转大福酱酱的抢单神器
// func ShowWebUI(port string)  {

// 	allocCtx, _ := chromedp.NewRemoteAllocator(context.Background(), remoteDebugUrl)
	
// 	// also set up a custom logger
// 	taskCtx, _ := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

// 	chromedp.Run(taskCtx,
// 		chromedp.Navigate(`http://localhost:` + port), 
// 	)
// }

// 打开网页
func OpenPage(url string)  error{

	if globalAllocCtx == nil {
		log.Println("NewRemoteAllocator")
		globalAllocCtx, _ = chromedp.NewRemoteAllocator(context.Background(), remoteDebugUrl)
	}

	if globalTaskCtx == nil {
		log.Println("NewContext")
		globalTaskCtx, _ = chromedp.NewContext(globalAllocCtx, chromedp.WithLogf(log.Printf))
	}

	err := chromedp.Run(globalTaskCtx,	chromedp.Navigate(url))
	if err!= nil {
		log.Println(err)
		return err
	}
	return nil
}

// 点击按钮


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
