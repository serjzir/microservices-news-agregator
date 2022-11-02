package storage


// Post публикация, полученная из RSS.
type Post struct {
	ID      int    `json:"id"`      // номер записи
	Title   string `json:"title"`   // заголовок публикации
	Content string `json:"content"` // содержание публикации
	PubTime int64  `json:"pubTime"` // время публикации
	Link    string `json:"link"`    // ссылка на источник
	CountPage 	int  `json:"countPage"`  // номер старницы
	Page 	int `json:"page"`
}
