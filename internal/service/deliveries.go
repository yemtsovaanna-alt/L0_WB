package deliveries

import (
	"context"
	"fmt"

	"github.com/yemtsovaanna-alt/L0_WB/internal/store/memory"
	"github.com/yemtsovaanna-alt/L0_WB/internal/store/persistent"
	"github.com/yemtsovaanna-alt/L0_WB/internal/types"
	"go.uber.org/zap"
)

type Memory interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte)
}

type Database interface {
	SaveOrUpdate(ctx context.Context, order types.Order, rawOrder []byte) error
	GetByID(ctx context.Context, id string) ([]byte, error)
	GetAll(ctx context.Context) ([]types.Message, error)
}

type Deliverer struct {
	store  Memory
	db     Database
	logger *zap.Logger
}

func New(store *memory.Store, database *persistent.Database, logger *zap.Logger) *Deliverer {
	return &Deliverer{
		store:  store,
		db:     database,
		logger: logger,
	}
}
func (d *Deliverer) SaveOrUpdate(ctx context.Context, order types.Order, rawOrder []byte) error {
	err := d.db.SaveOrUpdate(ctx, order, rawOrder)
	if err != nil {
		d.logger.Error("database", zap.Error(err))
		return fmt.Errorf("could not save or update order: %s", err.Error())
	}
	return nil
}

func (d *Deliverer) GetMessageById(id string) ([]byte, error) {
	if order, found := d.store.Get(id); found {
		return order, nil
	}
	raw, err := d.db.GetByID(context.Background(), id)
	if err != nil {
		d.logger.Error("message not found", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("message not found")
	}
	d.store.Set(id, raw)
	return raw, nil
}
