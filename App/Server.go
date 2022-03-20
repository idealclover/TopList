package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"text/template"

	"github.com/tophubs/TopList/Common"
	"github.com/tophubs/TopList/Config"
)

func GetTypeInfo(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal("系统错误" + err.Error())
	}
	id := r.Form.Get("id")
	re := regexp.MustCompile("[0-9]+")
	id = re.FindString(id)
	sql := "select str from hotData2 where id=" + id
	data := Common.MySql{}.GetConn().ExecSql(sql)
	if len(data) == 0 {
		fmt.Fprintf(w, "%s", `{"Code":1,"Message":"id错误，无该分类数据","Data":[]}`)
		return
	}
	w.Header().Add("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, "%s", data[0]["str"])
}

func GetType(w http.ResponseWriter, r *http.Request) {
	res := Common.MySql{}.GetConn().Select("hotData2", []string{"name", "id"}).QueryAll()
	Common.Message{}.Success("获取数据成功", res, w)
}

func GetAvailableType(w http.ResponseWriter, r *http.Request) {
	input := `
{
	"data": [
		{
			"name": "知乎",
			"type": "zhihu",
			"id": "1",
			"appid": "wxeb39b10e39bf6b54",
			"re": "\\d+",
			"prepath": "zhihu/question?id=",
			"sufpath": "&source=recommend"
		},
		{
			"name": "微博",
			"type": "weibo",
			"id": "58",
			"appid": "wx9074de28009e1111",
			"re": "#(.+?)#",
			"prepath": "pages/topic/topic?topicContent=",
			"sufpath": ""
		}
	]
}
	`
	var res map[string]interface{}
	json.Unmarshal([]byte(input), &res)
	// },
	// {
	// 	"name": "微信",
	// 	"type": "weixin"
	// },
	// {
	// 	"name": "哔哩",
	// 	"type": "bilibili",
	// 	"appid": "wx7564fd5313d24844",
	// 	"re": "\\d+",
	// 	"prepath": "pages/video/video?avid=",
	// 	"sufpath": ""
	// },
	// {
	// 	"name": "v2ex",
	// 	"type": "v2ex",
	// 	"appid": "wx3f56c5b9471bde01",
	// 	"re": "\\d{2,}",
	// 	"prepath": "pages/Detail?id=",
	// 	"sufpath": ""
	Common.Message{}.Success("获取数据成功", res, w)
}

func GetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	fmt.Fprintf(w, "%s", Config.MySql().Source)
}

/**
kill -SIGUSR1 PID 可平滑重新读取mysql配置
*/
//func SyncMysqlCfg() {
//	s := make(chan os.Signal, 1)
//	signal.Notify(s, syscall.SIGUSR1)
//	go func() {
//		for {
//			<-s
//			Config.ReloadConfig()
//			log.Println("Reloaded config")
//		}
//	}()
//}

func main() {
	//SyncMysqlCfg()
	http.HandleFunc("/GetTypeInfo", GetTypeInfo)           // 设置访问的路由
	http.HandleFunc("/GetType", GetType)                   // 设置访问的路由
	http.HandleFunc("/GetAvailableType", GetAvailableType) // 设置访问的路由
	//http.HandleFunc("/GetConfig", GetConfig)      // 设置访问的路由

	// 静态资源
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("../Html/css/"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("../Html/js/"))))

	// 首页
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		t, err := template.ParseFiles("../Html/hot.html")
		if err != nil {
			log.Println("err")
		}
		t.Execute(res, nil)
	})

	err := http.ListenAndServe(":9090", nil) // 设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
