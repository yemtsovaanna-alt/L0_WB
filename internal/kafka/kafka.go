package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
)

type Handler interface {
	Handle(ctx context.Context, message []byte) error
}

type Kafka struct {
	subs    map[*kafka.Consumer]Handler
	connect string
	logger  *zap.Logger
}

func New(connect string, logger *zap.Logger) (*Kafka, error) {
	return &Kafka{
		connect: connect,
		logger:  logger,
		subs:    make(map[*kafka.Consumer]Handler),
	}, nil
}

func (k *Kafka) AddWorker(topic string, handler Handler) error {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": k.connect,
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return fmt.Errorf("cound not create a consumer: %s", err.Error())
	}
	err = c.Subscribe(topic, nil)
	if err != nil {
		return fmt.Errorf("cound not subscribe to a topic: %s", err.Error())
	}
	k.subs[c] = handler
	return nil
}

func (k *Kafka) Start(ctx context.Context) error {
	wg := sync.WaitGroup{}

	for sub, handler := range k.subs {
		wg.Add(1)
		sub, handler := sub, handler

		go func() {
			for {
				msg, err := sub.ReadMessage(time.Second)
				if errors.Is(err, context.Canceled) {
					wg.Done()
					return
				}
				if err == nil {
					fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
				} else if !err.(kafka.Error).IsTimeout() {
					fmt.Printf("Consumer error: %v (%v)\n", err, msg)
				}
				if msg == nil {
					continue
				}
				if err := handler.Handle(ctx, msg.Value); err != nil {
					k.logger.Error("cant handle message", zap.Error(err))
					// not commiting
					if err != nil {
						k.logger.Error("nak", zap.Error(err))
						continue
					}
				}
				_, err = sub.CommitMessage(msg)
				if err != nil {
					k.logger.Error("ack", zap.Error(err))
					time.Sleep(time.Second * 5)
				}
			}
		}()
	}

	wg.Wait()

	return nil
}

func (k *Kafka) Stop() {
	for sub := range k.subs {
		if sub.IsClosed() {
			_ = sub.Close()
		}
	}
	k.logger.Info("kafka loop stopped")
}
