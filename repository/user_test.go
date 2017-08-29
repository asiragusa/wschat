package repository

import (
	"cloud.google.com/go/datastore"
	"context"
	"github.com/asiragusa/wschat/entity"
	"github.com/jonboulle/clockwork"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	userRepository *User
	clock          clockwork.FakeClock
}

func TestUserRepository(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func getDatastoreClient(projectID string) (*datastore.Client, error) {
	ctx := context.Background()

	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (suite *UserRepositoryTestSuite) SetupSuite() {
	client, err := getDatastoreClient("test")
	suite.Require().NoError(err)

	suite.userRepository = NewUserRepository()
	suite.userRepository.Client = client
}

func (suite *UserRepositoryTestSuite) cleanDb() {
	query := datastore.NewQuery("").KeysOnly()
	ctx := context.Background()

	keys, err := suite.userRepository.Client.GetAll(ctx, query, nil)
	suite.Require().NoError(err)

	err = suite.userRepository.Client.DeleteMulti(ctx, keys)
	suite.Require().NoError(err)
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	suite.cleanDb()

	suite.clock = clockwork.NewFakeClockAt(time.Now())
	suite.userRepository.Clock = suite.clock
}

func (suite *UserRepositoryTestSuite) createUser(email, password string) *entity.User {
	user, err := suite.userRepository.CreateUser(email, password)
	suite.Require().NoError(err)
	suite.Require().NotNil(user)

	return user
}

func (suite *UserRepositoryTestSuite) createUserRaw(email, password string) *entity.User {
	user := entity.User{
		Id:       uuid.NewV4().String(),
		Email:    email,
		Password: "",
		Secret:   uuid.NewV4().String(),
	}

	key := datastore.NameKey("User", user.Id, nil)

	ctx := context.Background()
	_, err := suite.userRepository.Client.Put(ctx, key, &user)

	suite.Require().NoError(err)

	return &user
}

var (
	email    = "test@test.com"
	password = "testPassword"
)

func (suite *UserRepositoryTestSuite) TestGetUserByIdNotExisting() {
	user, err := suite.userRepository.GetUserById("notExisting")
	suite.Nil(user)
	suite.EqualError(err, UserNotFoundError.Error())
}

func (suite *UserRepositoryTestSuite) TestGetUserByIdOK() {
	u := suite.createUser(email, password)

	user, err := suite.userRepository.GetUserById(u.Id)
	suite.NoError(err)
	suite.Require().NotNil(user)
	suite.Equal(email, user.Email)
}

func (suite *UserRepositoryTestSuite) TestGetUserByEmailNotExisting() {
	user, err := suite.userRepository.GetUserByEmail(email)
	suite.Nil(user)
	suite.EqualError(err, UserNotFoundError.Error())
}

func (suite *UserRepositoryTestSuite) TestGetUserByEmailDuplicate() {
	suite.createUser(email, password)
	suite.createUserRaw(email, password)

	user, err := suite.userRepository.GetUserByEmail(email)
	suite.Nil(user)
	suite.EqualError(err, UserDuplicateError.Error())
}

func (suite *UserRepositoryTestSuite) TestGetUserByEmailOK() {
	suite.createUser(email, password)

	user, err := suite.userRepository.GetUserByEmail(email)
	suite.NoError(err)
	suite.Require().NotNil(user)
	suite.Equal(email, user.Email)
}

func (suite *UserRepositoryTestSuite) TestCreateExistingUser() {
	suite.createUser(email, password)

	user, err := suite.userRepository.CreateUser(email, password)
	suite.Nil(user)
	suite.EqualError(err, UserAlreadyExistsError.Error())
}

func (suite *UserRepositoryTestSuite) TestCreateUserOK() {
	user := suite.createUser(email, password)

	suite.Equal(user.Email, email)
	suite.Equal(user.CreatedAt, suite.userRepository.Clock.Now())
	suite.NotEmpty(user.Password)
	suite.NotEmpty(user.Secret)
}

func (suite *UserRepositoryTestSuite) TestLoginNotExistingUser() {
	user, err := suite.userRepository.Login(email, password)
	suite.Nil(user)
	suite.EqualError(err, UserBadUsernameOrPasswordError.Error())
}

func (suite *UserRepositoryTestSuite) TestLoginBadPassword() {
	suite.createUser(email, password)

	user, err := suite.userRepository.Login(email, "badPassword")
	suite.Nil(user)
	suite.EqualError(err, UserBadUsernameOrPasswordError.Error())
}

func (suite *UserRepositoryTestSuite) TestLoginOk() {
	suite.createUser(email, password)

	user, err := suite.userRepository.Login(email, password)
	suite.NoError(err)
	suite.Require().NotNil(user)

	suite.Equal(email, user.Email)
}

func (suite *UserRepositoryTestSuite) TestAllOk() {
	email1 := "b@b.com"
	suite.createUser(email1, password)

	email2 := "a@b.com"
	suite.createUser(email2, password)

	users, err := suite.userRepository.All()
	suite.NoError(err)
	suite.Require().NotNil(users)
	suite.Require().Len(users, 2)

	suite.Equal(email2, users[0].Email)
	suite.Equal(email1, users[1].Email)
}
