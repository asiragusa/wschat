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

type SubscriptionRepositoryTestSuite struct {
	suite.Suite
	repository *Subscription
	clock      clockwork.FakeClock
}

func TestSubscriptionRepository(t *testing.T) {
	suite.Run(t, new(SubscriptionRepositoryTestSuite))
}

func (suite *SubscriptionRepositoryTestSuite) SetupSuite() {
	client, err := getDatastoreClient("test")
	suite.Require().NoError(err)

	suite.repository = NewSubscriptionRepository()
	suite.repository.Client = client
}

func (suite *SubscriptionRepositoryTestSuite) cleanDb() {
	query := datastore.NewQuery("").KeysOnly()
	ctx := context.Background()

	keys, err := suite.repository.Client.GetAll(ctx, query, nil)
	suite.Require().NoError(err)

	err = suite.repository.Client.DeleteMulti(ctx, keys)
	suite.Require().NoError(err)
}

func (suite *SubscriptionRepositoryTestSuite) SetupTest() {
	suite.cleanDb()

	suite.clock = clockwork.NewFakeClockAt(time.Now())
	suite.repository.Clock = suite.clock
}

func (suite *SubscriptionRepositoryTestSuite) createSubscription(id, to string) *entity.Subscription {
	entity, err := suite.repository.Create(id, to)
	suite.Require().NoError(err)
	suite.Require().NotNil(entity)

	return entity
}

func (suite *SubscriptionRepositoryTestSuite) TestCreateSameId() {
	suite.createSubscription("id", "to")
	entity, err := suite.repository.Create("id", "whatever")

	suite.Nil(entity)
	suite.EqualError(err, SubscriptionAlreadyExistsError.Error())

}

func (suite *SubscriptionRepositoryTestSuite) TestCreateOK() {
	Subscription := suite.createSubscription("id", "to")

	suite.Equal(Subscription.Id, "id")
	suite.Equal(Subscription.To, "to")
	suite.Equal(Subscription.CreatedAt, suite.repository.Clock.Now())
}

func (suite *SubscriptionRepositoryTestSuite) TestAllToOk() {
	suite.createSubscription("id1", "a")
	suite.createSubscription("id2", "a")
	suite.createSubscription("id3", "b")

	Subscriptions, err := suite.repository.AllTo("a")
	suite.NoError(err)
	suite.Require().NotNil(Subscriptions)
	suite.Require().Len(Subscriptions, 2)
}

func (suite *SubscriptionRepositoryTestSuite) TestDeleteOk() {
	suite.createSubscription("id1", "a")
	err := suite.repository.Delete("id1")
	suite.NoError(err)
}
