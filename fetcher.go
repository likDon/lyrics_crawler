/*
 * 爬取歌词代码 错误处理待修改
 */

package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)
type Album struct{
	Name string
	ID string
	SongList []SongInfo
}
type SongInfo struct{
	ID string
	Name string
	Album string
	Lyric string
}

var wg sync.WaitGroup // 创建同步等待组对象
/*
 * 传入歌手ID,爬取歌手所有专辑内歌曲
 */
func CrawlData(ID string){
	Url := "http://music.163.com/artist/album?id="+ID
	t1 := time.Now()
	resp, _ := Fetch(Url, "local")
	//获取专辑与歌曲信息,存入AlbumList
	GetAlbumPage(resp)
	wg.Add(len(AlbumList))
	//创建文件夹
	os.Mkdir("output", os.ModePerm)
	for i,_ :=range AlbumList {
		go GetSongID(i)
	}
	fmt.Println("main阻塞")
	wg.Wait() //表示main goroutine进入等待，意味着阻塞
	fmt.Println("main解除阻塞")
	fmt.Println("运行时长:",time.Since(t1))
	//将AlbumList写入list.json文件
	data, _ := json.MarshalIndent(AlbumList,"","    ")
	ioutil.WriteFile("list"+ID+".json", data, 0666)
}
/*
 * 对给定url和ip 返回response和error
 * 做简单的错误处理
 */
func Fetch(Url string, ip string) (*http.Response, error) {
	req, err := http.NewRequest("GET", Url, nil)
	if err != nil {
		fmt.Println(err)
	}
	//设置超时时间
	timeout := time.Duration(10 * time.Second)
	client := &http.Client{Timeout:timeout}
	//设置代理ip
	if ip!="local"{
		proxy, err := url.Parse(ip)
		if err!=nil{
			fmt.Errorf("Wrong proxy:%s\n",proxy)
		}
		fmt.Printf("使用代理:%s\n",proxy)
		client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}}
	}
	//设置header 防止网站反爬
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.86 Safari/537.36")
	//请求网页
	response, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil,err
	}
	// 出错处理
	if response.StatusCode != http.StatusOK {
		fmt.Errorf("wrong state code: %d", response.StatusCode)
		return nil,err
	}
	return response, err
}

var AlbumList []Album
/*
 * 对每个歌手的专辑页面第一页，将此歌手所有专辑加入AlbumList
 * 调用GetAlbumList实现
 */
func GetAlbumPage(resp *http.Response){
	fmt.Println("GetAlbumPage Running...")
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err!=nil{
		fmt.Println("ERROR:goquery.NewDocumentFromResponse")
	}
	//将本页专辑加入list
	//body只能打开一次?
	items := doc.Find(".tit.s-fc0")
	items.Each(func(index int, sel *goquery.Selection) {
		album:=new(Album)
		album.Name = sel.Text()
		album.ID,_ = sel.Attr("href")
		AlbumList = append(AlbumList,*album)
		fmt.Printf("NAME:%s   ID:%s\n ",album.Name,album.ID)
	})
	//遍历所有页，将专辑加入list
	pages := doc.Find("a.zpgi")
	pages.Each(func(index int, sel *goquery.Selection) {
		nextPage,_ := sel.Attr("href")
		if nextPage != "javascript:void(0)" {
			response,_ := Fetch("http://music.163.com"+string(nextPage),"local")
			GetAlbumList(response)
		}
	})
}
/*
 * 对每个专辑页面的response，将本页面的专辑加入AlbumList
 */
func GetAlbumList(resp *http.Response){
	fmt.Println("GetAlbumList Running...")
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err!=nil{
		fmt.Println("ERROR:goquery.NewDocumentFromResponse")
	}
	items := doc.Find(".tit.s-fc0")
	items.Each(func(index int, sel *goquery.Selection) {
		album:=new(Album)
		album.Name = sel.Text()
		album.ID,_ = sel.Attr("href")
		AlbumList = append(AlbumList,*album)
		fmt.Printf("NAME:%s ID:%s\n ",album.Name,album.ID)
	})
}
/*
 * 对AlbumList中的每个专辑获取所有歌曲信息，存入SongList
 */
func GetSongID(i int){
	fmt.Println("GetSongID Running...")
	n :=AlbumList[i]
	resp, _ := Fetch("https://music.163.com"+n.ID,"local")
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err!=nil{
		fmt.Println("ERROR:goquery.NewDocumentFromResponse")
	}
	//将当前专辑所有歌曲加入SongList
	fmt.Printf("当前专辑 :%s\n", n.Name)
	items := doc.Find("#song-list-pre-cache").Find("ul.f-hide").Find("a")
	items.Each(func(index int, sel *goquery.Selection) {
		song:=new(SongInfo)
		song.Album = n.Name
		song.Name = sel.Text()
		song.ID,_ = sel.Attr("href")
		//SongList = append(SongList,*song)
		AlbumList[i].SongList = append(AlbumList[i].SongList,*song) //这里不能用n,n是临时变量
		fmt.Printf("--NAME:%s   ID:%s\n ",song.Name,song.ID)
	})
	//获取每个专辑中的歌曲歌词
	GetLyric(i)
	wg.Done()
}

/*
 * 获得所有歌的歌词,写入output文件夹中的"专辑名_歌曲名.txt"
 */
func GetLyric(i int) {
	fmt.Println("GetLyric Running...")
	fmt.Printf("当前专辑：%s\n", AlbumList[i].Name)
	for _, n := range AlbumList[i].SongList {
		fmt.Printf("获取歌词：%s\n", n.Name)
		id := strings.TrimLeft(n.ID, "/song?id=")
		Url := "http://music.163.com/api/song/lyric?id=" + id + "&lv=1&kv=1&tv=-1"
		resp, _ := Fetch(Url, "local")
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			continue
		}
		//将JSON解析为map
		var event map[string]interface{}
		if err := json.Unmarshal(body, &event); err != nil {
			fmt.Println("【错误】：无法解析歌词JSON")
			continue
		}
		//查看是否存在lrc成员
		lrc, ok := event["lrc"]
		if !ok {
			fmt.Println("无歌词")
			continue
		}
		//获得 lrc 中的 lyric字符串
		lrcMap, ok := lrc.(map[string]interface{})
		if !ok {
			fmt.Println("【错误】：无法解析lrc为map")
			continue
		}
		lyric := lrcMap["lyric"].(string)
		pattern := regexp.MustCompile("\\[[\\d|\\.|:]*\\]")
		lyric = pattern.ReplaceAllString(lyric, "")
		//fmt.Println(lyric)
		filename := n.Album + "_" + n.Name + ".txt"
		ioutil.WriteFile("./output/"+filename, []byte(lyric), 0666)
	}
}
