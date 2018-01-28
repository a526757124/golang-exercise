package main

import (
	"strings"
	"net/http"
	"io/ioutil"

	"github.com/PuerkitoBio/goquery"
	"os"
	"io"
	"bytes"
	"sync"
	"log"
	"path/filepath"
	"fmt"
	"strconv"
	"time"
	"regexp"
	"math/rand"
)

const baseUrl = "http://www.mzitu.com/"

var (
	rwg       sync.WaitGroup
	basePath  string
	cFlag     int
	cPage     int
	flagLabel = []string{
		"",
		"xinggan",
		"japan",
		"taiwan",
		"mm",
		"zipai",
	}
	flagDesc = []string{
		"首页推荐",
		"性感妹子",
		"日本妹子",
		"台湾妹子",
		"清纯妹子",
		"妹子自拍",
	}
	userAgents = []string{
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; SV1; AcooBrowser; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.0; Acoo Browser; SLCC1; .NET CLR 2.0.50727; Media Center PC 5.0; .NET CLR 3.0.04506)",
		"Mozilla/4.0 (compatible; MSIE 7.0; AOL 9.5; AOLBuild 4337.35; Windows NT 5.1; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
		"Mozilla/5.0 (Windows; U; MSIE 9.0; Windows NT 9.0; en-US)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Win64; x64; Trident/5.0; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET CLR 2.0.50727; Media Center PC 6.0)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET CLR 1.0.3705; .NET CLR 1.1.4322)",
		"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 5.2; .NET CLR 1.1.4322; .NET CLR 2.0.50727; InfoPath.2; .NET CLR 3.0.04506.30)",
		"Mozilla/5.0 (Windows; U; Windows NT 5.1; zh-CN) AppleWebKit/523.15 (KHTML, like Gecko, Safari/419.3) Arora/0.3 (Change: 287 c9dfb30)",
		"Mozilla/5.0 (X11; U; Linux; en-US) AppleWebKit/527+ (KHTML, like Gecko, Safari/419.3) Arora/0.6",
		"Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US; rv:1.8.1.2pre) Gecko/20070215 K-Ninja/2.1.1",
		"Mozilla/5.0 (Windows; U; Windows NT 5.1; zh-CN; rv:1.9) Gecko/20080705 Firefox/3.0 Kapiko/3.0",
		"Mozilla/5.0 (X11; Linux i686; U;) Gecko/20070322 Kazehakase/0.4.5",
		"Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9.0.8) Gecko Fedora/1.9.0.8-1.fc10 Kazehakase/0.5.6",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_3) AppleWebKit/535.20 (KHTML, like Gecko) Chrome/19.0.1036.7 Safari/535.20",
		"Opera/9.80 (Macintosh; Intel Mac OS X 10.6.8; U; fr) Presto/2.9.168 Version/11.52",
	}
	userAgentsLen = len(userAgents)
)

func main() {

	l := len(flagDesc)

	fmt.Println("\n欢迎使用《极速漂移》，本程序由 Golang 强力驱动。\n希望本程序能给您的生活带来光明与希望！")
	fmt.Println("\n\n请输入图片保存目录路径和内容抓取类别，默认为首页推荐 ")
	fmt.Print("\r\n\r\n")

	for i := 0; i < l; i++ {
		fmt.Println(i, " - ", flagDesc[i])
	}

	fmt.Print("\r\n\r\n")

	for {
		fmt.Println("请输入合法的图片保存目录路径（如 /usr/image、E:\\image）：")
		fmt.Scanln(&basePath)

		if basePath != "" {
			_, err := os.Stat(basePath)

			if err == nil {
				break
			}

			if os.IsNotExist(err) {
				if err = os.MkdirAll(basePath, 0777); err == nil {
					break
				}
			}

			fmt.Println(err.Error())
		}
	}

	fmt.Println("请输入抓取内容类别编码（数字）：")
	fmt.Scanln(&cFlag)
	fmt.Println("请输入起始抓取页码（数字）：")
	fmt.Scanln(&cPage)

	if cFlag < 0 || cFlag > l-1 {
		cFlag = 0
	}

	path := flagLabel[cFlag]
	url := baseUrl

	if path != "" {
		url += "/" + path
	}

	if cPage > 0 {
		url += "/page/" + strconv.Itoa(cPage)
	}

	log.Println("MM Spider start!")
	execute(url, "")
	log.Println("MM Spider shutdown!")
}

func execute(url string, referer string) {
	resp, err := request(url, referer)

	if err != nil {
		return
	}

	doc, err := goquery.NewDocumentFromResponse(resp)

	if err != nil {
		log.Println("Parse page failed, SKIP! ", url, err.Error())
		return
	}

	var currentNode *goquery.Selection

	if currentNode = doc.Find(".main-image"); currentNode.Size() > 0 {
		imgUrl, err := currentNode.Find("p img").Attr("src")
		desc, _ := currentNode.Find("p img").Attr("alt")

		if err && imgUrl != "" {
			saveImage(url, imgUrl, desc)
		}

		selection := doc.Find(".pagenavi").Find("a").Last()
		var nextUrl string

		text := selection.Find("span").Text()

		if strings.TrimSpace(text) == "下一页»" {
			nextUrl, _ = selection.Attr("href")
		}
		if nextUrl == "" {
			return
		}
		time.Sleep(time.Duration(1) * time.Second)
		execute(nextUrl, "")
	} else if currentNode = doc.Find(".postlist"); currentNode.Size() > 0 {
		var index = 0
		var nextPage string

		currentNode.Children().Each(func(i int, s *goquery.Selection) {
			if index == 0 {
				s.Find("li>a").Each(func(idx int, sel *goquery.Selection) {
					detailUrl, exists := sel.Attr("href")

					if !exists || detailUrl == "" {
						return
					}

					detailUrl = strings.TrimPrefix(detailUrl, "/")

					rwg.Add(1)
					go func() {
						defer rwg.Done()
						// 缓冲执行协程，防止过快一起执行
						time.Sleep(time.Duration(i) * time.Second)
						execute(detailUrl, "")
					}()
				})
			} else if index == 1 {
				nextPage, _ = s.Find(".next").Attr("href")
			}
			index++
		})

		if nextPage == "" {
			log.Println("Page end!", url)
			return
		}

		rwg.Wait()
		log.Println("Next page ", nextPage)
		execute(nextPage, "")
	} else {
		log.Println("Invalid page ", url)
	}
}

func request(url string, referer string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	req.Header.Set("User-Agent", userAgents[r.Intn(userAgentsLen)])

	if referer != "" {
		req.Header.Set("Accept", "image/webp,image/*,*/*;q=0.8")
		req.Header.Set("Referer", referer)
	} else {
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		req.Header.Set("Cache-Control", "max-age=0")
		req.Header.Set("Referer", baseUrl)
	}

	return client.Do(req)
}

func saveImage(pageUrl string, imgUrl string, folder string) {
	fragment := strings.TrimPrefix(pageUrl, baseUrl)
	fragments := strings.Split(fragment, "/")
	fragLen := len(fragments)
	idx := "1"

	if fragLen == 0 {
		return
	}

	if folder != "" {
		// 过滤非法文件名
		reg, err := regexp.Compile("[\\\\/:*?\"<>|]")
		if err != nil {
			return
		}
		folder = reg.ReplaceAllString(folder, "_")
	}

	folder = fragments[0] + "_" + strings.TrimSpace(folder)

	if fragLen > 1 {
		idx = fragments[1]
	}

	path := filepath.Join(basePath, flagDesc[cFlag], folder)
	err := os.MkdirAll(path, 0777)

	if err != nil {
		log.Println("Create directory failed! ", path, err.Error())
		return
	}

	ext := filepath.Ext(imgUrl)
	path = filepath.Join(path, idx+ext)

	resp, err := request(imgUrl, pageUrl)

	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return
	}

	out, err := os.Create(path)

	if err != nil {
		log.Println("Create file failed! ", err.Error())
		return
	}

	_, err = io.Copy(out, bytes.NewReader(body))

	if err != nil {
		log.Println("Save file failed! ", err.Error())
	} else {
		log.Println("Saved file: ", path)
	}
}
