package persistent

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/yemtsovaanna-alt/L0_WB/internal/types"
)

type Database struct {
	pg *sql.DB
}

func New(driver string, connection string) (*Database, error) {
	db, err := sql.Open(driver, connection)
	if err != nil {
		return nil, fmt.Errorf("could not open db: %s", err.Error())
	}

	database := &Database{pg: db}
	return database, nil
}

// EnsureSchema creates the orders table with required columns if it doesn't exist.
func (d *Database) EnsureSchema(ctx context.Context) error {
	_, err := d.pg.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS orders (
            id text PRIMARY KEY,
            data bytea NOT NULL,
            created_at timestamptz NOT NULL DEFAULT now()
        );
    `)
	if err != nil {
		return fmt.Errorf("could not ensure schema: %s", err.Error())
	}
	return nil
}

func (d *Database) SaveOrUpdate(ctx context.Context, order types.Order, rawOrder []byte) error {
	queryString := `INSERT INTO orders (id, data) VALUES ($1, $2)
                    ON CONFLICT (id) DO UPDATE SET data = EXCLUDED.data, created_at = now()`
	values := []interface{}{order.Uid, rawOrder}

	_, err := d.pg.ExecContext(ctx, queryString, values...)
	if err != nil {
		return fmt.Errorf("could not execute save or update query: %s", err.Error())
	}
	return nil
}

func (d *Database) GetByID(ctx context.Context, id string) ([]byte, error) {
	var data []byte
	queryString := `SELECT data FROM orders WHERE id = $1`
	err := d.pg.QueryRowContext(ctx, queryString, id).Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("could not execute get by id query: %s", err.Error())
	}
	return data, nil
}

func (d *Database) GetAll(ctx context.Context) ([]types.Message, error) {
	var orders []types.Message
	dbOrder := new(types.Message)
	queryString := `SELECT id, data FROM orders`

	rows, err := d.pg.QueryContext(ctx, queryString)
	if err != nil {
		return nil, fmt.Errorf("could not execute get all query: %s", err.Error())
	}
	defer func(rows *sql.Rows) {
		rowsCloseErr := rows.Close()
		if rowsCloseErr != nil {
			err = fmt.Errorf("could not close rows: %s", rowsCloseErr.Error())
		}
	}(rows)

	for rows.Next() {
		err := rows.Scan(&dbOrder.Id, &dbOrder.Data)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *dbOrder)
	}
	return orders, nil
}

func (d *Database) GetRecent(ctx context.Context, limit int) ([]types.Message, error) {
	var orders []types.Message
	queryString := `SELECT id, data FROM orders ORDER BY created_at DESC LIMIT $1`
	rows, err := d.pg.QueryContext(ctx, queryString, limit)
	if err != nil {
		return nil, fmt.Errorf("could not execute get recent query: %s", err.Error())
	}
	defer func(rows *sql.Rows) {
		rowsCloseErr := rows.Close()
		if rowsCloseErr != nil {
			err = fmt.Errorf("could not close rows: %s", rowsCloseErr.Error())
		}
	}(rows)
	for rows.Next() {
		var m types.Message
		if err := rows.Scan(&m.Id, &m.Data); err != nil {
			return nil, err
		}
		orders = append(orders, m)
	}
	return orders, nil
}
