package clientrss

import (
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"github.com/serjzir/news-agregator/pkg/logging"
	"github.com/serjzir/news-agregator/pkg/storage"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var logger = logging.Init()

type Rss struct {
	Channel struct {
		Item []struct {
			Title   string `xml:"title"`
			Link    string `xml:"link"`
			Content string `xml:"description"`
			PubDate string `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}

// CreateHTTPClient создание http клиента который игнорирует ssl
func CreateHTTPClient() http.Client {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := http.Client{Timeout: time.Duration(14) * time.Second, Transport: customTransport}
	return client
}

// HTTPGetNews запрашивает список новостей и возвращает полученные данные в виде []byte
func HTTPGetNews(url string) ([]byte, error) {
	client := CreateHTTPClient()
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	req.Header = http.Header{
		"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36"},
	}
	res, err := client.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	return body, nil
}

// Parse читает rss-поток и возвращет массив раскодированных новостей.
func Parse(url string) ([]storage.Post, error) {
	req, err := HTTPGetNews(url)
	if err != nil {
		return nil, err
	}
	var post Rss
	xml.Unmarshal(req, &post)
	var data []storage.Post
	for _, item := range post.Channel.Item {
		var p storage.Post
		p.Title = item.Title
		p.Link = item.Link
		p.Content = item.Content
		item.PubDate = strings.ReplaceAll(item.PubDate, ",", "")
		t, err := time.Parse("Mon 2 Jan 2006 15:04:05 -0700", item.PubDate)
		if err != nil {
			t, err = time.Parse("Mon 2 Jan 2006 15:04:05 GMT", item.PubDate)
		}
		if err == nil {
			p.PubTime = t.Unix()
		}
		data = append(data, p)
	}
	return data, nil
}
