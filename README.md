# lyrics_crawler
爬取网易云音乐指定歌手的所有歌词，进行词频统计  

## 使用
接收一或多个歌手ID作为命令行参数  
歌词存入output文件夹中，命名格式为“专辑名_歌曲名.txt”  
专辑及所含歌曲列表以json格式存储，命名格式为“list{歌手ID}.json”  
‘go build ’
‘./lyrics_crawler 2116 2117 //分别为陈奕迅与侧田’


### fetcher.go
爬取歌词存入文件  
- func CrawlData(ID string)  
  传入歌手ID,爬取歌手所有专辑内歌曲  
- AlbumList  
  全局变量，Album专辑类型列表  
- func GetAlbumPage(resp *http.Response)  
  对每个歌手的专辑页面第一页，将此歌手所有专辑加入AlbumList  
  调用GetAlbumList实现  
   - func GetAlbumList(resp *http.Response)  
   对每个专辑页面，将页面中专辑加入AlbumList  
- func GetSongID(i int)  
  对第i个专辑爬取所有歌词，并发实现（对goroutine数目暂未限制）  
### main.go  
统计词频

# 引用项目  
gojieba分词来自：https://github.com/yanyiwu/gojieba  
停用词表来自：https://github.com/YueYongDev/stopwords

