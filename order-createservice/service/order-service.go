package service

import (
	"context"
	"encoding/json"
	"fmt"
	"order-payment-kafka/order-createservice/model"
	pb "order-payment-kafka/order-createservice/pb"
	"time"
)

type KafkaPublisher interface {
	Publish(topic string, key, value []byte) error
	Close() error
}

type OrderRepository interface {
	CreateOrder(order *model.Order) (*model.Order, error)
	GetOrderByID(id uint) (*model.Order, error)
	GetOrderByUserID(userID uint) ([]*model.Order, error)
}

type OrderService struct {
	db        OrderRepository
	publisher KafkaPublisher
	invClient pb.InventoryServiceClient
}

type OrderMessage struct {
	UserID   uint   `json:"user_id"`
	OrderID  uint   `json:"order_id"`
	ItemName string `json:"item_name"`
	Nums     int    `json:"nums"`
}

func NewOrderService(db OrderRepository, publisher KafkaPublisher, invClient pb.InventoryServiceClient) *OrderService {
	return &OrderService{db: db, publisher: publisher, invClient: invClient}
}

func (s *OrderService) CreateOrder(itemName string, nums int, userID uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	preDeDuctReq := &pb.PreDeductRequest{
		ProductId: itemName,
		Quantity:  int32(nums),
	}
	deductResp, err := s.invClient.PreDeduct(ctx, preDeDuctReq)
	if err != nil {
		return fmt.Errorf("failed to call PreDeduct: %v", err)
	}
	if !deductResp.Success {
		return fmt.Errorf("failed to deduct inventory: %s", deductResp.Message)
	}

	order := &model.Order{
		UserID:   userID,
		ItemName: itemName,
		Nums:     nums,
		Status:   "created",
	}
	savedOrder, err := s.db.CreateOrder(order)
	if err != nil {
		rollbackReq := &pb.RollbackDeductRequest{
			ProductId: itemName,
			Quantity:  int32(nums),
		}
		rollbackResp, rbErr := s.invClient.RollbackDeduct(ctx, rollbackReq)
		if rbErr != nil || !rollbackResp.Success {
			return fmt.Errorf("failed to rollback inventory after order creation failure: %v, rollback error: %v", err, rbErr)
		}
		return err
	}
	orderMessage := OrderMessage{
		OrderID:  savedOrder.ID,
		ItemName: savedOrder.ItemName,
		Nums:     savedOrder.Nums,
	}
	orderBytes, _ := json.Marshal(orderMessage)
	if err := s.publisher.Publish("orders-created", []byte(fmt.Sprintf("%d", savedOrder.ID)), orderBytes); err != nil {
		return err
	}
	return nil
}

func (s *OrderService) GetOrderByID(id uint) (*model.Order, error) {
	return s.db.GetOrderByID(id)
}

func (s *OrderService) GetOrdersByUserID(userID uint) ([]*model.Order, error) {
	return s.db.GetOrderByUserID(userID)
}
