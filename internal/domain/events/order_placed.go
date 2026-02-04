package events

import "fmt"

type OrderPlacedV1 struct {
	OrderID    string  `json:"order_id"`
	UserID     string  `json:"user_id"`
	Total      float64 `json:"total"`
	ItemsCount int     `json:"items_count"`
	Status     string  `json:"status"`
}

func (e *OrderPlacedV1) Validate() error {
	if e.OrderID == "" {
		return fmt.Errorf("order_id is required")
	}
	if e.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if e.Total <= 0 {
		return fmt.Errorf("total must be greater than 0")
	}
	if e.ItemsCount <= 0 {
		return fmt.Errorf("items_count must be greater than 0")
	}
	if e.Status == "" {
		return fmt.Errorf("status is required")
	}
	return nil
}
