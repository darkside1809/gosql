package managers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"time"
	"strconv"

	"github.com/darkside1809/gosql/cmd/app/middleware"
	"github.com/darkside1809/gosql/pkg/customers"
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

// Types
type Manager struct {
	ID      		int64     `json:"id"`
	Name    		string    `json:"name"`
	Phone   		string    `json:"phone"`
	Password 	string	 `json:"password"`
	Active  		bool      `json:"active"`
	Salary      int64     `json:"salary"`
	Plan        int64     `json:"plan"`
	BossID      int64     `json:"boss_id"`
	Departament string    `json:"departament"`
	IsAdmin     bool      `json:"is_admin"`
	Created     time.Time `json:"created"`
}
type Registration struct {
	Name     string 	`json:"name"`
	Phone    string 	`json:"phone"`
	Password string 	`json:"password"`
	Roles		[]string	`json:"roles"`
}
type Auth struct {
	Login 	string `json:"login"`
	Password string `json:"password"`
}
type Purchase struct {
	ID 			int64 `json:"id"`
	CustomerID	int	`json:"customer_id"`
	ManagerID	int	`json:"manager_id"`
}
type Product struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Price   int       `json:"price"`
	Qty     int       `json:"qty"`
	Active  bool      `json:"active"`
	Created time.Time `json:"created"`
}
type Sale struct {
	ID         int64           `json:"id"`
	ManagerID  int64           `json:"manager_id"`
	CustomerID int64           `json:"customer_id"`
	Created    time.Time       `json:"created"`
	Positions  []*SalesPosition `json:"positions"`
}
type SalesPosition struct {
	ID        int64 `json:"id"`
	ProductID int64 `json:"product_id"`
	Price     int   `json:"price"`
	Qty       int   `json:"qty"`
}
type SalesTotal struct {
	ManagerID int64 `json:"manager_id"`
	Total     int   `json:"total"`
}
func (s *Service) NewNullString(str string) sql.NullString{
	if len(str) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: str,
		Valid: true,
	}
}
func (s *Service) IsAdmin(ctx context.Context, id int64) (isAdmin bool) {
	err := s.pool.QueryRow(ctx, `SELECT is_admin FROM managers  WHERE id = $1`, id).Scan(&isAdmin)
	if errors.Is(err, pgx.ErrNoRows) {
		return false
	}
	if err != nil {
		return false
	}
	return
}
// Register user and add him to a database
func (s *Service) Register(ctx context.Context, manager *Manager) (string, error) {
	var id int64

	err := s.pool.QueryRow(ctx, `
		INSERT INTO managers(name, phone, is_admin)
			VALUES($1, $2, $3) ON CONFLICT (phone) DO NOTHING RETURNING id
			`, manager.Name, manager.Phone, manager.IsAdmin).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		log.Print("No rows")
		return "", ErrNoRows
	}

	if err != nil {
		log.Print(err)
		return "", ErrInternal
	}

	buffer := make([]byte, 256)
	n, err := rand.Read(buffer)
	if n != len(buffer) || err != nil {
		return "", ErrInternal
	}

	token := hex.EncodeToString(buffer)
	_, err = s.pool.Exec(ctx, `
		INSERT INTO managers_tokens(token, manager_id) VALUES($1, $2)`, token, id)
	if err != nil {
		log.Print(err)
		return "", ErrInternal
	}
	return token, nil
}
// Token create token for user,
// if user is not found, return ErrNoSuchUser,
// if password is not found, return ErrInvalidPassword,
// if something else goes wrong, return ErrInternal.
func (s *Service) Token(ctx context.Context, phone string, password string) (token string, err error) {
	var hash string
	var id int64

	err = s.pool.QueryRow(ctx, `SELECT id, password FROM managers WHERE phone = $1`, phone).Scan(&id, &hash)
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

	_, err = s.pool.Exec(ctx, `
		INSERT INTO managers_tokens(token, manager_id) VALUES($1, $2)`, token, id)
	if err != nil {
		log.Print(err)
		return "", ErrInternal
	}
	return token, nil
}
func (s *Service) IDByToken(ctx context.Context, token string) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
	 SELECT manager_id FROM managers_tokens WHERE token = $1
	 `, token).Scan(&id)

	if err == pgx.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, ErrInternal
	}
	return id, nil
}
func (s *Service) Purchases(ctx context.Context, id int64) ([]*Purchase, error) {
	items := make([]*Purchase, 0)
	rows, err := s.pool.Query(ctx, `
		SELECT id, manager_id, Manager_id FROM sales 
			WHERE Manager_id = $1 ORDER BY id LIMIT 500`, id)
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
		err = rows.Scan(&item.ID, &item.ManagerID, &item.ManagerID)
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
func (s *Service) Products(ctx context.Context) ([]*Product, error) {
	items := make([]*Product, 0)
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, price, qty FROM products 
			WHERE active ORDER BY id LIMIT 500 
	`)
	if errors.Is(err, pgx.ErrNoRows) {
		return items, nil
	}
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	defer rows.Close()
	for rows.Next() {
		item := &Product{}
		err = rows.Scan(&item.ID, &item.Name, &item.Price, &item.Qty)
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
func (s *Service) SaveProduct(ctx context.Context, product *Product) (*Product, error) {
	var err error
	if product.ID == 0 {
		err = s.pool.QueryRow(ctx, `
			INSERT INTO products(name, qty, price) 
				VALUES ($1, $2, $3) 
				RETURNING id, name, qty, price active, created;`, product.Name, product.Qty, product.Price).
			Scan(&product.ID, &product.Name, &product.Qty, &product.Price, &product.Active, &product.Created)

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRows
		}
	} 
	if product.ID != 0 {
		err = s.pool.QueryRow(ctx, `
			UPDATE products SET name = $1, qty = $2, price = $3  
				WHERE id = $4 
				RETURNING id, name, qty, price, active, created`, product.Name, product.Qty, product.Price, product.ID).
			Scan(&product.ID, &product.Name, &product.Qty, &product.Price, &product.Active, &product.Created)

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRows
		}
	}
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return product, nil
}
func (s *Service) ChangeProducts(ctx context.Context, product *Product) (*Product, error) {
	item := &Product{}

	if product.ID == 0 {
		err := s.pool.QueryRow(ctx, `
			INSERT INTO products(name, qty, price) 
				VALUES($1, $2, $3) RETURNING id, name, price, qty
		`, product.Name, product.Qty, product.Price).Scan(
			&item.ID, &item.Name, &item.Price, &item.Qty)

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRows
		}
		if err != nil {
			return nil, ErrInternal
		}
	}

	if product.ID != 0 {
		err := s.pool.QueryRow(ctx, `
			UPDATE products SET name = $2, qty = $3, price = $4 
				WHERE id = $1 RETURNING id, name, price, qty
		`, product.ID, product.Name, product.Qty, product.Price).Scan(
			&item.ID, &item.Name, &item.Price, &item.Qty)

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRows
		}
		if err != nil {
			return nil, ErrInternal
		}
	}
	return item, nil
}

func (s *Service) GetSales(ctx context.Context, id int64) (total int, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT COALESE(SUM(sp.qty * sp.price),0) total
			FROM managers m
		LEFT JOIN sales s on s.manager_id = $1
		LEFT JOIN sales_positions sp on sp.sale_id = s.id
		GROUP BY m.id
		LIMIT 1`, id).Scan(&total)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNoRows
	}
	if err != nil {
		log.Print(err)
		return 0, ErrInternal
	}
	return total, nil
}
func (s *Service) MakeSalePosition(ctx context.Context, position *SalesPosition) bool {
	active := false
	qty := 0
	err := s.pool.QueryRow(ctx, `
		SELECT qty, active FROM products WHERE id = $1`, position.ProductID).
		Scan(&qty, &active)
	if err != nil {
		return false
	}
	if qty < position.Qty || !active {
		return false
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE products SET qty = $1 WHERE id = $2`, qty-position.Qty, position.ProductID)
	if err != nil {
		log.Print(err)
		return false
	}

	return true
}
func (s *Service) MakeSale(ctx context.Context, sale *Sale) (*Sale, error) {
	sqlQuery := "INSERT INTO sales_positions (sale_id, product_id, qty, price) VALUES "
	sqlQuery2 := `INSERT INTO sales(manager_id,customer_id) 
		VALUES ($1,$2) RETURNING id, created;`

	err := s.pool.QueryRow(ctx, sqlQuery2, sale.ManagerID, sale.CustomerID).Scan(&sale.ID, &sale.Created)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	for _, position := range sale.Positions {
		if !s.MakeSalePosition(ctx, position) {
			log.Print("Invalid position")
			return nil, ErrInternal
		}
		sqlQuery += "(" + strconv.FormatInt(sale.ID, 10) + "," + strconv.FormatInt(position.ProductID, 10) + "," + strconv.Itoa(position.Price) + "," + strconv.Itoa(position.Qty) + "),"
	}

	sqlQuery = sqlQuery[0 : len(sqlQuery)-1]
	_, err = s.pool.Exec(ctx, sqlQuery)
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return sale, nil
}

func (s *Service) ManagerRole(ctx context.Context, roles ...string) bool {
	id, err := middleware.Authentication(ctx)
	if err != nil {
		log.Print(err)
		return false
	}

	err = s.pool.QueryRow(ctx, `
		SELECT roles FROM managers WHERE id = $1`, id).Scan(&roles)
	if err == pgx.ErrNoRows {
		return false
	}
	if err != nil {
		return false
	}
	for _, v := range roles {
		if v == "ADMIN" {
			return true
		}
	}
	return false
}
// Remove product by id
func (s *Service) RemoveProductByID(ctx context.Context, id int64) (err error) {
	_, err = s.pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		log.Print(err)
		return ErrInternal
	}
	return nil
}
// Remove customer by id
func (s *Service) RemoveCustomerByID(ctx context.Context, id int64) (err error) {
	_, err = s.pool.Exec(ctx, `DELETE FROM customers WHERE id = $1`, id)
	if err != nil {
		log.Print(err)
		return ErrInternal
	}
	return nil
}
// Get customers for managers stat
func (s *Service) GetCustomers(ctx context.Context) ([]*customers.Customer, error) {
	items := make([]*customers.Customer, 0)
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, phone, active, created 
			FROM customers WHERE active = true ORDER BY id LIMIT 500`)
	if err == pgx.ErrNoRows {
		return items, nil
	}
	if err != nil {
		return nil, ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		item := &customers.Customer{}
		err = rows.Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
// Change customer by manager
func (s *Service) ChangeCustomer(ctx context.Context, item *customers.Customer) (*customers.Customer, error) {
	err := s.pool.QueryRow(ctx, `
		UPDATE customers 
			SET name = $1, phone = $2, active = $3 WHERE id = $4
			RETURNING name, phone, active`, item.Name, item.Phone, item.Active, item.ID).Scan(
			&item.Name, &item.Phone, &item.Active)

	if errors.Is(err, pgx.ErrNoRows) {
		log.Print("No rows")
		return nil, ErrNoRows
	}
	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}			
	return item, nil
}