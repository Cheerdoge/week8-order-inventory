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
	UpdateOrder(order *model.Order) error
}

type OrderService struct {
	repo      OrderRepository
	publisher KafkaPublisher
	invClient pb.InventoryServiceClient
}

type OrderMessage struct {
	UserID   uint   `json:"user_id"`
	OrderID  uint   `json:"order_id"`
	ItemName string `json:"item_name"`
	Nums     int    `json:"nums"`
}

func NewOrderService(repo OrderRepository, publisher KafkaPublisher, invClient pb.InventoryServiceClient) *OrderService {
	return &OrderService{repo: repo, publisher: publisher, invClient: invClient}
}

func (s *OrderService) CreateOrder(itemName string, nums int, userID uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	order := &model.Order{
		UserID:   userID,
		ItemName: itemName,
		Nums:     nums,
		Status:   model.StatusCreated,
	}

	preDeDuctReq := &pb.PreDeductRequest{
		ProductId: itemName,
		Quantity:  int32(nums),
	}
	deductResp, err := s.invClient.PreDeduct(ctx, preDeDuctReq)
	if err != nil {
		order.CanTransitionTo(model.StatusFailed)
		//log.Printf("PreDeduct failed: %v", err)
		return fmt.Errorf("failed to call PreDeduct: %v", err)
	}
	if !deductResp.Success {
		order.CanTransitionTo(model.StatusFailed)
		//log.Printf("PreDeduct failed: %s", deductResp.Message)
		return fmt.Errorf("failed to deduct inventory: %s", deductResp.Message)
	}

	order.CanTransitionTo(model.StatusCreated)

	savedOrder, err := s.repo.CreateOrder(order)
	if err != nil {
		order.CanTransitionTo(model.StatusFailed)
		//log.Printf("Failed to create order: %v", err)
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
	return s.repo.GetOrderByID(id)
}

func (s *OrderService) GetOrdersByUserID(userID uint) ([]*model.Order, error) {
	return s.repo.GetOrderByUserID(userID)
}

func (s *OrderService) CancelOrder(orderID uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	order, err := s.repo.GetOrderByID(orderID)
	if err != nil {
		return err
	}
	if err := order.CanTransitionTo(model.StatusCancelled); err != nil {
		return err
	}
	rollbackReq := &pb.RollbackDeductRequest{
		ProductId: order.ItemName,
		Quantity:  int32(order.Nums),
	}
	rollbackResp, rbErr := s.invClient.RollbackDeduct(ctx, rollbackReq)
	if rbErr != nil || !rollbackResp.Success {
		return fmt.Errorf("failed to rollback inventory after order creation failure: %v, rollback error: %v", err, rbErr)
	}
	if err := s.repo.UpdateOrder(order); err != nil {
		return err
	}
	return nil
}

func (s *OrderService) PaidOrder(orderID uint) error {
	order, err := s.repo.GetOrderByID(orderID)
	if err != nil {
		return err
	}
	if err := order.CanTransitionTo(model.StatusPaid); err != nil {
		return err
	}
	if err := s.repo.UpdateOrder(order); err != nil {
		return err
	}
	return nil
}
