package messaging

import (
	"context"

	"github.com/ecommerce-platform/order-service/internal/models"
)

type OrderConsumer interface {
	Consume(ctx context.Context) error
	Close() error
}

type OrderCreatedHandler interface {
	HandleOrderCreated(ctx context.Context, order *models.Order) error
}
