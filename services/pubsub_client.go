package services

import (
	"cloud.google.com/go/pubsub"
	"encoding/json"
	"fmt"
	"github.com/asiragusa/wschat/entity"
	"github.com/asiragusa/wschat/repository"
	"github.com/satori/go.uuid"
	"golang.org/x/net/context"
	"time"
)

// Interface used mainly for Unit testing
type PubsubClient interface {
	Publish(entity.Message) error
	Subscribe(string, func(message entity.Message)) (error, func())
}

// Pubsub client
type Pubsub struct {
	// Injected via DI
	Client *pubsub.Client `inject:""`

	// Injected via DI
	SubscriptionRepository repository.SubscriptionRepository `inject:""`
}

func NewPubsubClient() *Pubsub {
	return &Pubsub{}
}

// Creates a new subscription
func (p Pubsub) createSubscription(to string) (*pubsub.Subscription, error) {
	// Get a random subscription name
	name := "T" + uuid.NewV4().String()

	// First create the topic
	ctx := context.Background()
	topic, err := p.Client.CreateTopic(ctx, name)
	if err != nil {
		return nil, err
	}
	defer topic.Stop()

	// Create the sub
	subscription, err := p.Client.CreateSubscription(ctx, name, pubsub.SubscriptionConfig{
		Topic:       topic,
		AckDeadline: 10 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	// Store the subscription in the db
	_, err = p.SubscriptionRepository.Create(name, to)
	if err != nil {
		p.deleteSubscription(subscription)
		return nil, err
	}

	return subscription, nil
}

// Deletes a subscription
func (p Pubsub) deleteSubscription(subscription *pubsub.Subscription) {
	// First remove it from the db
	if err := p.SubscriptionRepository.Delete(subscription.ID()); err != nil {
		// TODO: properly log the error
		fmt.Println(err.Error())
	}

	ctx := context.Background()
	config, err := subscription.Config(ctx)
	if err != nil {
		// TODO: properly log the error
		fmt.Println(err.Error())

		err = subscription.Delete(ctx)
		// TODO: properly log the error
		fmt.Println(err.Error())
		return
	}

	if err := subscription.Delete(ctx); err != nil {
		// TODO: properly log the error
		fmt.Println(err.Error())
	}

	topic := config.Topic
	if err := topic.Delete(ctx); err != nil {
		// TODO: properly log the error
		fmt.Println(err.Error())
	}

	topic.Stop()
}

// Get all the topics belonging to a given user
func (p Pubsub) getTopics(to string) ([]*pubsub.Topic, error) {
	topicList, err := p.SubscriptionRepository.AllTo(to)
	if err != nil {
		return nil, err
	}

	var topics []*pubsub.Topic
	ctx := context.Background()

	for _, t := range topicList {
		topic := p.Client.Topic(t.Id)
		ok, err := topic.Exists(ctx)
		if err != nil {
			return nil, err
		}

		if ok {
			topics = append(topics, topic)
		}
	}

	return topics, nil
}

// Publish a message. The field message.To is used to identify the receivers
func (p Pubsub) Publish(message entity.Message) error {
	json, err := json.Marshal(&message)
	if err != nil {
		return err
	}

	// Get all the topics belonging to the mesage receiver
	topics, err := p.getTopics(message.To)
	if err != nil {
		return err
	}

	// Publish the message on every open topic. If the receiver is not connected, no message will be published
	for _, topic := range topics {
		topic.Publish(context.Background(), &pubsub.Message{
			Data: json,
		})
	}
	return nil
}

// Subscribe to the messages sent to the to user. The cb function is called when a new message is received.
//
// Returns an error if something went wrong and the cancel function, used to delete the subscription
func (p Pubsub) Subscribe(to string, cb func(message entity.Message)) (error, func()) {
	// Creates the subscription
	subscription, err := p.createSubscription(to)
	if err != nil {
		return err, nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Receive the messages in a new goroutine
	go func() {
		err := subscription.Receive(ctx, func(context context.Context, m *pubsub.Message) {
			// Decode the received message
			var message entity.Message
			err := json.Unmarshal(m.Data, &message)

			if err != nil {
				// TODO: properly log the error
				fmt.Println(err.Error())

				// Don't acknowledge de message
				m.Nack()
				return
			}

			// Call th cb function, ignoring errors from the receiver function
			cb(message)

			// Acknowledge the message
			m.Ack()
		})

		if err != nil {
			// TODO: properly log the error
			fmt.Println(err.Error())
		}
	}()

	return nil, func() {
		cancel()
		p.deleteSubscription(subscription)
	}
}
