package router

import (
	"order-payment-kafka/order-createservice/handler"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, orderhandler *handler.OrderHandler) {
	r.POST("/orders", orderhandler.CreateOrder)
	r.GET("/orders/:id", orderhandler.GetOrderByID)
	r.GET("/users/:user-id/orders", orderhandler.GetOrdersByUserID)
	r.POST("/orders/:id/paid", orderhandler.PaidOrder)
	r.POST("/orders/:id/cancel", orderhandler.CancelOrder)
}
