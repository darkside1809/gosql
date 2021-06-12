package security

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) Auth(login string, password string) (ok bool) {
	ctx := context.Background()
	rows, err := s.pool.Query(ctx, 
		`SELECT id, name FROM managers 
			WHERE login = $1 AND password = $2`, login, password)
	defer rows.Close()

	if errors.Is(err, pgx.ErrNoRows) {
		log.Print("Rows not found")
		return false
	}
	if !rows.Next() {
		return false
	}
	if err != nil {
		log.Print(err)
		return false
	}
	return true
}