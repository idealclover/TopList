package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/tophubs/TopList/Common"
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
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", data[0]["str"])
}

func GetType(w http.ResponseWriter, r *http.Request) {
	res := Common.MySql{}.GetConn().Select("hotData2", []string{"name", "id"}).QueryAll()
	w.Header().Set("Content-Type", "application/json")
	Common.Message{}.Success("获取数据成功", res, w)
}

func GetAvailableType(w http.ResponseWriter, r *http.Request) {
	input := `
{
	"mode": "production",
	"data": [
		{
			"name": "知乎",
			"type": "zhihu",
			"id": "1",
			"appid": "wxeb39b10e39bf6b54",
			"re": "\\d+",
			"prepath": "zhihu/question?id=",
			"sufpath": "&source=recommend",
			"scheme": "zhihu",
			"scheme_prepath": "question/",
			"scheme_sufpath": ""
		},
		{
			"name": "微博",
			"type": "weibo",
			"id": "58",
			"appid": "wx9074de28009e1111",
			"re": "#(.+?)#",
			"prepath": "pages/topic/topic?topicContent=",
			"sufpath": ""
		},
		{
			"name": "哔哩",
			"type": "bilibili",
			"id": "5"
		},
	    {
			"name": "v2ex",
			"type": "v2ex",
			"id": "59",
			"appid": "wx3f56c5b9471bde01",
			"re": "\\d{2,}",
			"prepath": "pages/Detail?id=",
			"sufpath": ""
		},
	    {
			"name": "36Kr",
			"type": "36Kr",
			"id": "12"
		},
	    {
			"name": "虎嗅",
			"type": "HuXiu",
			"id": "8",
			"appid": "wxd1f72ce26251f419",
			"re": "\\d+",
			"prepath": "pages/article?aid=",
			"sufpath": ""
		},
		{
			"name": "微信",
			"type": "WeiXin",
			"id": "11"
		},
		{
			"name": "什么值得买",
			"type": "SMZDM",
			"id": "64"
		},
		{
			"name": "NGA",
			"type": "NGA",
			"id": "63"
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
	w.Header().Set("Content-Type", "application/json")
	Common.Message{}.Success("获取数据成功", res, w)
}

func GetMockAvailableType(w http.ResponseWriter, r *http.Request) {
	input := `
{
	"mode": "dev",
	"data": [
		{
			"name": "随想",
			"type": "SuiXiang",
			"id": "201"
		},
		{
			"name": "技术",
			"type": "JiShu",
			"id": "202"
		},
	    {
			"name": "生活",
			"type": "ShengHuo",
			"id": "203"
		},
	    {
			"name": "测评",
			"type": "CePing",
			"id": "204"
		}
	]
}
	`
	var res map[string]interface{}
	json.Unmarshal([]byte(input), &res)
	w.Header().Set("Content-Type", "application/json")
	Common.Message{}.Success("获取数据成功", res, w)
}

func main() {
	//SyncMysqlCfg()
	http.HandleFunc("/GetTypeInfo", GetTypeInfo)             // 设置访问的路由
	http.HandleFunc("/GetType", GetType)                     // 设置访问的路由
	http.HandleFunc("/GetAvailableType", GetAvailableType)   // 设置访问的路由
	http.HandleFunc("/GetAvailableType/1", GetAvailableType) // 设置访问的路由
	// http.HandleFunc("/GetAvailableType/2", GetMockAvailableType) // 设置访问的路由
	http.HandleFunc("/GetAvailableType/2", GetAvailableType) // 设置访问的路由
	//http.HandleFunc("/GetConfig", GetConfig)      // 设置访问的路由

	// 静态资源
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("../Html/css/"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("../Html/js/"))))

	// 首页
	//http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
	//	t, err := template.ParseFiles("../Html/hot.html")
	//	if err != nil {
	//		log.Println("err")
	//	}
	//	t.Execute(res, nil)
	//})

	err := http.ListenAndServe(":9090", nil) // 设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
