package adminauth

import "time"

type Admin struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type adminCredentials struct {
	Admin
	PasswordHash string
}

type LoginResult struct {
	Admin        Admin
	SessionToken string
	ExpiresAt    time.Time
}
