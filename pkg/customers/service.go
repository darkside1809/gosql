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
var ErrPhoneUsed = errors.New("phone already registered")
var ErrTokenNotFound = errors.New("token not found")
var ErrTokenExpired = errors.New("token expired")
var ErrNoSuchUser = errors.New("no such user")
var ErrInvalidPassword = errors.New("invalid password")
var ErrNoRows = errors.New("no rows")

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

type Registration struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}
type Auth struct {
	Login 	string `json:"login"`
	Password string `json:"password"`
}
type Token struct {
	Token string `json:"token"`
}
type Products struct {
	ID 	int64  `json:"id"`
	Name 	string `json:"name"`
	Price int 	 `json:"price"`
	Qty 	int 	 `json:"qty"`
}

type Purchase struct {
	ID 			int64 `json:"id"`
	CustomerID	int	`json:"customer_id"`
	ManagerID	int	`json:"manager_id"`
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
// Register user and add him to a database
func (s *Service) Register(ctx context.Context, registration *Registration) (item *Customer, err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(registration.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrInternal
	}

	err = s.pool.QueryRow(ctx, `
		INSERT INTO customers(name, phone, password)
			VALUES($1, $2, $3)
			ON CONFLICT (phone) DO NOTHING RETURNING id, name, phone, active, created
			`, registration.Name, registration.Phone, hash ).Scan(
				&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created,
			)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, ErrInternal
	}
	return item, nil
}
// Token create token for user,
// if user is not found, return ErrNoSuchUser,
// if password is not found, return ErrInvalidPassword,
// if something else goes wrong, return ErrInternal.
func (s *Service) Token(ctx context.Context, phone string, password string) (token string, err error) {
	var hash string
	var id int64
	err = s.pool.QueryRow(ctx, `
		SELECT id, password FROM customers
			WHERE phone = $1`, phone).Scan(&id, &hash)
	if err == pgx.ErrNoRows {
		return "", ErrNotFound
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
	_, err = s.pool.Exec(ctx, `
		INSERT INTO customers_tokens(token, customer_id)
			VALUES($1, &2)`, token, id)
	if err != nil {
		return "", ErrInternal
	}
	return token, nil
}
// Get products 
func (s *Service) Products(ctx context.Context) ([]*Products, error) {
	items := make([]*Products, 0)
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, price, qty FROM products 
			WHERE active = true ORDER BY id LIMIT 500`)
	if errors.Is(err, pgx.ErrNoRows) {
		return items, err 
	}
	if err != nil {
		return nil, ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		item := &Products{}
		err = rows.Scan(&item.ID, &item.Price, &item.Qty)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		log.Print(err)
		return nil, err
	}

	return items, nil
}
// Find customer's id by his token
func (s *Service) IDByToken(ctx context.Context, token string) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		SELECT customer_id FROM customers_tokens WHERE token = $1`, token).Scan(&id)
	if err == pgx.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, ErrInternal
	}
	return id, nil
}
// Get purchases of customers or managers 
func (s *Service) Purchases(ctx context.Context, id int64) ([]*Purchase, error) {
	items := make([]*Purchase, 0)
	rows, err := s.pool.Query(ctx, `
	 SELECT manager_id, customer_id, FROM sales 
	 	WHERE customer_id = $1 ORDER BY id LIMIT 500`, id)
	
	if errors.Is(err, pgx.ErrNoRows) {
		return items, nil
	}
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		item := &Purchase{}
		err = rows.Scan(&item.ID, &item.ManagerID, &item.CustomerID)
		if err != nil {
			log.Print(err)
			return nil, ErrInternal
		}
		items = append(items, item)
	}

	err = rows.Err()
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return items, nil
}
// Make purchase
// func (s *Service) MakePurchases(ctx context.Context, id int64) ([]*Purchases, error) {
// 	var item *Customer
// 	err := s.pool.QueryRow(ctx, `
// 		INSERT INTO customers(name, phone) VALUES($1, $2) RETURNING id, name, phone, active, created
// 		`, item.Name, item.Phone).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
		
// 	if errors.Is(err, pgx.ErrNoRows) {
// 		log.Print("No rows")
// 		return nil, ErrNoRows
// 	}

// 	if err != nil {
// 		log.Print(err)
// 		return nil, ErrInternal
// 	}
	
// 	items := make([]*Purchases, 0)
// 	rows, err := s.pool.Query(ctx, `
// 		SELECT id, manager_id, customer_id FROM sales 
// 			WHERE customer_id = $1 ORDER BY id LIMIT 500 
// 	`,id)
// 	if errors.Is(err, pgx.ErrNoRows) {
// 		return items, nil
// 	}
// 	if err != nil {
// 		log.Print(err)
// 		return nil, ErrInternal
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		item := &Purchases{}
// 		err = rows.Scan(&item.ID, &item.ManagerID, &item.CustomerID)
// 		if err != nil {
// 			log.Print(err)
// 			return nil, err
// 		}
// 		items = append(items, item)
// 	}

// 	err = rows.Err()
// 	if err != nil {
// 		log.Print(err)
// 		return nil, err
// 	}
// 	return items, nil
// }