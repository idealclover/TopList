package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tophubs/TopList/Common"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	uUrl "net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bitly/go-simplejson"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type HotData struct {
	Code    int
	Message string
	Data    interface{}
}

type Spider struct {
	DataType string
}

func SaveDataToJson(data interface{}) string {
	Message := HotData{}
	Message.Code = 0
	Message.Message = "获取成功"
	Message.Data = data
	jsonStr, err := json.Marshal(Message)
	if err != nil {
		log.Fatal("序列化json错误")
	}
	return string(jsonStr)

}

// GetZhiHu 知乎
func (spider Spider) GetZhiHu() []map[string]interface{} {
	// 基础设置
	timeout := 5 * time.Second //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	// 请求网络
	url := "https://www.zhihu.com/api/v3/feed/topstory/hot-lists/total"
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	res, err := client.Do(request)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//解析数据
	str, _ := ioutil.ReadAll(res.Body)
	js, err := simplejson.NewJson(str)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	// 装填数据
	var allData []map[string]interface{}
	lens := len(js.Get("data").MustArray())
	for i := 0; i < lens; i++ {
		item := js.Get("data").GetIndex(i).MustMap()
		id := item["target"].(map[string]interface{})["id"].(json.Number).String()
		title := item["target"].(map[string]interface{})["title"]
		rUrl := "https://www.zhihu.com/question/" + item["target"].(map[string]interface{})["id"].(json.Number).String()
		desc := item["detail_text"]
		pic := item["children"].([]interface{})[0].(map[string]interface{})["thumbnail"]
		allData = append(allData, map[string]interface{}{
			"id":    id,
			"title": title,
			"url":   rUrl,
			"desc":  desc,
			"pic":   pic})
	}
	return allData
}

// GetWeiBo 微博
func (spider Spider) GetWeiBo() []map[string]interface{} {
	// 基础设置
	timeout := 5 * time.Second //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	// 请求网络
	url := "https://weibo.com/ajax/statuses/hot_band"
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	res, err := client.Do(request)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//解析数据
	str, _ := ioutil.ReadAll(res.Body)
	js, err := simplejson.NewJson(str)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	// 装填数据
	var allData []map[string]interface{}
	lens := len(js.Get("data").Get("band_list").MustArray())
	for i := 0; i < lens; i++ {
		item := js.Get("data").Get("band_list").GetIndex(i).MustMap()
		// 先判断是不是广告，是就可以直接跳过
		isAd, ok := item["is_ad"].(json.Number)
		if !ok {
			isAd = "0"
		}
		if isAd.String() != "0" {
			i++
			continue
		}
		// TODO: 这里写的很trick
		id := item["mid"]
		title := item["note"]
		rUrl := "https://s.weibo.com/weibo?q=" + uUrl.QueryEscape(item["word_scheme"].(string))
		desc := string(item["raw_hot"].(json.Number)) + " 热度"
		var pic = ""
		mblog, ok := item["mblog"].(map[string]interface{})
		if !ok {
			mblog = map[string]interface{}{}
		}
		pics, ok := mblog["pic_ids"].([]interface{})
		if !ok {
			pics = []interface{}{}
		}
		if len(pics) > 0 {
			pic = "https://wx1.sinaimg.cn/orj360/" + pics[0].(string) + ".jpg"
		}
		allData = append(allData, map[string]interface{}{
			"id":    id,
			"title": title,
			"url":   rUrl,
			"desc":  desc,
			"pic":   pic})
	}
	return allData
}

// GetBili 哔哩哔哩
func (spider Spider) GetBili() []map[string]interface{} {
	// 基础设置
	timeout := 5 * time.Second //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	// 请求网络
	url := "https://api.bilibili.com/x/web-interface/ranking"
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	res, err := client.Do(request)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//解析数据
	str, _ := ioutil.ReadAll(res.Body)
	js, err := simplejson.NewJson(str)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	// 装填数据
	var allData []map[string]interface{}
	//lens := len(js.Get("data").MustArray())
	for i := 0; i < 50; i++ {
		item := js.Get("data").Get("list").GetIndex(i).MustMap()
		id := item["aid"]
		title := item["title"]
		rUrl := "https://www.bilibili.com/video/" + item["bvid"].(string)
		desc := item["play"].(json.Number).String() + " 次播放"
		pic := item["pic"]
		allData = append(allData, map[string]interface{}{
			"id":    id,
			"title": title,
			"url":   rUrl,
			"desc":  desc,
			"pic":   pic})
	}
	return allData
}

// GetV2EX V2EX
func (spider Spider) GetV2EX() []map[string]interface{} {
	// 基础设置
	timeout := 5 * time.Second //超时时间5s
	//proxy, _ := uUrl.Parse("http://127.0.0.1:10809")
	proxy, _ := uUrl.Parse("http://127.0.0.1:12307")
	tr := &http.Transport{
		Proxy: http.ProxyURL(proxy),
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}
	// 请求网络
	url := "https://www.v2ex.com/api/topics/hot.json"
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	res, err := client.Do(request)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//解析数据
	str, _ := ioutil.ReadAll(res.Body)
	js, err := simplejson.NewJson(str)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	// 装填数据
	var allData []map[string]interface{}
	lens := len(js.MustArray())
	for i := 0; i < lens; i++ {
		item := js.GetIndex(i).MustMap()
		id := item["id"].(json.Number).String()
		title := item["title"]
		rUrl := item["url"]
		//desc := item["content"]
		desc := item["replies"].(json.Number).String() + " 条回复"
		pic := item["node"].(map[string]interface{})["avatar_large"]
		allData = append(allData, map[string]interface{}{
			"id":    id,
			"title": title,
			"url":   rUrl,
			"desc":  desc,
			"pic":   pic})
	}
	return allData
}

// Get36Kr 没有api，先用原来的吧
func (spider Spider) Get36Kr() []map[string]interface{} {
	url := "https://36kr.com/"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	request.Header.Add("Host", `36kr.com`)
	request.Header.Add("Referer", `https://36kr.com/`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str,_ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find(".hotlist-item-toptwo").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := selection.Find("a p").Text()
		//TODO: 36kr 的题图
		//img := selection.Find("img")
		//println(img2)
		if boolUrl {
			allData = append(allData, map[string]interface{}{
				"id":    strings.Replace(url, "/p/", "", -1),
				"title": string(text),
				"url":   "https://36kr.com" + url})
		}
	})
	document.Find(".hotlist-item-other-info").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if boolUrl {
			allData = append(allData, map[string]interface{}{
				"id":    strings.Replace(url, "/p/", "", -1),
				"title": string(text),
				"url":   "https://36kr.com" + url})
		}
	})
	return allData
}

// GetHuXiu 没有api，先用原来的吧
func (spider Spider) GetHuXiu() []map[string]interface{} {
	url := "https://www.huxiu.com/article"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	request.Header.Add("Host", `www.guokr.com`)
	request.Header.Add("Referer", `https://www.huxiu.com/channel/107.html`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str,_ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find(".article-item--large__content").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		//TODO: 虎嗅的题图
		//img := selection.Find("img").First()
		//pic, _ := img.Attr("src")
		text := s.Find("h5").Text()
		if len(text) != 0 {
			if boolUrl {
				var id string = strings.Replace(url, "https://www.huxiu.com/article/", "", -1)
				id = strings.Replace(id, ".html", "", -1)
				allData = append(allData, map[string]interface{}{
					"id":    id,
					"title": strings.TrimSpace(text),
					"url":   url})
			}
		}
	})
	document.Find(".article-item--normal").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		//img := selection.Find("img").First()
		//pic, _ := img.Attr("data-src")
		text := s.Find("h5").Text()
		if len(text) != 0 {
			if boolUrl {
				var id string = strings.Replace(url, "/article/", "", -1)
				id = strings.Replace(id, ".html", "", -1)
				allData = append(allData, map[string]interface{}{
					"id":    id,
					"title": strings.TrimSpace(text),
					"url":   "https://www.huxiu.com" + url})
			}
		}
	})
	return allData
}

// GetClover 随想
func (spider Spider) clover(mid string) []map[string]interface{} {
	// 基础设置
	timeout := 5 * time.Second //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	// 请求网络
	url := "https://idealclover.top/uniapi/getpostsbymid?mid=" + mid + "&apisec=idealclover"
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	res, err := client.Do(request)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//解析数据
	str, _ := ioutil.ReadAll(res.Body)
	js, err := simplejson.NewJson(str)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	// 装填数据
	var allData []map[string]interface{}
	lens := len(js.Get("data").MustArray())
	for i := 0; i < lens; i++ {
		item := js.Get("data").GetIndex(i).MustMap()
		id := item["cid"].(string)
		title := item["title"]
		rUrl := "https://idealclover.top/archives/" + item["cid"].(string)
		desc := item["commentsNum"].(string) + "条评论"
		pic := ""
		allData = append(allData, map[string]interface{}{
			"id":    id,
			"title": title,
			"url":   rUrl,
			"desc":  desc,
			"pic":   pic})
	}
	return allData
}

func (spider Spider) GetSuiXiang() []map[string]interface{} {
	return spider.clover("466")
}

func (spider Spider) GetJiShu() []map[string]interface{} {
	return spider.clover("446")
}

func (spider Spider) GetShengHuo() []map[string]interface{} {
	return spider.clover("445")
}

func (spider Spider) GetCePing() []map[string]interface{} {
	return spider.clover("410")
}

func (spider Spider) GetITHome() []map[string]interface{} {
	url := "https://www.ithome.com/"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Mobile Safari/537.36`)
	res, err := client.Do(request)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	var allData []map[string]interface{}
	document.Find(".hot-list .bx ul li").Each(func(i int, selection *goquery.Selection) {
		url, boolUrl := selection.Find("a").Attr("href")
		text := selection.Find("a").Text()
		if boolUrl {
			allData = append(allData, map[string]interface{}{"title": text, "url": url})
		}
	})
	return allData
}

// 贴吧
func (spider Spider) GetTieBa() []map[string]interface{} {
	url := "http://tieba.baidu.com/hottopic/browse/topicList"
	res, err := http.Get(url)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	str, _ := ioutil.ReadAll(res.Body)
	js, err2 := simplejson.NewJson(str)
	if err2 != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	var allData []map[string]interface{}
	i := 1
	for i < 30 {
		test := js.Get("data").Get("bang_topic").Get("topic_list").GetIndex(i).MustMap()
		allData = append(allData, map[string]interface{}{"title": test["topic_name"], "url": test["topic_url"]})
		i++
	}
	return allData

}

// 豆瓣：暂不可用
//func (spider Spider) GetDouBan() []map[string]interface{} {
//	// 基础设置
//	timeout := 5 * time.Second //超时时间5s
//	client := &http.Client{
//		Timeout: timeout,
//	}
//	// 请求网络
//	url := "https://m.douban.com/rexxar/api/v2/subject_collection/book_/items?start=0&count=10"
//	var Body io.Reader
//	request, err := http.NewRequest("GET", url, Body)
//	request.Header.Add("Referer", `https://m.douban.com/book/`)
//	if err != nil {
//		fmt.Println("抓取" + spider.DataType + "失败")
//		return []map[string]interface{}{}
//	}
//	res, err := client.Do(request)
//	if err != nil {
//		fmt.Println("抓取" + spider.DataType + "失败")
//		return []map[string]interface{}{}
//	}
//	defer res.Body.Close()
//	//解析数据
//	str, _ := ioutil.ReadAll(res.Body)
//	js, err := simplejson.NewJson(str)
//	if err != nil {
//		fmt.Println("抓取" + spider.DataType + "失败")
//		return []map[string]interface{}{}
//	}
//	println(js.String())
//
//	//url := "https://www.douban.com/group/explore"
//	//timeout := time.Duration(5 * time.Second) //超时时间5s
//	//client := &http.Client{
//	//	Timeout: timeout,
//	//}
//	//var Body io.Reader
//	//request, err := http.NewRequest("GET", url, Body)
//	//if err != nil {
//	//	fmt.Println("抓取" + spider.DataType + "失败")
//	//	return []map[string]interface{}{}
//	//}
//	//request.Header.Add("User-Agent", `Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Mobile Safari/537.36`)
//	//request.Header.Add("Upgrade-Insecure-Requests", `1`)
//	//request.Header.Add("Referer", `https://www.douban.com/group/explore`)
//	//request.Header.Add("Host", `www.douban.com`)
//	//res, err := client.Do(request)
//	//
//	//if err != nil {
//	//	fmt.Println("抓取" + spider.DataType + "失败")
//	//	return []map[string]interface{}{}
//	//}
//	//defer res.Body.Close()
//	////str,_ := ioutil.ReadAll(res.Body)
//	////fmt.Println(string(str))
//	var allData []map[string]interface{}
//	//document, err := goquery.NewDocumentFromReader(res.Body)
//	//if err != nil {
//	//	fmt.Println("抓取" + spider.DataType + "失败")
//	//	return []map[string]interface{}{}
//	//}
//	//document.Find(".channel-item").Each(func(i int, selection *goquery.Selection) {
//	//	url, boolUrl := selection.Find("h3 a").Attr("href")
//	//	text := selection.Find("h3 a").Text()
//	//	if boolUrl {
//	//		allData = append(allData, map[string]interface{}{"title": text, "url": url})
//	//	}
//	//})
//	return allData
//}

// 天涯
func (spider Spider) GetTianYa() []map[string]interface{} {
	url := "http://bbs.tianya.cn/list.jsp?item=funinfo&grade=3&order=1"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Mobile Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	request.Header.Add("Referer", `http://bbs.tianya.cn/list.jsp?item=funinfo&grade=3&order=1`)
	request.Header.Add("Host", `bbs.tianya.cn`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str,_ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find("table tr").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("td a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if boolUrl {
			allData = append(allData, map[string]interface{}{"title": text, "url": "http://bbs.tianya.cn/" + url})
		}
	})
	return allData
}

// 虎扑
func (spider Spider) GetHuPu() []map[string]interface{} {
	url := "https://bbs.hupu.com/all-gambia"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Mobile Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	request.Header.Add("Referer", `https://bbs.hupu.com/`)
	request.Header.Add("Host", `bbs.hupu.com`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str,_ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find(".bbsHotPit li").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find(".textSpan a")
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if boolUrl {
			allData = append(allData, map[string]interface{}{"title": text, "url": "https://bbs.hupu.com/" + url})
		}
	})
	return allData
}

// Github
func (spider Spider) GetGitHub() []map[string]interface{} {
	url := "https://github.com/trending"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Mobile Safari/537.36`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str,_ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}

	document.Find(".Box article").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find(".lh-condensed a")
		//desc := selection.Find(".col-9 .text-gray .my-1 .pr-4")
		//descText := desc.Text()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		descText := selection.Find("p").Text()
		if boolUrl {
			allData = append(allData, map[string]interface{}{"title": text, "desc": descText, "url": "https://github.com" + url})
		}
	})
	return allData
}

func (spider Spider) GetBaiDu() []map[string]interface{} {
	url := "http://top.baidu.com/buzz?b=341&c=513&fr=topbuzz_b1"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Mobile Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	request.Header.Add("Host", `top.baidu.com`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find("table tr").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		MyText, _ := GbkToUtf8([]byte(text))
		if boolUrl {
			allData = append(allData, map[string]interface{}{"title": string(MyText), "url": url})
		}
	})
	return allData

}

func (spider Spider) GetQDaily() []map[string]interface{} {
	url := "https://www.qdaily.com/tags/29.html"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	request.Header.Add("Host", `www.qdaily.com`)
	request.Header.Add("Referer", `https://www.qdaily.com/tags/30.html`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str,_ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find(".packery-item").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := selection.Find(".grid-article-bd h3").Text()
		if len(text) != 0 {
			if boolUrl {
				allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://www.qdaily.com/" + url})
			}
		}
	})
	return allData
}

func (spider Spider) GetGuoKr() []map[string]interface{} {
	url := "https://www.guokr.com/scientific/"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	request.Header.Add("Host", `www.guokr.com`)
	request.Header.Add("Referer", `https://www.guokr.com/scientific/`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str,_ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find("div .article").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("h3 a")
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				allData = append(allData, map[string]interface{}{"title": string(text), "url": url})
			}
		}
	})
	return allData
}

func (spider Spider) GetDBMovie() []map[string]interface{} {
	url := "https://movie.douban.com/"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find(".slide-container").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a")
		url, boolUrl := s.Attr("href")
		text := s.Find("p").Text()
		if len(text) != 0 {
			if boolUrl {
				allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://www.huxiu.com" + url})
			}
		}
	})
	return allData
}

func (spider Spider) GetZHDaily() []map[string]interface{} {
	url := "http://daily.zhihu.com/"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find(".row .box").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Find("span").Text()
		if len(text) != 0 {
			if boolUrl {
				allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://daily.zhihu.com" + url})
			}
		}
	})
	return allData
}

func (spider Spider) GetSegmentfault() []map[string]interface{} {
	url := "https://segmentfault.com/hottest"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find(".news-list .news__item-info").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a:nth-child(2)").First()
		url, boolUrl := s.Attr("href")
		text := s.Find("h4").Text()
		if len(text) != 0 {
			if boolUrl {
				allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://segmentfault.com" + url})
			}
		}
	})
	return allData
}

func (spider Spider) GetHacPai() []map[string]interface{} {
	url := "https://hacpai.com/domain/play"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find(".hotkey li").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("h2 a")
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				allData = append(allData, map[string]interface{}{"title": string(text), "url": url})
			}
		}
	})
	return allData
}

func (spider Spider) GetWYNews() []map[string]interface{} {
	url := "http://news.163.com/special/0001386F/rank_whole.html"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find("table tr").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("td a").First()
		url, boolUrl := s.Attr("href")
		text, _ := GbkToUtf8([]byte(s.Text()))
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": url})
				}
			}
		}
	})
	return allData
}

func (spider Spider) GetWaterAndWood() []map[string]interface{} {
	url := "https://www.newsmth.net/nForum/mainpage?ajax"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str, _ := ioutil.ReadAll(res.Body)
	//sss,_ := GbkToUtf8([]byte(string(str)))
	//fmt.Println(string(sss))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	// topics
	document.Find("#top10 li").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a:nth-child(2)").First()
		url, boolUrl := s.Attr("href")
		text, _ := GbkToUtf8([]byte(s.Text()))
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://www.newsmth.net" + url})
				}
			}
		}
	})
	document.Find(".topics").Find("li").Each(func(i int, selection *goquery.Selection) {
		if i > 10 {
			s := selection.Find("a:nth-child(2)").First()
			url, boolUrl := s.Attr("href")
			text, _ := GbkToUtf8([]byte(s.Text()))
			if len(text) != 0 {
				if boolUrl {
					if len(allData) <= 100 {
						allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://www.newsmth.net" + url})
					}
				}
			}
		}
	})
	return allData
}

// http://nga.cn/

func (spider Spider) GetNGA() []map[string]interface{} {
	url := "http://nga.cn/"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find("h2").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": url})
				}
			}
		}
	})
	return allData
}

func (spider Spider) GetCSDN() []map[string]interface{} {
	url := "https://www.csdn.net/"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find("#feedlist_id li").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("h2 a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": url})
				}
			}
		}
	})
	return allData
}

// https://weixin.sogou.com/?pid=sogou-wsse-721e049e9903c3a7&kw=
func (spider Spider) GetWeiXin() []map[string]interface{} {
	url := "https://weixin.sogou.com/?pid=sogou-wsse-721e049e9903c3a7&kw="
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find(".news-list li").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("h3 a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": url})
				}
			}
		}
	})
	return allData
}

//

func (spider Spider) GetKD() []map[string]interface{} {
	url := "http://www.kdnet.net/"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find(".indexside-box-hot li").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text, _ := GbkToUtf8([]byte(s.Text()))
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": url})
				}
			}
		}
	})
	return allData
}

// http://www.mop.com/

func (spider Spider) GetMop() []map[string]interface{} {
	url := "http://www.mop.com/"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find(".swiper-slide").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := selection.Find("h2").Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": url})
				}
			}
		}
	})
	document.Find(".tabel-right").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := selection.Find("h3").Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": url})
				}
			}
		}
	})
	return allData[:15]
}

// https://www.chiphell.com/

func (spider Spider) GetChiphell() []map[string]interface{} {
	url := "https://www.chiphell.com/"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find("#frameZ3L5I7 li").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://www.chiphell.com/" + url})
				}
			}
		}
	})
	// portal_block_530_content
	document.Find("#portal_block_530_content dt").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://www.chiphell.com/" + url})
				}
			}
		}
	})
	// frame-tab move-span cl
	document.Find("#portal_block_560_content dt").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://www.chiphell.com/" + url})
				}
			}
		}
	})
	// portal_block_564_content
	document.Find("#portal_block_564_content dt").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://www.chiphell.com/" + url})
				}
			}
		}
	})
	// portal_block_568_content
	document.Find("#portal_block_568_content dt").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://www.chiphell.com/" + url})
				}
			}
		}
	})
	// portal_block_569_content
	document.Find("#portal_block_569_content dt").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://www.chiphell.com/" + url})
				}
			}
		}
	})
	// portal_block_570_content
	document.Find("#portal_block_570_content dt").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": "https://www.chiphell.com/" + url})
				}
			}
		}
	})
	return allData
}

// http://jandan.net/

func (spider Spider) GetJianDan() []map[string]interface{} {
	url := "http://jandan.net/"
	timeout := time.Duration(5 * time.Second) //超时时间5s
	client := &http.Client{
		Timeout: timeout,
	}
	var Body io.Reader
	request, err := http.NewRequest("GET", url, Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	request.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36`)
	request.Header.Add("Upgrade-Insecure-Requests", `1`)
	res, err := client.Do(request)

	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	defer res.Body.Close()
	//str, _ := ioutil.ReadAll(res.Body)
	//fmt.Println(string(str))
	var allData []map[string]interface{}
	document, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	document.Find("h2").Each(func(i int, selection *goquery.Selection) {
		s := selection.Find("a").First()
		url, boolUrl := s.Attr("href")
		text := s.Text()
		if len(text) != 0 {
			if boolUrl {
				if len(allData) <= 100 {
					allData = append(allData, map[string]interface{}{"title": string(text), "url": url})
				}
			}
		}
	})
	return allData
}

// https://dig.chouti.com/

func (spider Spider) GetChouTi() []map[string]interface{} {
	url := "https://dig.chouti.com/top/24hr?_=" + strconv.FormatInt(time.Now().Unix(), 10) + "163"
	url2 := "https://dig.chouti.com/link/hot?afterTime=" + strconv.FormatInt(time.Now().Unix(), 10) + "026000" + "&_=" + strconv.FormatInt(time.Now().Unix(), 10) + "667"
	res, err := http.Get(url)
	res2, _ := http.Get(url2)
	if err != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	str, _ := ioutil.ReadAll(res.Body)
	str2, _ := ioutil.ReadAll(res2.Body)
	js, err2 := simplejson.NewJson(str)
	js2, _ := simplejson.NewJson(str2)
	if err2 != nil {
		fmt.Println("抓取" + spider.DataType + "失败")
		return []map[string]interface{}{}
	}
	var allData []map[string]interface{}
	i := 1
	for i < 30 {
		test := js.Get("data").GetIndex(i).MustMap()
		if test["title"] != nil && test["url"] != nil {
			allData = append(allData, map[string]interface{}{"title": test["title"], "url": test["url"]})
		}
		i++
	}
	j := 1
	for j < 60 {
		test := js2.Get("data").GetIndex(j).MustMap()
		if test["title"] != nil && test["url"] != nil {
			allData = append(allData, map[string]interface{}{"title": test["title"], "url": test["url"]})
		}
		j++
	}
	return allData

}

/**
部分热榜标题需要转码
*/
func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

/**
执行每个分类数据
*/
func ExecGetData(spider Spider) {
	reflectValue := reflect.ValueOf(spider)
	dataType := reflectValue.MethodByName("Get" + spider.DataType)
	data := dataType.Call(nil)
	originData := data[0].Interface().([]map[string]interface{})
	start := time.Now()
	//fmt.Printf(SaveDataToJson(originData))
	Common.MySql{}.GetConn().Where(map[string]string{"dataType": spider.DataType}).Update("hotData2", map[string]string{"str": SaveDataToJson(originData)})
	group.Done()
	seconds := time.Since(start).Seconds()
	fmt.Printf("耗费 %.2fs 秒完成抓取%s", seconds, spider.DataType)
	fmt.Println()

}

var group sync.WaitGroup

func main() {
	allData := []string{
		"ZhiHu",
		"WeiBo",
		"Bili",
		"V2EX",
		"36Kr",
		"HuXiu",

		"SuiXiang",
		"JiShu",
		"ShengHuo",
		"CePing",

		//"TieBa",
		//"DouBan",
		//"TianYa",
		//"HuPu",
		//"GitHub",
		//"BaiDu",
		//"QDaily",
		//"GuoKr",
		//"ZHDaily",
		//"Segmentfault",
		//"WYNews",
		//"WaterAndWood",
		//"HacPai",
		//"KD",
		//"NGA",
		//"WeiXin",
		//"Mop",
		//"Chiphell",
		//"JianDan",
		//"ChouTi",
		//"ITHome",
	}
	fmt.Println("开始抓取" + strconv.Itoa(len(allData)) + "种数据类型")
	group.Add(len(allData))
	var spider Spider
	for _, value := range allData {
		fmt.Println("开始抓取" + value)
		spider = Spider{DataType: value}
		go ExecGetData(spider)
	}
	group.Wait()
	fmt.Print("完成抓取")
}
