package apiclient

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/net/html/charset"
)

type KommersantAPI struct {
	HttpClient *http.Client
}

type Rss struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Language      string `xml:"language"`
	Copyright     string `xml:"copyright"`
	Docs          string `xml:"docs"`
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Description   string `xml:"description"`
	LastBuildDate string `xml:"lastBuildDate"`
	Image         Image  `xml:"image"`
	Item          []Item `xml:"item"`
}

type Image struct {
	Url   string `xml:"url"`
	Title string `xml:"title"`
	Link  string `xml:"link"`
}

type Item struct {
	Guid        string `xml:"guid"`
	Category    string `xml:"category"`
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
}

func (k *KommersantAPI) CallKommersantAPI() ([]Item, error) {
	const max_retry int = 3
	base_url := "https://www.kommersant.ru/RSS/news.xml"
	resp := &http.Response{}

	for i := 1; i <= max_retry; i += 1 {
		resp, err := k.HttpClient.Get(base_url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode/100 == 2 {
			break
		} else if i < max_retry {
			slog.Info("got error response from api, retrying...", "retry-cnt", i, "status", resp.Status)
			time.Sleep(3 * time.Second)
		} else {
			return nil, fmt.Errorf("got error response from api: %s", resp.Status)
		}
	}
	var rss Rss
	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = charset.NewReaderLabel
	err := decoder.Decode(&rss)
	if err != nil {
		return nil, err
	}
	return rss.Channel.Item, nil
}
