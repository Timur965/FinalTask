package db

import (
	storage "FinalTask/Storage"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresNews struct {
	db *pgxpool.Pool
}

func NewNews(connection string) (*PostgresNews, error) {
	db, err := pgxpool.Connect(context.Background(), connection)
	if err != nil {
		return nil, err
	}
	s := PostgresNews{
		db: db,
	}
	return &s, nil
}

func (p *PostgresNews) GetAllNews() ([]storage.News, error) {
	rows, err := p.db.Query(context.Background(), `SELECT * FROM news;`)
	if err != nil {
		return nil, err
	}

	var allNews []storage.News
	for rows.Next() {
		var news storage.News
		var date time.Time
		err = rows.Scan(
			&news.Id,
			&news.Title,
			&news.Content,
			&date,
		)
		if err != nil {
			return nil, err
		}
		news.CreatedAt = date.Unix()
		allNews = append(allNews, news)

	}
	return allNews, rows.Err()
}

func (p *PostgresNews) GetCurrentNew(idNews int) (storage.News, error) {
	rows, err := p.db.Query(context.Background(), `SELECT * FROM news WHERE id = $1;`, idNews)
	if err != nil {
		return storage.News{}, err
	}
	defer rows.Close()

	var news storage.News
	var date time.Time

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return storage.News{}, fmt.Errorf("error during iteration: %w", err)
		}
		return storage.News{}, fmt.Errorf("not found")
	}

	err = rows.Scan(
		&news.Id,
		&news.Title,
		&news.Content,
		&date,
	)
	if err != nil {
		return storage.News{}, err
	}
	news.CreatedAt = date.Unix()

	if err := rows.Err(); err != nil {
		return storage.News{}, fmt.Errorf("error during rows scan: %w", err)
	}
	return news, nil
}

func (p *PostgresNews) GetFilterNews(typeFilter int, values ...interface{}) ([]storage.News, error) {

	var textFilter string
	var dateFilter int64
	var dateRangeStart, dateRangeEnd int64
	var typeSort int

	switch typeFilter {
	case storage.FullMatchText, storage.PartialMatchText, storage.FullMatchHeader,
		storage.PartialMatchHeader, storage.ExcludedPhrases:
		if len(values) < 1 {
			log.Printf("not enough arguments provided for text filter type %d", typeFilter)
		}
		if str, ok := values[0].(string); ok {
			textFilter = str
		} else {
			log.Printf("expected string for text filter, got %T", values[0])
		}

	case storage.SelectionDate:
		if len(values) < 1 {
			log.Printf("not enough arguments provided for date filter")
		}
		if date, ok := values[0].(int64); ok {
			dateFilter = date
		} else {
			log.Printf("expected string for date filter, got %T", values[0])
		}

	case storage.DateRange:
		if len(values) < 2 {
			log.Printf("not enough arguments provided for date range filter")
		}
		if start, ok := values[0].(int64); ok {
			dateRangeStart = start
		} else {
			log.Printf("expected string for start date, got %T", values[0])
		}
		if end, ok := values[1].(int64); ok {
			dateRangeEnd = end
		} else {
			log.Printf("expected string for end date, got %T", values[1])
		}
	case storage.FieldSort:
		if len(values) < 1 {
			log.Printf("not enough arguments provided for fieldSort filter")
		}
		if tmp, ok := values[0].(int); ok {
			typeSort = tmp
		} else {
			log.Printf("expected int for fieldSort filter, got %T", values[0])
		}
	}

	var query string
	var rows pgx.Rows
	var err error
	if textFilter != "" {
		switch typeFilter {
		case storage.FullMatchText:
			query = `SELECT * FROM news WHERE content = $1;`
		case storage.PartialMatchText:
			query = `SELECT * FROM news WHERE content ILIKE '%' || $1 || '%';`
		case storage.FullMatchHeader:
			query = `SELECT * FROM news WHERE title = $1;`
		case storage.PartialMatchHeader:
			query = `SELECT * FROM news WHERE title ILIKE '%' || $1 || '%';`
		case storage.ExcludedPhrases:
			query = `SELECT * FROM news WHERE content NOT ILIKE '%' || $1 || '%';`
		}

		rows, err = p.db.Query(context.Background(), query, textFilter)
		if err != nil {
			log.Printf("failed to execute query: %v", err)
		}

	} else if dateFilter != 0 && typeFilter == storage.DateRange {
		query = `SELECT * FROM news WHERE created_at::date = to_timestamp($1)::date;`

		rows, err = p.db.Query(context.Background(), query, dateFilter)
		if err != nil {
			log.Printf("failed to execute query: %v", err)
		}
	} else if dateRangeStart != 0 && dateRangeEnd != 0 && typeFilter == storage.SelectionDate {
		query = `SELECT * FROM news WHERE created_at BETWEEN to_timestamp($1) AND to_timestamp($2);`

		rows, err = p.db.Query(context.Background(), query, dateRangeStart, dateRangeEnd)
		if err != nil {
			log.Printf("failed to execute query: %v", err)
		}
	} else if typeFilter == storage.FieldSort {
		switch typeSort {
		case 0:
			query = `SELECT * FROM news ORDER BY created_at;`
		case 1:
			query = `SELECT * FROM news ORDER BY title;`
		}

		rows, err = p.db.Query(context.Background(), query)
		if err != nil {
			log.Printf("failed to execute query: %v", err)
		}
	}

	if query == "" {
		return nil, fmt.Errorf("empty query")
	}

	var allNews []storage.News
	for rows.Next() {
		var news storage.News
		var date time.Time
		err = rows.Scan(
			&news.Id,
			&news.Title,
			&news.Content,
			&date,
		)
		if err != nil {
			return nil, err
		}
		news.CreatedAt = date.Unix()
		allNews = append(allNews, news)

	}
	return allNews, rows.Err()

}

func (p *PostgresNews) AddNews(news []storage.News) error {
	tx, err := p.db.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		} else {
			tx.Commit(context.Background())
		}
	}()

	batch := &pgx.Batch{}
	for _, post := range news {
		batch.Queue(`
		INSERT INTO news(title, content, created_at)
		VALUES ($1, $2, to_timestamp($3))`,
			post.Title,
			post.Content,
			post.CreatedAt,
		)
	}

	results := tx.SendBatch(context.Background(), batch)
	if err := results.Close(); err != nil {
		return fmt.Errorf("failed to execute batch insert: %w", err)
	}

	return nil
}
