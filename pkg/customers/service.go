package customers

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/pgx/v4"
)

// Errors
var ErrNotFound = errors.New("item not found")
var ErrInternal = errors.New("internal error")

// Types
type Service struct {
	pool	*pgxpool.Pool
}
type Customer struct {
	ID			int64			`json:"id"`
	Name		string		`json:"name"`
	Phone		string		`json:"phone"`
	Active	bool			`json:"active"`
	Created	time.Time	`json:"created"`
}
// Constructor of service
func NewService(pool	*pgxpool.Pool) *Service {
	return &Service{pool: pool}
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