package script

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"path/filepath"
	"github.com/go-basic/uuid"
	jsoniter "github.com/json-iterator/go"
)

// 标签类型
type TagType string

const (
	_     TagType = ""
	A             = "A"
	Input         = "Input"
)

// 操作类型
type ActionType string

const (
	_           ActionType = ""
	WaitVisible            = "WaitVisible"
	Click                  = "Click"
	SendKey                = "SendKey"
)

// 定位方式
type LocateType string

const (
	_      ActionType = ""
	ByText            = "ByText"
	ByID              = "ByID"
)

// 动作
type Action struct {
	Action      ActionType
	Tag         TagType
	LocateBy    string
	LocateParam string
	Param       string
}

// 任务构造体
type Task struct {
	ID          string   // 处理ID
	Name        string   // 任务名称
	Actions     []Action // 动作列表
	Time        string   // 定时 yyyyMMddHHmmss
	DefaultTime string   // 默认时间点 HHmmss
}

var GlobalTasks []Task
var TempTasks []Task
var pathName string = "./script/jsons/"
func LoadScripts() {
	log.Println("SCRIPT MANAGER > ...")
	rd, err := ioutil.ReadDir(pathName)
	if err != nil {
		log.Println(err.Error())
		log.Println("SCRIPT MANAGER > 失败！")
		panic(err)
	}
	TempTasks = TempTasks[0:0]

	for _, fi := range rd {
		if fi.IsDir() {
			continue
		} else {
			_loadJson(pathName + fi.Name())
		}
	}
	GlobalTasks = TempTasks
	log.Println("SCRIPT MANAGER > 结束， 成功加载 " + strconv.Itoa(len(GlobalTasks)) + "个脚本")
	log.Println("-------------------------------------------")
}

// 获取任务列表json字符串
func GetTaskListJson() (string, error) {

	bjson, err := jsoniter.Marshal(GlobalTasks)
	if err != nil {
		return "", err
	}

	return string(bjson), nil
}

// 解析json
func ReadJson(json string) (Task, error) {
	var task Task

	err := jsoniter.Unmarshal([]byte(json), &task)
	if err != nil {
		log.Println("json解析出错!")
		return task, err
	}
	return task, nil
}

func _loadJson(path string) error {
	var task Task
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("SCRIPT MANAGER > ...")
		log.Println("	加载文件：" + path + "读取出错!")
		return err
	}
	err = jsoniter.Unmarshal(b, &task)
	if err != nil {
		log.Println("	加载文件：" + path + "json解析出错!")
		return err
	}

	TempTasks = append(TempTasks, task)

	log.Println("	成功加载文件：" + path)
	return nil
}

func SaveJson(taskJson string) error{

	// 解析json
	var task Task
	var err error
	task, err = ReadJson(taskJson)
	if err != nil {
		log.Println(err)
		return err
	}
	
	err = _writeJson(pathName, task)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// 判断文件或文件夹是否存在
// isPathExist
func isPathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func _writeJson(path string, tsk Task) error{

	uuid := uuid.New()
	tsk.ID = uuid

	jsonByte, err := jsoniter.MarshalIndent(tsk, "" ,"")
	if err != nil{
       return err
	}

	absPath, _ := filepath.Abs(path + uuid + ".json")
	log.Println("absPath > " + absPath)

	exist, err := isPathExist(absPath)
	if err != nil {
		log.Println(err)
		return err
	} else {

		if !exist {

			file, err := os.Create(absPath)
			defer file.Close()
			if err != nil {
				log.Println(err)
				return err
			}
		}
	}

	file, err := os.OpenFile(absPath, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("file > ")
	_, err = file.Write(jsonByte)
	if err != nil {
		log.Println(err)
		return err
	}

	LoadScripts()
	return nil
}
