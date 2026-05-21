package handler

import (
	"fmt"
	"log"
	"net/http"
	"order-payment-kafka/order-createservice/model"

	"github.com/gin-gonic/gin"
)

type CreateOrderRequest struct {
	UserID   uint   `json:"user_id"`
	ItemName string `json:"item_name"`
	Nums     int    `json:"nums"`
}

type OrderChangeRequest struct {
	Paymentmethod string `json:"payment_method"`
}

type OrderService interface {
	CreateOrder(itemName string, nums int, userID uint) error
	GetOrderByID(id uint) (*model.Order, error)
	GetOrdersByUserID(userID uint) ([]*model.Order, error)
	PaidOrder(orderID uint) error
	CancelOrder(orderID uint) error
}

type OrderHandler struct {
	service OrderService
}

func NewOrderHandler(service OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Printf("Failed to bind request: %v", err)
		return
	}
	err := h.service.CreateOrder(req.ItemName, req.Nums, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("Failed to create order: %v", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Order created successfully"})

}

func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	orderIDstr := c.Param("id")
	var orderID uint
	_, err := fmt.Sscanf(orderIDstr, "%d", &orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		log.Printf("Failed to parse order ID: %v", err)
		return
	}
	order, err := h.service.GetOrderByID(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		log.Printf("Failed to get order: %v", err)
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) GetOrdersByUserID(c *gin.Context) {
	userIDstr := c.Param("user-id")
	var userID uint
	_, err := fmt.Sscanf(userIDstr, "%d", &userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		log.Printf("Failed to parse user ID: %v", err)
		return
	}
	orders, err := h.service.GetOrdersByUserID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Orders not found"})
		log.Printf("Failed to get orders: %v", err)
		return
	}
	orders, err = h.service.GetOrdersByUserID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Orders not found"})
		log.Printf("Failed to get orders: %v", err)
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) PaidOrder(c *gin.Context) {
	var req OrderChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Printf("Failed to bind request: %v", err)
		return
	}
	orderIDstr := c.Param("id")
	var orderID uint
	_, err := fmt.Sscanf(orderIDstr, "%d", &orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		log.Printf("Failed to parse order ID: %v", err)
		return
	}
	err = h.service.PaidOrder(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("Failed to mark order as paid: %v", err)
		return
	}
	log.Printf("Order %d marked as paid successfully, payment method: %s", orderID, req.Paymentmethod)
	c.JSON(http.StatusOK, gin.H{"message": "Order marked as paid successfully"})
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderIDstr := c.Param("id")
	var orderID uint
	_, err := fmt.Sscanf(orderIDstr, "%d", &orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		log.Printf("Failed to parse order ID: %v", err)
		return
	}
	err = h.service.CancelOrder(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("Failed to cancel order: %v", err)
		return
	}
	log.Printf("Order %d cancelled successfully", orderID)
	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled successfully"})
}
