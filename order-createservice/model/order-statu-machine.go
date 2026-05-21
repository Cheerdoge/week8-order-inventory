package model

import "fmt"

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusCreated   OrderStatus = "created"
	StatusPaid      OrderStatus = "paid"
	StatusCancelled OrderStatus = "cancelled"
	StatusFailed    OrderStatus = "failed"
)

var allowedTransitions = map[OrderStatus][]OrderStatus{
	StatusPending:   {StatusCreated, StatusFailed},
	StatusCreated:   {StatusPaid, StatusCancelled, StatusFailed},
	StatusPaid:      {StatusCancelled},
	StatusCancelled: {},
	StatusFailed:    {},
}

func (o *Order) IsValidTransition(next OrderStatus) bool {
	allowed, exists := allowedTransitions[o.Status]
	if !exists {
		return false
	}
	for _, status := range allowed {
		if status == next {
			return true
		}
	}
	return false
}

func (o *Order) CanTransitionTo(next OrderStatus) error {
	if !o.IsValidTransition(next) {
		return fmt.Errorf("invalid transition from %s to %s", o.Status, next)
	}
	o.Status = next
	return nil
}
