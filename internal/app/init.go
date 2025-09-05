package app

import (
	"context"
	"fmt"

	"github.com/yemtsovaanna-alt/L0_WB/internal/configs"
	"github.com/yemtsovaanna-alt/L0_WB/internal/http"
	worker "github.com/yemtsovaanna-alt/L0_WB/internal/kafka"
	deliveries "github.com/yemtsovaanna-alt/L0_WB/internal/service"
	"github.com/yemtsovaanna-alt/L0_WB/internal/store/memory"
	"github.com/yemtsovaanna-alt/L0_WB/internal/store/persistent"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type App struct {
	kafka *worker.Kafka
	http  *http.Server
}

func Initialize(ctx context.Context) (*App, error) {
	dbConfig, err := configs.NewConfigDB()
	if err != nil {
		return nil, err
	}
	kafkaConfig, err := configs.NewConfigKafka()
	if err != nil {
		return nil, err
	}
	cacheConfig, err := configs.NewConfigCache()
	if err != nil {
		return nil, err
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("could not create new logger: %s", err.Error())
	}
	store := memory.New(cacheConfig.Size)
	dbConnectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
		dbConfig.ModeSSL)
	database, err := persistent.New("postgres", dbConnectionString)
	if err != nil {
		return nil, fmt.Errorf("could not create new database: %s", err.Error())
	}
	if err := database.EnsureSchema(ctx); err != nil {
		return nil, err
	}

	// Optional cache warm-up: most recent N orders
	if cacheConfig.PreloadLimit > 0 {
		recent, err := database.GetRecent(ctx, cacheConfig.PreloadLimit)
		if err != nil {
			logger.Error("could not preload recent orders", zap.Error(err))
		} else {
			for _, m := range recent {
				store.Set(m.Id, m.Data)
			}
			logger.Info("preloaded recent orders", zap.Int("count", len(recent)))
		}
	}

	// Build Kafka bootstrap connection string from env-configured host and port
	kafkaConnectionString := fmt.Sprintf("%s:%s", kafkaConfig.Host, kafkaConfig.Port)
	connection := kafkaConnectionString
	if err != nil {
		return nil, fmt.Errorf("connect: %s", err.Error())
	}

	storeService := deliveries.New(store, database, logger)
	newWorker, err := worker.New(connection, logger)
	if err != nil {
		return nil, fmt.Errorf("could not create new worker: %s", err.Error())
	}

	ordersHandler := worker.NewOrdersHandler(storeService, logger)
	err = newWorker.AddWorker("orders", ordersHandler)
	if err != nil {
		return nil, err
	}

	httpServer := http.New(storeService, logger)

	return &App{
		kafka: newWorker,
		http:  httpServer,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	errGroup, ctx := errgroup.WithContext(ctx)
	errGroup.Go(func() error {
		return a.kafka.Start(ctx)
	})
	errGroup.Go(func() error {
		return a.http.Start(ctx)
	})
	errGroup.Go(func() error {
		<-ctx.Done()
		a.kafka.Stop()
		return a.http.Stop(ctx)
	})

	return errGroup.Wait()
}
