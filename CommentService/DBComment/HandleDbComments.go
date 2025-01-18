package DBComment

import (
	storage "FinalTask/Storage"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresComments struct {
	dbComment *pgxpool.Pool
}

func New(connection string) (*PostgresComments, error) {
	db, err := pgxpool.Connect(context.Background(), connection)
	if err != nil {
		return nil, err
	}

	var pc PostgresComments
	pc.dbComment = new(pgxpool.Pool)
	pc.dbComment = db

	return &pc, nil
}

func (pc *PostgresComments) AddComments(comment storage.Comments) error {
	query := `
        INSERT INTO comments (news_id, content, created_at)
        VALUES ($1, $2, to_timestamp($3))
        RETURNING id;
    `

	var newID int
	err := pc.dbComment.QueryRow(context.Background(), query, comment.NewsId, comment.Content, comment.CreatedAt).Scan(&newID)
	if err != nil {
		return fmt.Errorf("failed to insert comment: %w", err)
	}

	log.Printf("Comment added with ID: %d", newID)
	return nil
}

func (pc *PostgresComments) GetComments(idNews int) ([]storage.Comments, error) {
	rows, err := pc.dbComment.Query(context.Background(), `SELECT * FROM comments WHERE news_id = $1;`, idNews)
	if err != nil {
		return nil, err
	}

	var allComments []storage.Comments
	for rows.Next() {
		var comment storage.Comments
		var date time.Time
		err = rows.Scan(
			&comment.Id,
			&comment.NewsId,
			&comment.Content,
			&date,
		)
		if err != nil {
			return nil, err
		}
		comment.CreatedAt = date.Unix()
		allComments = append(allComments, comment)

	}
	return allComments, rows.Err()
}
