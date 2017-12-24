package main

import (
	"log"

	"github.com/PuerkitoBio/goquery"
	"net/http"
	"io/ioutil"
	"os"
	"io"
	"bytes"
	"strings"
	"sync"
	"fmt"
)

var wg sync.WaitGroup

func ExampleScrape() {
	doc, err := goquery.NewDocument("http://md.itlun.cn/")

	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".pic li a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")

		if exists == false || href == "" {
			return
		}

		wg.Add(1)

		go func() {
			defer wg.Done()
			fetchPage(href)
		}()
	})
}

func fetchPage(href string) {
	doc, err := goquery.NewDocument("http://md.itlun.cn" + href);
	if err != nil {
		return
	}

	imgEl := doc.Find("#imgString a")
	nextLink, exists := imgEl.Attr("href")

	if exists == false || nextLink == "" {
		return
	}

	imgLink, exists := imgEl.Find("img").Attr("src")

	if exists == false || imgLink == "" {
		return
	}

	wg.Add(1)
	downloadImg(imgLink)

	fetchPage(nextLink)
}

func downloadImg(imgPath string) {
	defer wg.Done()

	if strings.HasPrefix(imgPath, "//") {
		imgPath = "http:" + imgPath
	}
	fmt.Println("Fetch picture: ", imgPath)
	filename := imgPath[strings.LastIndex(imgPath, "/")+1:]
	resp, _ := http.Get(imgPath)
	body, _ := ioutil.ReadAll(resp.Body)
	out, _ := os.Create("img/" + filename)

	io.Copy(out, bytes.NewReader(body))
}

func main() {
	os.Mkdir("img", 0777)
	ExampleScrape()
	wg.Wait()
}
