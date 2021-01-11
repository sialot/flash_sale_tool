package script

import (
	"log"
	"io/ioutil"
)


// 标签类型
type TagType int

const (
	_ TagType = iota
	A
	Input
)

// 操作类型
type ActionType int

const (
	_ ActionType = iota
	WaitVisible
	ClickByID
)

// 定位方式
type LocateType int

const (
	_ ActionType = iota
	ByText
	ByID
)

// 动作
type Action struct {
	Tag      TagType
	Action   ActionType
	LocateBy int
	DomText  string
	DomValue string
	DomID    string
}

// 任务构造体
type Task struct {
	ID      string   // 处理ID
	Name    string   // 任务名称
	Actions []Action // 动作列表
}

func LoadScripts() {
	log.Println("加载抢单脚本  > ...")


	




	err := _getAllFile("./script/jsons/")
	if err != nil {
		log.Println(err.Error())
		panic(err)
	}
	log.Println("加载抢单脚本  > 结束")
}

func _getAllFile(pathname string) error {
    rd, err := ioutil.ReadDir(pathname)
    for _, fi := range rd {
        if fi.IsDir() {
            log.Println("忽略" + pathname+"\\"+fi.Name())
        } else {
            log.Println(fi.Name())
        }
    }
    return err
}

func _loadJson(path string) {

}
func _writeJson(path string, tsk Task){

}