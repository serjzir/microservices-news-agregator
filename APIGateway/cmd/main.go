package main

import (
	"bytes"
	"encoding/json"
	"github.com/serjzir/APIGateway/model"
	"github.com/serjzir/APIGateway/pkg/logging"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
)

func main()  {
	logger := logging.Init()
	logger.Info("Run APIGateway")
	router := gin.Default()
	router.GET("/news/:id", newsById)
	router.GET("/news", allNews)
	router.GET("/news/search", search)
	router.POST("/comments/:id", addComment)
	router.POST("/comments/:id/:slug", addResponnseComment)
	router.Run(":8000")

}

func newsById(c *gin.Context) {
	id := c.Param("id")
	urlNews := (fmt.Sprintf("http://localhost:8080/news/%s", id))
	urlComments := fmt.Sprintf("http://localhost:8888/comments/%s", id)
	newsChan := make(chan *http.Response, 1)
	commentChan := make(chan *http.Response, 1)
	errGrp, _ := errgroup.WithContext(context.Background())
	errGrp.Go(func() error { return makeApiCall(urlNews, newsChan)})

	errGrp.Go(func() error { return makeApiCall(urlComments, commentChan)})
	err := errGrp.Wait()
	if err != nil {
		c.IndentedJSON(http.StatusBadGateway, gin.H{"message": "Error with submitting the news, try again later..."})
		return
	}

	newsResponse := <-newsChan
	defer newsResponse.Body.Close()
	newsBytes, _ := ioutil.ReadAll(newsResponse.Body)

	var news []model.NewsFullDetailed
	_ = json.Unmarshal(newsBytes, &news)


	commentResponse := <- commentChan
	defer commentResponse.Body.Close()
	commentBytes, _ := ioutil.ReadAll(commentResponse.Body)
	var comment []model.Comment
	err = json.Unmarshal(commentBytes, &comment)
	news[0].Comment = comment
	c.IndentedJSON(http.StatusOK, news)
}

func allNews(c *gin.Context) {
	urlNews := fmt.Sprintf("http://localhost:8080/news?limit=%s&page=%s", c.Query("limit"), c.Query("page"))
	errGrp, _ := errgroup.WithContext(context.Background())
	newsChan := make(chan *http.Response, 1)
	errGrp.Go(func() error { return makeApiCall(urlNews, newsChan)})

	newsResponse := <-newsChan
	defer newsResponse.Body.Close()
	newsBytes, _ := ioutil.ReadAll(newsResponse.Body)

	var news []model.NewsShortDetailed
	_ = json.Unmarshal(newsBytes, &news)
	if news == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Page not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, news)
}


func search(c *gin.Context) {
	limit := c.Query("limit")
	page := c.Query("page")
	title := c.Query("title")
	urlNews := fmt.Sprintf("http://localhost:8080/news/search?title=%s&limit=%s&page=%s", title, limit, page)
	errGrp, _ := errgroup.WithContext(context.Background())
	newsChan := make(chan *http.Response, 1)
	errGrp.Go(func() error { return makeApiCall(urlNews, newsChan)})

	newsResponse := <-newsChan
	defer newsResponse.Body.Close()

	newsBytes, _ := ioutil.ReadAll(newsResponse.Body)


	var news []model.NewsFullDetailed
	_ = json.Unmarshal(newsBytes, &news)
	c.IndentedJSON(http.StatusOK, news)
}

func addComment(c *gin.Context) {
	id := c.Param("id")
	var body model.Comment
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	urlComments := fmt.Sprintf("http://localhost:8888/comments/%s", id)
	resp, err := apiCallPost(urlComments, body, "")
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "comment not added"})
	}
	c.IndentedJSON(http.StatusCreated, gin.H{"message": resp})
}

func addResponnseComment(c *gin.Context) {
	id := c.Param("id")
	parentSlug := c.Param("slug")
	var body model.Comment
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	urlComments := fmt.Sprintf("http://localhost:8888/comments/%s/%s", id, parentSlug)
	resp, err := apiCallPost(urlComments, body, parentSlug)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "comment not added"})
	}
	c.IndentedJSON(http.StatusCreated, gin.H{"message": resp})
}

func apiCallPost(urls string, body model.Comment, slug string) (string, error) {
	data, err := json.Marshal(map[string]string{
		"author": body.Author,
		"text": body.Text,
	})
	if slug != "" {
		data, err = json.Marshal(map[string]string{
			"slug": body.Slug,
			"author": body.Author,
			"text": body.Text,
			})
	}
	if err != nil {
		return "", err
	}
	response, err := http.Post(urls, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	return "Comment added", nil
}
func makeApiCall(url string, rc chan *http.Response) error {
	response, err := http.Get(url)
	if err == nil {
		rc <- response
	}
	return  err
}