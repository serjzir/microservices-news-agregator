package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/serjzir/service-comments/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CommentHandler struct {
	collection  *mongo.Collection
	ctx         context.Context
	redisClient *redis.Client
}

func NewCommentHandler(ctx context.Context, collection *mongo.Collection, redisClient *redis.Client) *CommentHandler {
	return &CommentHandler{
		collection:  collection,
		ctx:         ctx,
		redisClient: redisClient,
	}
}

func (handler *CommentHandler) ListCommentHandler(c *gin.Context) {
	val, err := handler.redisClient.Get("comments").Result()
	if err == redis.Nil {
		log.Printf("Request to MongoDB")
		findOptions := options.Find()
		findOptions.SetSort(bson.D{{"parentID", 1},{"fullSlug", 1}})
		cur, err := handler.collection.Find(handler.ctx, bson.M{}, findOptions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cur.Close(handler.ctx)

		comments := make([]models.Comment, 0)
		for cur.Next(handler.ctx) {
			var comment models.Comment
			cur.Decode(&comment)
			comments = append(comments, comment)
		}

		data, _ := json.Marshal(comments)
		handler.redisClient.Set("comments", string(data), 0)
		c.JSON(http.StatusOK, comments)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		log.Printf("Request to Redis")
		comments := make([]models.Comment, 0)

		json.Unmarshal([]byte(val), &comments)
		c.JSON(http.StatusOK, comments)
	}
}

func generatePseudoRandomSlug() string {
	rand.Seed(time.Now().Unix())

	var output strings.Builder

	charSet := "abcdedfghijklmnopqrstABCDEFGHIJKLMNOP"
	lenght := 4
	for i := 0; i < lenght; i++ {
		random := rand.Intn(len(charSet))
		randomChar := charSet[random]
		output.WriteString(string(randomChar))
	}
	return output.String()
}

func (handler *CommentHandler) AddCommentHandler(c *gin.Context) {
	strID := c.Param("id")
	newsID, _ := strconv.Atoi(strID)
	var comment models.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := handler.collection.InsertOne(handler.ctx,
		bson.D{
		{"_id", primitive.NewObjectID()},
		{"newsID", newsID},
		{"slug", generatePseudoRandomSlug()},
		{"author", comment.Author},
		{"createdAt", time.Now()},
		{"text", comment.Text},
		{"deleted", comment.Delete},
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting a new comment"})
		return
	}

	log.Println("Remove data from Redis")

	handler.redisClient.Del("comments")

	c.JSON(http.StatusCreated, "Comment added")
}

func (handler *CommentHandler) AddResponeCommentHandler(c *gin.Context) {
	strID := c.Param("id")
	newsID, _ := strconv.Atoi(strID)
	timeNow := time.Now().Unix()
	posted := strconv.FormatInt(timeNow, 9)
	slugPart := generatePseudoRandomSlug()
	fullSlugPart := fmt.Sprintf("%s:%s", posted, slugPart)
	var commentSlug models.Comment
	var comment models.Comment
	var slug, fullSlug string
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	parentSlug := c.Param("slug")
	if parentSlug != ""  {
		parent := handler.collection.FindOne(handler.ctx, bson.M{
			"newsID": newsID, "slug": parentSlug})
		err := parent.Decode(&commentSlug)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "NewsID not found or slug not found"})
			return
}
		slug = fmt.Sprintf("%s:%s", commentSlug.Slug, slugPart)
		fmt.Println(commentSlug.FullSlug)
		fullSlug = fmt.Sprintf("%s:%s",commentSlug.FullSlug, fullSlugPart)
		fmt.Println(fullSlug)
	} else {
		slug = slugPart
		fullSlug = fullSlugPart
	}
	_, err := handler.collection.InsertOne(handler.ctx,
		bson.D{
		{"_id", primitive.NewObjectID()},
		{"newsID", newsID},
		{"slug", slug},
		{"fullSlug", fullSlug},
		{"parentID", commentSlug.ID},
		{"author", comment.Author},
		{"createdAt", time.Now()},
		{"text", comment.Text},
		{"deleted", comment.Delete},
		},
		)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting a new comment"})

		return
	}
	log.Println("Remove data from Redis")

	handler.redisClient.Del("comments")

	c.JSON(http.StatusCreated, "Comment added")
}


func (handler *CommentHandler) GetOneCommentHandler(c *gin.Context) {
	strId := c.Param("id")
	id, _ := strconv.Atoi(strId)
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"parentID", 1},{"createdAt", 1}})
	cur, err := handler.collection.Find(handler.ctx, bson.D{{"newsID", id}}, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "id not found"})
		return
	}
	comments := make([]models.Comment, 0)
	for cur.Next(handler.ctx) {
		var comment models.Comment
		cur.Decode(&comment)

		comments = append(comments, comment)
	}

	c.JSON(http.StatusOK, comments)
}
