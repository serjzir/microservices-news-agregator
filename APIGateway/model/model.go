package model

import "time"



type NewsShortDetailed struct {
	ID      int    `json:"id"`      // номер записи
	Title   string `json:"title"`   // заголовок публикации
	Content string `json:"content"` // содержание публикации
	PubTime int64  `json:"pubTime"` // время публикации
	Link    string `json:"link"`    // ссылка на источник
}

type  NewsFullDetailed struct {
	ID      int    `json:"id"`      // номер записи
	Title   string `json:"title"`   // заголовок публикации
	Content string `json:"content"` // содержание публикации
	PubTime int64  `json:"pubTime"` // время публикации
	Link    string `json:"link"`    // ссылка на источник
	CountPage 	int  `json:"countPage"`  // номер старницы
	Page 	int `json:"page"`
	Comment []Comment `json:"comment"`
}

type Comment struct {
	ID          string `json:"id" bson:"_id"`
	NewsID       int `json:"news_id" bson:"newsID"`
	ParentID     string `json:"parent_id" bson:"parentID"`
	Slug         string `json:"slug" bson:"slug"`
	FullSlug     string `json:"full_slug" bson:"fullSlug"`
	CreatedAt	 time.Time  `json:"created_at" bson:"createdAt"`
	Author       string `json:"author" bson:"author"`
	Text 		 string `json:"text" bson:"text"`
	Delete		 bool `json:"-" bson:"delete"`
}
