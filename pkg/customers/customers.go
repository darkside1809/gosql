package customers

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"
)
// Errors
var ErrNotFound = errors.New("item not found")
var ErrInternal = errors.New("internal error")

// Types
type Service struct {
	db	*sql.DB
}
type Customer struct {
	ID			int64			`json:"id"`
	Name		string		`json:"name"`
	Phone		string		`json:"phone"`
	Active	bool			`json:"active"`
	Created	time.Time	`json:"created"`
}
// Constructor of service
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// Get customers By Id
func (s *Service) ByID(ctx context.Context, id int64) (*Customer, error) {
	item := &Customer{}

	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, phone, active, created 
			FROM customers WHERE id = $1
	`, id).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)

	if errors.Is(err, sql.ErrNoRows) {
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
	
	rows, err := s.db.QueryContext(ctx, `SELECT * FROM customers`)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, ErrInternal
	}
	
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			if err == nil {
				err = cerr
			}
		}
	}()

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

	rows, err := s.db.QueryContext(ctx, `SELECT * FROM customers WHERE active = true`)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, ErrInternal
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			if err == nil {
				err = cerr
			}
		}
	}()

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
func (s *Service) Save(ctx context.Context, customer *Customer) (c *Customer, err error) {
	item := &Customer{}

	if customer.ID == 0 {
		err = s.db.QueryRowContext(ctx,
			`INSERT INTO customers(name, phone) 
				VALUES($1, $2) RETURNING *`, customer.Name, customer.Phone).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
	} else {
		err = s.db.QueryRowContext(ctx, 
			`UPDATE customers SET name = $1, phone = $2
				 WHERE id = $3 RETURNING *`, customer.Name, customer.Phone, customer.ID).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, ErrInternal
	}
	return item, nil
} 	
// Delete customer by id
func (s *Service) RemoveByID(ctx context.Context, id int64) (*Customer, error) {
	item := &Customer{}

	err := s.db.QueryRowContext(ctx, 
		`DELETE FROM customers
			WHERE id = $1 RETURNING *`, id).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
	if err == sql.ErrNoRows {
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

	err := s.db.QueryRowContext(ctx, 
		`UPDATE customers SET active = $1 
			WHERE id = $2 RETURNING *`, active, id).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return item, nil
}