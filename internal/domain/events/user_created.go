package events

import "fmt"

type UserCreatedV1 struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Verified bool   `json:"verified"`
}

func (e *UserCreatedV1) Validate() error {
	if e.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if e.Email == "" {
		return fmt.Errorf("email is required")
	}
	if e.Name == "" {
		return fmt.Errorf("name is required")
	}
	if e.Role == "" {
		return fmt.Errorf("role is required")
	}
	return nil
}
