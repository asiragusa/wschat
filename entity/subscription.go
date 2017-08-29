package entity

import "time"

// The type Subscription indexes all the pubsub subscriptions.
//
// It allows fast searching for the subscriptions belonging to a given user.
// Used by services/pubsub_client
type Subscription struct {
	// Subscription Id
	Id string

	// User subscribed
	To string

	// Created At
	CreatedAt time.Time
}
