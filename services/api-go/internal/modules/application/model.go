package application

import "time"

type Application struct {
	ID           string    `json:"id"`
	PositionCode string    `json:"position_code"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	Experience   string    `json:"experience"`
	SubmittedAt  time.Time `json:"submitted_at"`
}

type CreateInput struct {
	FullName     string `json:"full_name"`
	Email        string `json:"email"`
	PositionCode string `json:"position_code"`
	Experience   string `json:"experience"`
}
