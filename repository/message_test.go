package repository

import (
	"cloud.google.com/go/datastore"
	"context"
	"github.com/asiragusa/wschat/entity"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type MessageRepositoryTestSuite struct {
	suite.Suite
	repository *Message
	clock      clockwork.FakeClock
}

func TestMessageRepository(t *testing.T) {
	suite.Run(t, new(MessageRepositoryTestSuite))
}

func (suite *MessageRepositoryTestSuite) SetupSuite() {
	client, err := getDatastoreClient("test")
	suite.Require().NoError(err)

	suite.repository = NewMessageRepository()
	suite.repository.Client = client
}

func (suite *MessageRepositoryTestSuite) cleanDb() {
	query := datastore.NewQuery("").KeysOnly()
	ctx := context.Background()

	keys, err := suite.repository.Client.GetAll(ctx, query, nil)
	suite.Require().NoError(err)

	err = suite.repository.Client.DeleteMulti(ctx, keys)
	suite.Require().NoError(err)
}

func (suite *MessageRepositoryTestSuite) SetupTest() {
	suite.cleanDb()

	suite.clock = clockwork.NewFakeClockAt(time.Now())
	suite.repository.Clock = suite.clock
}

func (suite *MessageRepositoryTestSuite) createMessage(from, to, message string) *entity.Message {
	entity, err := suite.repository.Create(from, to, message)
	suite.Require().NoError(err)
	suite.Require().NotNil(entity)

	return entity
}

func (suite *MessageRepositoryTestSuite) TestGetByIdNotExisting() {
	user, err := suite.repository.GetById("notExisting")
	suite.Nil(user)
	suite.EqualError(err, MessageNotFoundError.Error())
}

func (suite *MessageRepositoryTestSuite) TestGetByIdOK() {
	m := suite.createMessage("a", "b", "txt")

	message, err := suite.repository.GetById(m.Id)
	suite.NoError(err)
	suite.Require().NotNil(message)
	suite.Equal("a", message.From)
	suite.Equal("b", message.To)
	suite.Equal("txt", message.Message)
}

func (suite *MessageRepositoryTestSuite) TestCreateOK() {
	message := suite.createMessage("a", "b", "txt")

	suite.Equal(message.From, "a")
	suite.Equal(message.To, "b")
	suite.Equal(message.CreatedAt, suite.repository.Clock.Now())
	suite.Equal([]string{"a", "b"}, message.Users)
	suite.NotEmpty(message.Id)
}

func (suite *MessageRepositoryTestSuite) TestAllWithUserOk() {
	suite.createMessage("a", "b", "txt1")
	suite.clock.Advance(time.Microsecond)
	suite.createMessage("a", "c", "txt2")
	suite.clock.Advance(time.Microsecond)
	suite.createMessage("b", "c", "txt3")
	suite.clock.Advance(time.Microsecond)
	suite.createMessage("a", "b", "txt4")

	messages, err := suite.repository.AllWithUser("a")
	suite.NoError(err)
	suite.Require().NotNil(messages)
	suite.Require().Len(messages, 3)

	suite.Equal("txt1", messages[0].Message)
	suite.Equal("txt2", messages[1].Message)
	suite.Equal("txt4", messages[2].Message)
}
