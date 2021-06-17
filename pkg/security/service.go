package security

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var ErrNotFound = errors.New("item not found")
var ErrExpired = errors.New("token is expired")
var ErrInternal = errors.New("internal error")
var ErrNoSuchUser = errors.New("no such user")
var ErrInvalidPassword = errors.New("invalid password")
var (
	ErrStatusNotFound int64 = 404
	ErrBadRequest int64 = 400
	StatusOk int64 = 200
)

type Service struct {
	pool *pgxpool.Pool
}
type Token struct {
	Token string `json:"token"`
}

type Responce struct {
	CustomerID int64  `json:"customerId"`
	Status     string `json:"status"`
	Reason     string `json:"reason"`
}

type ResponceOk struct {
	Status     string `json:"status"`
	CustomerID int64  `json:"customerId"`
}

type ResponceFail struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}
type Auth struct {
	Login    string `json:"login"`
	Password string `json:"password"`
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

func (s *Service) AuthenticateCustomer(ctx context.Context, token string,) (id int64, err error) {
	expiredTime := time.Now()
	nowTimeInSec := expiredTime.UnixNano()
	err = s.pool.QueryRow(ctx, `SELECT customer_id, expire FROM customers_tokens WHERE token = $1`, token).Scan(&id, &expiredTime)
	if err != nil {
		log.Print(err)
		return 0, ErrNoSuchUser
	}

	if nowTimeInSec > expiredTime.UnixNano() {
		return -1, ErrExpired
	}
	return id, nil
}

func (s *Service) TokenForCustomer(ctx context.Context, phone string, password string) (token string, err error) {
	var hash string
	var id int64
	err = s.pool.QueryRow(ctx, 
		`SELECT id, password FROM customers WHERE phone = $1
	`, phone).Scan(&password)
	
	if err == pgx.ErrNoRows {
		return "", ErrNoSuchUser
	}
	if err != nil {
		return "", ErrInternal
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return "", ErrInvalidPassword
	}

	buffer := make([]byte, 256)
	n, err := rand.Read(buffer)
	if n != len(buffer) || err != nil {
		return "", ErrInternal
	}

	token = hex.EncodeToString(buffer)
	_, err = s.pool.Exec(ctx, `INSERT INTO customers_tokens(token, customer_id) VALUES($1, $2)`, token, id)
	if err != nil {
		return "", ErrInternal
	}
	return token, nil
}