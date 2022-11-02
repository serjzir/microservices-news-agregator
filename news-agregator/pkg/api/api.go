package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	middlewares "github.com/serjzir/news-agregator/pkg/middleware"
	"github.com/serjzir/news-agregator/pkg/storage"
	"net/http"
	"strconv"
)

type API struct {
	posts storage.Repository
	r     *gin.Engine
}

// New конструктор API
func New(db *storage.Repository) *API {
	router := gin.Default()
	return &API{posts: *db, r: router}
}

// GetRouter конструктор API - содает router и запускает веб сервис.
func (api *API) GetRouter(ip, port string) *gin.Engine {
	api.r.Use(middlewares.RequestID()) // добавить request_id в заголовок каждого запроса
	api.r.GET("/news/:id", api.getNewsId)
	api.r.GET("/news", api.newsPagination)
	api.r.GET("/news/search", api.search)
	api.r.Run(ip + ":" + port)
	return api.r
}

func (api *API) getNewsId(c *gin.Context) {
	strID := c.Param("id")
	id, _ := strconv.Atoi(strID)
	news, err := api.posts.News(c, id)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "BadRequest"})
	}
	c.IndentedJSON(http.StatusOK, news)
}

func (api *API) search(c *gin.Context) {
	limit := c.Query("limit")
	page := c.Query("page")
	title := c.Query("title")
	fmt.Println(title)
	news, err := api.posts.Searche(c, title, limit, page)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Title not found"})
	}
	if news == nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "BadRequest. Title Not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, news)
}

func (api *API) newsPagination(c *gin.Context) {
	limit := c.DefaultQuery("limit", "10")
	page := c.Query("page")
	news, err := api.posts.PaginationRequest(c, limit, page)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"status": "fail", "message": err.Error()})
		return
	}
	if news == nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "BadRequest. Page Not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, news)
}