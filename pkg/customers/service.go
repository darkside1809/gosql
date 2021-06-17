package customers

import (
	"context"
	"errors"
	"log"
	"time"
	"crypto/rand"
	"encoding/hex"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var ErrNotFound = errors.New("item not found")
var ErrInternal = errors.New("internal error")
var ErrNoSuchUser = errors.New("no such user")
var ErrInvalidPassword = errors.New("invalid password")

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type Customer struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Phone   string    `json:"phone"`
	Active  bool      `json:"active"`
	Created time.Time `json:"created"`
}

type CustomerAuth struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

// Get customers By Id
func (s *Service) ByID(ctx context.Context, id int64) (*Customer, error) {
	item := &Customer{}

	err := s.pool.QueryRow(ctx, `
		SELECT id, name, phone, active, created 
			FROM customers WHERE id = $1
	`, id).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return item, nil
}
// Get All customers
func (s *Service) All(ctx context.Context) ([]*Customer, error) {
	customers := []*Customer{}
	
	rows, err := s.pool.Query(ctx, `SELECT * FROM customers`)
	defer rows.Close()

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, ErrInternal
	}

	defer rows.Close()

	for rows.Next() {
		item := &Customer{}
		err := rows.Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
		if err != nil {
			log.Print(err)
		}
		customers = append(customers, item)
	}
	return customers, nil
}
// Get All active customers
func (s *Service) AllActive(ctx context.Context) ([]*Customer, error) {
	customers := []*Customer{}

	rows, err := s.pool.Query(ctx, `SELECT * FROM customers WHERE active = true`)
	defer rows.Close()

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, ErrInternal
	}

	for rows.Next() {
		item := &Customer{}
		err := rows.Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
		if err != nil {
			log.Print(err)
		}
		customers = append(customers, item)
	}
	return customers, nil
}
// Save customers By id
func (s *Service) Save(ctx context.Context, customer *Customer) (*Customer, error) {
	item := &Customer{}
	
	if customer.ID == 0 {
		err := s.pool.QueryRow(ctx, `
		INSERT INTO customers(name, phone) VALUES($1, $2) RETURNING id, name, phone, active, created
		`, customer.Name, customer.Phone).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
		if errors.Is(err, pgx.ErrNoRows) {
			log.Print("No rows")
			return nil, ErrNotFound
		}
		if err != nil {
			log.Print(err)
			return nil, ErrInternal
		}
	}
	
	if customer.ID != 0 {
		err := s.pool.QueryRow(ctx, `
		UPDATE customers SET name = $2, phone = $3 WHERE id = $1 RETURNING id, name, phone, active, created
		`, customer.ID, customer.Name, customer.Phone).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
		
		if errors.Is(err, pgx.ErrNoRows) {
			log.Print("No rows")
			return nil, ErrNotFound
		}
		if err != nil {
			log.Print(err)
			return nil, ErrInternal
		}
	}

	return item, nil	
}
// Delete customer by id
func (s *Service) RemoveByID(ctx context.Context, id int64) (*Customer, error) {
	item := &Customer{}

	err := s.pool.QueryRow(ctx, 
		`DELETE FROM customers
			WHERE id = $1 RETURNING *`, id).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return item, nil
}
// Block and Unblock customer By his id
func (s *Service) BlockAndUnblockByID(ctx context.Context, id int64, active bool) (*Customer, error) {
	item := &Customer{}

	err := s.pool.QueryRow(ctx, 
		`UPDATE customers SET active = $1 
			WHERE id = $2 RETURNING *`, active, id).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)

	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return item, nil
}
// Create token for customers
func (s *Service) TokenForCustomer(ctx context.Context, phone string, password string,) (token string, err error) {
	var hash string
	var id int64
	err = s.pool.QueryRow(ctx, `SELECT id,password From customers WHERE phone = $1`, phone).Scan(&id, &hash)

	if err == pgx.ErrNoRows {
		return "", ErrInvalidPassword
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
	_, err = s.pool.Exec(ctx, `INSERT INTO customers_tokens(token,customer_id) VALUES($1,$2)`, token, id)
	if err != nil {
		return "", ErrInternal
	}

	return token, nil
}