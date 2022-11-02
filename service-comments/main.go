package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/serjzir/service-comments/handlers"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var commentHandler *handlers.CommentHandler

func init() {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")
	collection := client.Database("comments").Collection("comments")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		})

	status := redisClient.Ping()
	log.Println(status)

	commentHandler = handlers.NewCommentHandler(ctx, collection, redisClient)
}

func main() {
	router := gin.Default()
	router.POST("/comments/:id", commentHandler.AddCommentHandler)
	router.POST("/comments/:id/:slug", commentHandler.AddResponeCommentHandler)
	router.GET("/comments", commentHandler.ListCommentHandler)
	router.GET("/comments/:id", commentHandler.GetOneCommentHandler)
	router.Run(":8888")
}