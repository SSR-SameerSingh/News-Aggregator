package main 

//importing packages
import (
	"fmt"
	"net/http"
	"html/template"
	"io/ioutil"
	"encoding/xml"
	"sync"
)

var wg sync.WaitGroup

type SitemapIndex struct {
	Locations []string `xml:"sitemap>loc"` // '>' to go to next tag
}

type News struct {
	Titles []string `xml:"url>news>title"`
	Keywords []string `xml:"url>news>keywords"`
	Locations []string `xml:"url>loc"`
}

type NewsMap struct {
	Keywords string
	Locations string
}

type NewsAggPage struct {
	Title string
	News map[string]NewsMap
}

func newsAggHandler(w http.ResponseWriter, r *http.Request) {
	var s SitemapIndex

	news_map := make(map[string]NewsMap)

	resp, _ := http.Get("https://www.washingtonpost.com/news-sitemaps/index.xml") //response, error (make request and get response)
	bytes, _ := ioutil.ReadAll(resp.Body)

	xml.Unmarshal(bytes, &s) // unmarshal "bytes" into memory address of "s"

	resp.Body.Close()

	queue := make(chan News, 500)
	for _, loc := range(s.Locations) {
		wg.Add(1)
		go newsRoutine(queue, loc)
	}

	wg.Wait()
	close(queue)
	for elem := range queue {
		for idx, _ := range elem.Keywords {
			news_map[elem.Titles[idx]] = NewsMap{elem.Keywords[idx], elem.Locations[idx]}
		}
	}

	p := NewsAggPage{Title:"Washington Post News Aggregator", News: news_map}
	
	t, _ := template.ParseFiles("template.html")
	fmt.Println(t.Execute(w, p))
}

func newsRoutine(c chan News, loc string) {
	defer wg.Done()
	var n News

	resp, _ := http.Get(loc[1:len(loc)-1])
	bytes, _ := ioutil.ReadAll(resp.Body)
	xml.Unmarshal(bytes, &n)
	resp.Body.Close()

	c <- n
}

func main() {
	http.HandleFunc("/aggregator", newsAggHandler)
	http.ListenAndServe(":8000", nil)
}