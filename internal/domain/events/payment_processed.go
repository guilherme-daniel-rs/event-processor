package events

import "fmt"

type PaymentProcessedV1 struct {
	PaymentID     string  `json:"payment_id"`
	OrderID       string  `json:"order_id"`
	Amount        float64 `json:"amount"`
	PaymentMethod string  `json:"payment_method"`
	Status        string  `json:"status"`
}

func (e *PaymentProcessedV1) Validate() error {
	if e.PaymentID == "" {
		return fmt.Errorf("payment_id is required")
	}
	if e.OrderID == "" {
		return fmt.Errorf("order_id is required")
	}
	if e.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if e.PaymentMethod == "" {
		return fmt.Errorf("payment_method is required")
	}
	if e.Status == "" {
		return fmt.Errorf("status is required")
	}
	return nil
}
