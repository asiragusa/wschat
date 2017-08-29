package repository

import (
	"cloud.google.com/go/datastore"
	"context"
	"errors"
	"github.com/asiragusa/wschat/entity"
	"github.com/jonboulle/clockwork"
	"github.com/satori/go.uuid"
)

var (
	// Error thrown when the message has not been found
	MessageNotFoundError = errors.New("Not found")
)

// Interface used mainly for Unit testing
type MessageRepository interface {
	GetById(string) (*entity.Message, error)
	AllWithUser(string) ([]entity.Message, error)
	Create(string, string, string) (*entity.Message, error)
}

// Message Repository
type Message struct {
	// Injected via DI
	Client *datastore.Client `inject:""`

	// Injected via DI
	Clock clockwork.Clock `inject:""`
	kind  string
}

func NewMessageRepository() *Message {
	return &Message{
		kind: "Message",
	}
}

// Fetch a message by ID
func (r Message) GetById(id string) (*entity.Message, error) {
	key := datastore.NameKey(r.kind, id, nil)

	ctx := context.Background()

	entity := &entity.Message{}
	err := r.Client.Get(ctx, key, entity)
	if err == datastore.ErrNoSuchEntity {
		return nil, MessageNotFoundError
	}

	if err != nil {
		return nil, err
	}

	return entity, nil
}

// Fetch all messages belonging to an user
func (r Message) AllWithUser(email string) ([]entity.Message, error) {
	query := datastore.NewQuery(r.kind).Filter("Users =", email).Order("CreatedAt")

	entities := []entity.Message{}
	ctx := context.Background()
	_, err := r.Client.GetAll(ctx, query, &entities)

	if err != nil {
		return nil, err
	}

	return entities, nil

}

// Creates a new message
func (r Message) Create(from, to, message string) (*entity.Message, error) {
	entity := &entity.Message{
		Id:        uuid.NewV4().String(),
		From:      from,
		To:        to,
		Message:   message,
		Users:     []string{from, to},
		CreatedAt: r.Clock.Now(),
	}

	key := datastore.NameKey(r.kind, entity.Id, nil)

	ctx := context.Background()
	if _, err := r.Client.Put(ctx, key, entity); err != nil {
		return nil, err
	}

	return entity, nil
}
