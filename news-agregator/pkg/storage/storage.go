package storage

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgconn"
	"github.com/serjzir/news-agregator/pkg/client/postgresql"
	"github.com/serjzir/news-agregator/pkg/logging"
)

type Repository struct {
	client postgresql.Client
	logger *logging.Logger
}

func (r *Repository) Create(ctx context.Context, posts []Post) error {
	for _, p := range posts {
		r.logger.Info("Вставка")
		q := "INSERT INTO news (title, content, pub_time, link) VALUES($1, $2, $3, $4) RETURNING id"
		r.logger.Trace(fmt.Sprintf("SQL Query: %s", q))
		if err := r.client.QueryRow(ctx, q, p.Title, p.Content, p.PubTime, p.Link).Scan(&p.ID); err != nil {
			var pgErr *pgconn.PgError
			if errors.Is(err, pgErr) {
				pgErr = err.(*pgconn.PgError)
				newErr := fmt.Errorf(fmt.Sprintf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQL State: %s",
					pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState()))
				r.logger.Error(newErr)
				return newErr
			}
			return err
		}
	}
	return nil
}

// News возвращает последние новости из БД.
func (r *Repository) News(c *gin.Context, id int) (p []Post, err error) {
	q := "SELECT id, title, content, pub_time, link FROM news WHERE id=$1"
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", q))
	rows, err := r.client.Query(c, q, id)
	if err != nil {
		return nil, err
	}
	var news []Post
	for rows.Next() {
		var new Post
		err = rows.Scan(&new.ID, &new.Title, &new.Content, &new.PubTime, &new.Link)
		if err != nil {
			return nil, err
		}
		news = append(news, new)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return news, nil
}

func (r *Repository) Searche(c *gin.Context, title, limitStr, pageStr string) (p []Post, err error) {
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "BadRequest"})
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "BadRequest"})
		return
	}
	if page == 0 || limit == 0 {
		page = 1
		limit = 10
	}

	offset := limit*page - limit
	if page == 1 {
		offset = 0
	}
	queryCount := "SELECT COUNT(id)/10 AS id FROM news WHERE title ILIKE $1"
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", queryCount))
	countPage, err := r.client.Query(c, queryCount, "%"+title+"%")
	var count int
	for countPage.Next() {
		err = countPage.Scan(&count)
		if err != nil {
			return nil, err
		}
	}
	q := "SELECT id, title, content, pub_time, link FROM news WHERE title ILIKE $1 ORDER BY pub_time ASC LIMIT $2 OFFSET $3"

	rows, err := r.client.Query(c, q, "%"+title+"%", limit, offset)
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", q))
	if err != nil {
		return nil, err
	}

	var news []Post
	for rows.Next() {
		var new Post
		err = rows.Scan(&new.ID, &new.Title, &new.Content, &new.PubTime, &new.Link)
		if err != nil {
			return nil, err
		}
		new.CountPage = count + 1
		new.Page = page
		news = append(news, new)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return news, nil
}
func (r *Repository) PaginationRequest(c *gin.Context, limitStr, pageStr string) (p []Post, err error) {
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "BadRequest."})
		return
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "BadRequest"})
		return
	}
	if page == 0 {
		page = 1
	}
	offset := limit*page - limit

	if page == 1 {
		offset = 0
	}

	queryCount := "SELECT COUNT(id)/10 AS id FROM news"
	r.logger.Trace(fmt.Sprintf("SQL Query: %s", queryCount))
	countPage, err := r.client.Query(c, queryCount)
	var count int
	for countPage.Next() {
		err = countPage.Scan(&count)
		if err != nil {
			return nil, err
		}
	}

	q := "SELECT id, title, content, pub_time, link FROM news ORDER BY id ASC LIMIT $1 OFFSET $2"

	rows, err := r.client.Query(c, q, limit, offset)
	r.logger.Trace(fmt.Sprintf("SQL Query: %s %d %d", q, limit, offset))
	if err != nil {
		return nil, err
	}

	var news []Post
	for rows.Next() {
		var new Post
		err = rows.Scan(&new.ID, &new.Title, &new.Content, &new.PubTime, &new.Link)

		if err != nil {
			return nil, err
		}
		new.CountPage = count + 1
		new.Page = page
		news = append(news, new)
	}
	news = append(news)
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return news, nil
}

func NewRepository(client postgresql.Client, logger *logging.Logger) *Repository {
	return &Repository{
		client: client,
		logger: logger,
	}
}
