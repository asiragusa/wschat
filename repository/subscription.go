package repository

import (
	"cloud.google.com/go/datastore"
	"context"
	"errors"
	"github.com/asiragusa/wschat/entity"
	"github.com/jonboulle/clockwork"
)

var (
	// Error thrown if a subscription already exists
	SubscriptionAlreadyExistsError = errors.New("Already exists")
)

// Interface used mainly for Unit testing
type SubscriptionRepository interface {
	AllTo(string) ([]entity.Subscription, error)
	Create(string, string) (*entity.Subscription, error)
	Delete(string) error
}

// Subscription repository, used by services.PubsubClient
type Subscription struct {
	Client *datastore.Client `inject:""`
	Clock  clockwork.Clock   `inject:""`
	kind   string
}

func NewSubscriptionRepository() *Subscription {
	return &Subscription{
		kind: "Subscription",
	}
}

// Returns all the subscriptions belonging the "to" user
func (r Subscription) AllTo(to string) ([]entity.Subscription, error) {
	query := datastore.NewQuery(r.kind).Filter("To =", to)

	entities := []entity.Subscription{}
	ctx := context.Background()
	_, err := r.Client.GetAll(ctx, query, &entities)

	if err != nil {
		return nil, err
	}

	return entities, nil

}

// Creates a new subscription for the to user.
// The caller is responsible for the uniqueness of the ID (eg. using an UUID generator)
func (r Subscription) Create(id, to string) (*entity.Subscription, error) {
	subscription := &entity.Subscription{
		Id:        id,
		To:        to,
		CreatedAt: r.Clock.Now(),
	}

	key := datastore.NameKey(r.kind, subscription.Id, nil)

	ctx := context.Background()

	_, err := r.Client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		var existing entity.Subscription

		err := tx.Get(key, &existing)
		if err == nil {
			return SubscriptionAlreadyExistsError
		}
		if err != datastore.ErrNoSuchEntity {
			return err
		}

		_, err = tx.Put(key, subscription)
		return err
	})

	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// Deletes a subscription, by ID
func (r Subscription) Delete(id string) error {
	key := datastore.NameKey(r.kind, id, nil)

	ctx := context.Background()
	return r.Client.Delete(ctx, key)
}
