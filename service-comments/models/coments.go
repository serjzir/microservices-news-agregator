package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)


type Comment struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	NewsID       int `json:"news_id" bson:"newsID"`
	ParentID     primitive.ObjectID `json:"parent_id" bson:"parentID"`
	Slug         string `json:"slug" bson:"slug"`
	FullSlug     string `json:"full_slug" bson:"fullSlug"`
	CreatedAt	 time.Time  `json:"created_at" bson:"createdAt"`
	Author       string `json:"author" bson:"author"`
	Text 		 string `json:"text" bson:"text"`
	Delete		 bool `json:"-" bson:"delete"`
}