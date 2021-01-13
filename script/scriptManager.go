package script

import (
	"log"
	"io/ioutil"
	"strconv"
	// "github.com/go-basic/uuid"
	jsoniter "github.com/json-iterator/go"
)

// 标签类型
type TagType string
const (
	_ TagType = ""
	A = "A"
	Input = "Input"
)

// 操作类型
type ActionType string
const (
	_ ActionType = ""
	WaitVisible = "WaitVisible"
	Click = "Click"
	SendKey = "SendKey"
)

// 定位方式
type LocateType string
const (
	_ ActionType = ""
	ByText = "ByText"
	ByID = "ByID"
)

// 动作
type Action struct {
	Action   ActionType
	Tag      TagType
	LocateBy string
	LocateParam string
	Param  string
}

// 任务构造体
type Task struct {
	ID      string   // 处理ID
	Name    string   // 任务名称
	Actions []Action // 动作列表
	Time string // 定时
}

var GlobalTasks []Task

func LoadScripts() {
	log.Println("SCRIPT MANAGER > ...")

	var pathName string = "./script/jsons/"

	rd, err := ioutil.ReadDir(pathName)
	if err != nil {
		log.Println(err.Error())
		log.Println("SCRIPT MANAGER > 失败！")
		panic(err)
	}

    for _, fi := range rd {
        if fi.IsDir() {
			continue
        } else {
			_loadJson(pathName + fi.Name())
        }
	}

	log.Println("SCRIPT MANAGER > 结束， 成功加载 " + strconv.Itoa(len(GlobalTasks)) + "个脚本" )
	log.Println("-------------------------------------------")
}

// 获取任务列表json字符串
func GetTaskListJson() (string, error) {
	
	bjson,err := jsoniter.Marshal(GlobalTasks)
	if err!=nil {
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

	GlobalTasks = append(GlobalTasks, task)

	log.Println("	成功加载文件：" + path)
	return nil
}

func _writeJson(path string, tsk Task){

}