package repository

import (
	"cloud.google.com/go/datastore"
	"context"
	"errors"
	"github.com/asiragusa/wschat/entity"
	"github.com/jonboulle/clockwork"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/iterator"
)

var (
	// The user does not exists
	UserNotFoundError = errors.New("User not found")
	// The fetched user is duplicated
	UserDuplicateError = errors.New("Found duplicate user")
	// The user already exists
	UserAlreadyExistsError = errors.New("User already exists")
	// Bad username or password
	UserBadUsernameOrPasswordError = errors.New("Bad username or password")
)

// Interface used mainly for Unit testing
type UserRepository interface {
	GetUserById(string) (*entity.User, error)
	GetUserByEmail(string) (*entity.User, error)
	CreateUser(string, string) (*entity.User, error)
	Login(string, string) (*entity.User, error)
	All() ([]entity.User, error)
}

// User Repository
type User struct {
	// Injected via DI
	Client *datastore.Client `inject:""`

	// Injected via DI
	Clock clockwork.Clock `inject:""`
	kind  string
}

func NewUserRepository() *User {
	return &User{
		kind: "User",
	}
}

// Fetches an user by ID
func (r User) GetUserById(id string) (*entity.User, error) {
	key := datastore.NameKey(r.kind, id, nil)

	ctx := context.Background()

	user := &entity.User{}
	err := r.Client.Get(ctx, key, user)
	if err == datastore.ErrNoSuchEntity {
		return nil, UserNotFoundError
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

// Fetches an user by Email
func (r User) GetUserByEmail(email string) (*entity.User, error) {
	query := datastore.NewQuery(r.kind).Filter("Email =", email)

	ctx := context.Background()
	it := r.Client.Run(ctx, query)

	var user *entity.User = nil
	for {
		var u entity.User
		_, err := it.Next(&u)
		if err == iterator.Done {
			break
		}
		if user != nil {
			return nil, UserDuplicateError
		}
		if err != nil {
			return nil, err
		}
		user = &u
	}

	if user == nil {
		return nil, UserNotFoundError
	}

	return user, nil
}

// Creates a new user, given its email and password
func (r User) CreateUser(email string, password string) (*entity.User, error) {
	_, err := r.GetUserByEmail(email)
	if err == nil {
		return nil, UserAlreadyExistsError
	}
	if err != UserNotFoundError {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		Id:        uuid.NewV4().String(),
		Email:     email,
		Password:  string(hash),
		Secret:    uuid.NewV4().String(),
		CreatedAt: r.Clock.Now(),
	}

	key := datastore.NameKey(r.kind, user.Id, nil)

	ctx := context.Background()
	_, err = r.Client.Put(ctx, key, user)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// Logs in an user by email and password
func (r User) Login(email, password string) (*entity.User, error) {
	user, err := r.GetUserByEmail(email)
	if err == UserNotFoundError {
		return nil, UserBadUsernameOrPasswordError
	}
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, UserBadUsernameOrPasswordError
	}

	return user, nil
}

// Fetches all the users in the DB
func (r User) All() ([]entity.User, error) {
	query := datastore.NewQuery(r.kind).Order("Email")

	users := []entity.User{}
	ctx := context.Background()
	_, err := r.Client.GetAll(ctx, query, &users)

	if err != nil {
		return nil, err
	}

	return users, nil
}
