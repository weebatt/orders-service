package repository

import (
	"context"
	"database/sql"
	_ "fmt"

	test "310499-itmobatareyka-course-1343/pkg/api/test/proto/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Инициализация таблицы (создаём при запуске, если ещё не создана).
func (r *PostgresRepository) InitSchema(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS orders (
	    id TEXT PRIMARY KEY,
	    item TEXT NOT NULL,
	    quantity INT NOT NULL
	);
	`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *PostgresRepository) Create(order *test.Order) error {
	_, err := r.db.Exec(`
		INSERT INTO orders (id, item, quantity)
		VALUES ($1, $2, $3)
	`, order.Id, order.Item, order.Quantity)

	if err != nil {
		return status.Errorf(codes.Internal, "failed to create order in DB: %v", err)
	}

	return nil
}

func (r *PostgresRepository) Find(id string) (*test.Order, error) {
	row := r.db.QueryRow(`
		SELECT id, item, quantity
		FROM orders
		WHERE id = $1
	`, id)

	order := &test.Order{}
	if err := row.Scan(&order.Id, &order.Item, &order.Quantity); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "order not found: Find")
		}
		return nil, status.Errorf(codes.Internal, "failed to find order in DB: %v", err)
	}
	return order, nil
}

func (r *PostgresRepository) Update(order *test.Order) (*test.Order, error) {
	res, err := r.db.Exec(`
		UPDATE orders
		SET item = $1, quantity = $2
		WHERE id = $3
	`, order.Item, order.Quantity, order.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update order in DB: %v", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "order not found: Update")
	}
	return order, nil
}

func (r *PostgresRepository) Delete(id string) (bool, error) {
	res, err := r.db.Exec(`
		DELETE FROM orders
		WHERE id = $1
	`, id)
	if err != nil {
		return false, status.Errorf(codes.Internal, "failed to delete order in DB: %v", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return false, status.Errorf(codes.NotFound, "order not found: Delete")
	}
	return true, nil
}

func (r *PostgresRepository) GetAllOrders() ([]*test.Order, error) {
	rows, err := r.db.Query(`
		SELECT id, item, quantity
		FROM orders
	`)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to select all orders in DB: %v", err)
	}
	defer rows.Close()

	var orders []*test.Order
	for rows.Next() {
		var order test.Order
		if err = rows.Scan(&order.Id, &order.Item, &order.Quantity); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to read order from DB: %v", err)
		}
		orders = append(orders, &order)
	}
	return orders, nil
}
