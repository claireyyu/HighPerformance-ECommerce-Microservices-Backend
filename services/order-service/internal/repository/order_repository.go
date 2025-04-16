package repository

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/ecommerce-platform/order-service/internal/config"
	"github.com/ecommerce-platform/order-service/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrInvalidOrder  = errors.New("invalid order")
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, id uint) (*models.Order, error)
	List(ctx context.Context, userID *uint) ([]models.Order, error)
	UpdateStatus(ctx context.Context, id uint, status models.OrderStatus) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(cfg *config.Config) (OrderRepository, error) {
	// Create DSN string
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
	)

	// Configure GORM logger
	gormLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logger.Config{
			LogLevel: logger.Info,
		},
	)

	// Open database connection
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate the schema
	if err := db.AutoMigrate(&models.Order{}, &models.OrderItem{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database schema: %w", err)
	}

	return &orderRepository{db: db}, nil
}

func (r *orderRepository) Create(ctx context.Context, order *models.Order) error {
	if order == nil {
		return ErrInvalidOrder
	}

	if order.UserID == 0 {
		return fmt.Errorf("%w: user ID is required", ErrInvalidOrder)
	}

	if len(order.Items) == 0 {
		return fmt.Errorf("%w: order must have at least one item", ErrInvalidOrder)
	}

	// Validate order items
	var totalAmount float64
	for _, item := range order.Items {
		if item.ProductID == 0 {
			return fmt.Errorf("%w: product ID is required", ErrInvalidOrder)
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("%w: quantity must be positive", ErrInvalidOrder)
		}
		if item.Price <= 0 {
			return fmt.Errorf("%w: price must be positive", ErrInvalidOrder)
		}
		totalAmount += float64(item.Quantity) * item.Price
	}

	// Validate total
	if order.TotalAmount <= 0 {
		order.TotalAmount = totalAmount
	} else if order.TotalAmount != totalAmount {
		return fmt.Errorf("%w: order total does not match sum of items", ErrInvalidOrder)
	}

	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepository) GetByID(ctx context.Context, id uint) (*models.Order, error) {
	if id == 0 {
		return nil, fmt.Errorf("%w: invalid order ID", ErrInvalidOrder)
	}

	var order models.Order
	err := r.db.WithContext(ctx).Preload("Items").First(&order, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return &order, nil
}

func (r *orderRepository) List(ctx context.Context, userID *uint) ([]models.Order, error) {
	var orders []models.Order
	query := r.db.WithContext(ctx).Preload("Items")

	if userID != nil {
		if *userID == 0 {
			return nil, fmt.Errorf("%w: invalid user ID", ErrInvalidOrder)
		}
		query = query.Where("user_id = ?", *userID)
	}

	if err := query.Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	return orders, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id uint, status models.OrderStatus) error {
	if id == 0 {
		return fmt.Errorf("%w: invalid order ID", ErrInvalidOrder)
	}

	// Check if order exists
	exists := true
	err := r.db.WithContext(ctx).Model(&models.Order{}).Select("1").First(&exists, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrderNotFound
		}
		return fmt.Errorf("failed to check order existence: %w", err)
	}

	// Update status
	result := r.db.WithContext(ctx).Model(&models.Order{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to update order status: %w", result.Error)
	}

	return nil
}
