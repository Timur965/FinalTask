// Пакет для работы с RSS-потоками.
package rss

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	storage "FinalTask/Storage"

	strip "github.com/grokify/html-strip-tags-go"
)

type Feed struct {
	XMLName xml.Name `xml:"rss"`
	Chanel  Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Link        string `xml:"link"`
}

func Parse(url string) ([]storage.News, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var f Feed
	err = xml.Unmarshal(b, &f)
	if err != nil {
		return nil, err
	}
	var data []storage.News
	for _, item := range f.Chanel.Items {
		var p storage.News
		p.Title = item.Title
		p.Content = item.Description
		p.Content = strip.StripTags(p.Content)
		// Sat, 15 May 2021 04:05:00 +0300
		item.PubDate = strings.ReplaceAll(item.PubDate, ",", "")
		t, err := time.Parse("Mon 2 Jan 2006 15:04:05 -0700", item.PubDate)
		if err != nil {
			t, err = time.Parse("Mon 2 Jan 2006 15:04:05 GMT", item.PubDate)
		}
		if err == nil {
			p.CreatedAt = t.Unix()
		}
		data = append(data, p)
	}
	return data, nil
}
