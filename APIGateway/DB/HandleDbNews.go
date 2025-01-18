package db

import (
	storage "FinalTask/Storage"
	"context"
	"fmt"
	"log"

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
		err = rows.Scan(
			&news.Id,
			&news.Title,
			&news.Content,
			&news.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		allNews = append(allNews, news)

	}
	return allNews, rows.Err()
}

func (p *PostgresNews) GetCurrentNew(idNews int) (storage.News, error) {
	rows, err := p.db.Query(context.Background(), `SELECT * FROM news WHERE id = $1;`, idNews)
	if err != nil {
		return storage.News{}, err
	}

	var news storage.News
	err = rows.Scan(
		&news.Id,
		&news.Title,
		&news.Content,
		&news.CreatedAt,
	)
	if err != nil {
		return storage.News{}, err
	}
	return news, rows.Err()
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
			query = `SELECT * FROM news WHERE content LIKE '%' || $1 || '%';`
		case storage.FullMatchHeader:
			query = `SELECT * FROM news WHERE title = $1;`
		case storage.PartialMatchHeader:
			query = `SELECT * FROM news WHERE title LIKE '%' || $1 || '%';`
		case storage.ExcludedPhrases:
			query = `SELECT * FROM news WHERE content NOT LIKE '%' || $1 || '%';`
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
		err = rows.Scan(
			&news.Id,
			&news.Title,
			&news.Content,
			&news.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		allNews = append(allNews, news)

	}
	return allNews, rows.Err()

}
