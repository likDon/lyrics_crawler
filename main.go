package main

import (
	"bufio"
	"fmt"
	"github.com/yanyiwu/gojieba"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"unicode/utf8"
)

type WordNum struct {
	Word   string
	Num int
}

func main() {
	//爬取文件
	for i := 1; i < len(os.Args); i++ {
		CrawlData(os.Args[i])
	}
	//词频统计
	var words []string
	counts := make(map[string]int)
	var WordsList []WordNum
	use_hmm := true
	x := gojieba.NewJieba()
	defer x.Free()

	//遍历文件夹,统计词频
	dir,_ := ioutil.ReadDir("./output")
	for _, file := range dir {
		data, err := ioutil.ReadFile("./output/"+file.Name())
		if err!=nil{
			fmt.Println("File reading error", err)
		}
		//去除文本中的标点
		punc := regexp.MustCompile("[\\pP]")
		data = []byte(punc.ReplaceAllString(string(data), " "))
		words = x.Cut(string(data), use_hmm)//精确模式
		for _, word := range words {
			//不统计字数为1的词
			if utf8.RuneCountInString(word) == 1 {
				continue
			}
			counts[word] += 1
		}
	}
	//设置停用词
	stopwords := []string{"作词","作曲"}
	fi, err := os.Open("stopwords.txt")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	defer fi.Close()
	br := bufio.NewReader(fi)
	for {
		w, _, end := br.ReadLine()
		if end == io.EOF {
			break
		}
		stopwords=append(stopwords,string(w))
	}
	//去除停用词
	for _,word := range stopwords{
		if _, ok := counts[word]; ok {
			delete(counts, word)
		}
	}
	//把字典转换成成列表
	for word,n := range counts {
		WordsList = append(WordsList, WordNum{word, n})
	}
	//降序排列
	sort.Slice(WordsList, func(i, j int) bool {
		return WordsList[i].Num > WordsList[j].Num
	})
	//展示的词频个数
	showCount := 100
	for i := 0; i < showCount; i++ {
		fmt.Printf("%s\t %d\n", WordsList[i].Word, WordsList[i].Num)
	}
}
