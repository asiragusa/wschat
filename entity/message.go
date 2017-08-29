// The entity package contains all the entities of google cloud datastore
package entity

import "time"

// Sent message struct
type Message struct {
	// Message Id
	Id string `json:"id"`

	// Users belonging to the message, used for searching
	Users []string

	// Sender
	From string `json:"from"`

	// Receiver
	To string `json:"to"`

	// Message text
	Message string `json:"message"`

	// Created at
	CreatedAt time.Time `json:"createdAt"`
}
