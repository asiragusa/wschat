package entity

import "time"

// User entity
type User struct {
	// User id
	Id string

	// User email
	Email string

	// User password
	Password string

	// Secret, used for the JWT Token ID. Changing this invalidates the tokens
	Secret string

	// Created At
	CreatedAt time.Time
}
