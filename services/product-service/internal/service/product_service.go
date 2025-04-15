package service

import (
	"encoding/json"
	"time"

	"github.com/Shopify/sarama"
	"github.com/ecommerce-platform/product-service/internal/models"
	"github.com/ecommerce-platform/product-service/internal/repository"
)

type ProductService struct {
	repo     *repository.ProductRepository
	producer sarama.SyncProducer
	topic    string
}

func NewProductService(repo *repository.ProductRepository, producer sarama.SyncProducer, topic string) *ProductService {
	return &ProductService{
		repo:     repo,
		producer: producer,
		topic:    topic,
	}
}

func (s *ProductService) CreateProduct(product *models.Product) error {
	if err := s.repo.Create(product); err != nil {
		return err
	}

	// Publish event to Kafka
	event := models.ProductEvent{
		EventType: "product_created",
		Product:   *product,
		Timestamp: time.Now().Unix(),
	}

	return s.publishEvent(event)
}

func (s *ProductService) GetProduct(id uint) (*models.Product, error) {
	return s.repo.GetByID(id)
}

func (s *ProductService) UpdateProduct(product *models.Product) error {
	if err := s.repo.Update(product); err != nil {
		return err
	}

	// Publish event to Kafka
	event := models.ProductEvent{
		EventType: "product_updated",
		Product:   *product,
		Timestamp: time.Now().Unix(),
	}

	return s.publishEvent(event)
}

func (s *ProductService) DeleteProduct(id uint) error {
	product, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	// Publish event to Kafka
	event := models.ProductEvent{
		EventType: "product_deleted",
		Product:   *product,
		Timestamp: time.Now().Unix(),
	}

	return s.publishEvent(event)
}

func (s *ProductService) ListProducts() ([]models.Product, error) {
	return s.repo.List()
}

func (s *ProductService) publishEvent(event models.ProductEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: s.topic,
		Value: sarama.StringEncoder(value),
	}

	_, _, err = s.producer.SendMessage(msg)
	return err
}
