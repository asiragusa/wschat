package services

import (
	"cloud.google.com/go/pubsub"
	"github.com/asiragusa/wschat/entity"
	"github.com/asiragusa/wschat/mocks"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
	"sync"
	"testing"
	"time"
)

type PubsubClientTestSuite struct {
	suite.Suite
	client         *Pubsub
	clock          clockwork.FakeClock
	subsRepository *mocks.SubscriptionRepository
}

func TestPubsubClient(t *testing.T) {
	suite.Run(t, new(PubsubClientTestSuite))
}

func getPubsubClient(projectID string) (*pubsub.Client, error) {
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (suite *PubsubClientTestSuite) SetupSuite() {
	client, err := getPubsubClient("test")
	suite.Require().NoError(err)

	suite.client = NewPubsubClient()
	suite.client.Client = client
}

func (suite *PubsubClientTestSuite) cleanSubs() {
	ctx := context.Background()
	it := suite.client.Client.Subscriptions(ctx)
	for {
		sub, err := it.Next()
		if err == iterator.Done {
			break
		}
		suite.Require().Nil(err)

		suite.Require().NoError(sub.Delete(ctx))
	}
}
func (suite *PubsubClientTestSuite) cleanPubs() {
	ctx := context.Background()
	it := suite.client.Client.Topics(ctx)
	for {
		topic, err := it.Next()
		if err == iterator.Done {
			break
		}
		suite.Require().Nil(err)

		suite.Require().NoError(topic.Delete(ctx))
	}
}

func (suite *PubsubClientTestSuite) SetupTest() {
	suite.cleanSubs()
	suite.cleanPubs()

	suite.subsRepository = &mocks.SubscriptionRepository{}
	suite.client.SubscriptionRepository = suite.subsRepository
}

func (suite *PubsubClientTestSuite) TearDownTest() {
	suite.subsRepository.AssertExpectations(suite.T())
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

type MockReceiver struct {
	mock.Mock
}

func (m *MockReceiver) Receive(message string) {
	m.Called(message)
}

func (suite *PubsubClientTestSuite) TestGetTopicsRepoError() {
	suite.subsRepository.On("AllTo", "to1").Return(nil, assert.AnError)
	topics, err := suite.client.getTopics("to1")
	suite.Nil(topics)
	suite.Require().NotNil(err)
	suite.EqualError(err, assert.AnError.Error())
}

func (suite *PubsubClientTestSuite) TestPublishRepoError() {
	suite.subsRepository.On("AllTo", "to1").Return(nil, assert.AnError)
	err := suite.client.Publish(entity.Message{
		To: "to1",
	})
	suite.Require().NotNil(err)
	suite.EqualError(err, assert.AnError.Error())
}

func (suite *PubsubClientTestSuite) TestGetTopicsNotExisting() {
	suite.subsRepository.On("AllTo", "to1").Return([]entity.Subscription{
		{
			Id: "id1",
			To: "to1",
		},
	}, nil)

	topics, err := suite.client.getTopics("to1")
	suite.Require().NoError(err)

	suite.Len(topics, 0)
}

func (suite *PubsubClientTestSuite) TestGetTopicsOK() {
	subscriptions := []entity.Subscription{
		{
			Id: "id1",
			To: "to1",
		},
		{
			Id: "id2",
			To: "to1",
		},
		{
			Id: "id3",
			To: "to1",
		},
	}
	suite.subsRepository.On("AllTo", "to1").Return(subscriptions, nil)

	ctx := context.Background()

	topic1, err := suite.client.Client.CreateTopic(ctx, "id1")
	suite.Require().NoError(err)
	defer topic1.Stop()

	topic2, err := suite.client.Client.CreateTopic(ctx, "id2")
	suite.Require().NoError(err)
	defer topic2.Stop()

	topics, err := suite.client.getTopics("to1")
	suite.Require().NoError(err)

	defer func() {
		for _, t := range topics {
			t.Stop()
		}
	}()

	suite.Require().Len(topics, 2)
	suite.Equal(topic1.ID(), topics[0].ID())
	suite.Equal(topic2.ID(), topics[1].ID())
}

func (suite *PubsubClientTestSuite) TestCreateSubscriptionRepoError() {
	to := "test@test.com"

	suite.subsRepository.On("Create", mock.AnythingOfType("string"), to).Return(nil, assert.AnError)
	suite.subsRepository.On("Delete", mock.AnythingOfType("string")).Return(nil)
	subscription, err := suite.client.createSubscription(to)

	suite.Require().NotNil(err)
	suite.EqualError(err, assert.AnError.Error())
	suite.Nil(subscription)

}

func (suite *PubsubClientTestSuite) createSubscription(to string) *pubsub.Subscription {
	subEntity := &entity.Subscription{}
	suite.subsRepository.On("Create", mock.AnythingOfType("string"), to).Return(subEntity, nil)

	subscription, err := suite.client.createSubscription(to)
	suite.Require().NoError(err)

	return subscription
}

func (suite *PubsubClientTestSuite) TestCreateSubscriptionOK() {
	subscription := suite.createSubscription("to1")

	ok, err := subscription.Exists(context.Background())
	suite.Require().NoError(err)

	suite.True(ok)
}

func (suite *PubsubClientTestSuite) TestDeleteSubscriptionOK() {
	subscription := suite.createSubscription("to1")

	suite.subsRepository.On("Delete", subscription.ID()).Return(nil)

	suite.client.deleteSubscription(subscription)
}

func (suite *PubsubClientTestSuite) TestPublishSubscribe() {
	to := "a@b.com"
	now := time.Now()
	messages := []entity.Message{
		{
			Id:        "test1",
			From:      "from1",
			To:        to,
			Message:   "test1",
			CreatedAt: now,
		},
		{
			Id:        "test2",
			From:      "from1",
			To:        to,
			Message:   "test2",
			CreatedAt: now,
		},
		{
			Id:        "test3",
			From:      "from1",
			To:        "b@b.com",
			Message:   "test3",
			CreatedAt: now,
		},
	}

	var receiver MockReceiver

	var wg sync.WaitGroup
	wg.Add(4) //subscriptions * len messages to a@b.com

	receiveFn := func(message entity.Message) {
		defer wg.Done()
		receiver.Receive(message.Id)
	}

	var subIds []string

	subEntity := &entity.Subscription{}
	suite.subsRepository.On("Create", mock.MatchedBy(func(id string) bool {
		subIds = append(subIds, id)
		return true
	}), to).Twice().Return(subEntity, nil)

	err, cancel1 := suite.client.Subscribe(to, receiveFn)
	suite.Require().NoError(err)

	err, cancel2 := suite.client.Subscribe(to, receiveFn)
	suite.Require().NoError(err)

	receiver.On("Receive", "test1").Twice()
	receiver.On("Receive", "test2").Twice()

	suite.subsRepository.On("AllTo", to).Return([]entity.Subscription{
		{
			Id: subIds[0],
		},
		{
			Id: subIds[1],
		},
	}, nil)

	suite.subsRepository.On("AllTo", "b@b.com").Return([]entity.Subscription{}, nil)

	for _, message := range messages {
		err := suite.client.Publish(message)
		suite.Require().NoError(err)
	}

	timeout := waitTimeout(&wg, time.Second)
	suite.Require().False(timeout)

	suite.subsRepository.On("Delete", subIds[0]).Return(nil)
	cancel1()

	suite.subsRepository.On("Delete", subIds[1]).Return(nil)
	cancel2()

	receiver.AssertExpectations(suite.T())
}
