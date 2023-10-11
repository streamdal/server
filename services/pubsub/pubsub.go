package pubsub

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/streamdal/snitch-server/util"
)

// IPubSub is an interface for an internal, in-memory pub/sub system. We use it
// to signal events to subscribers (such as when new clients Register() to the
// server or when disconnects occur). This is used to tell the frontend that
// a change has occurred (and the frontend must refresh its view.
type IPubSub interface {
	// HaveTopic determines if there are any open pubsub channels for a given topic
	HaveTopic(topic string) bool

	// Listen will "subscribe" to a topic and return a channel that will receive
	// any messages that are published to that topic. Listen accepts an optional
	// identifier that will be used to identify a specific listener. The identifier
	// is useful for being able to close a _specific_ channel.
	Listen(topic string, channelID ...string) chan interface{}

	// Publish will publish a message to a topic, which may have multiple channels associated
	// with it. Each channel will receive the message
	Publish(topic string, m interface{})

	// Close will delete the channel from the topic map and close the channel.
	// WARNING: Make sure to call Close() only on listeners that no longer Listen()'ing
	Close(topic, channelID string)

	// CloseTopic is used when a tail request is stopped to close all associated channels
	// and prevent a dead-lock
	CloseTopic(topic string) bool

	// Reset will delete all channels from the topic map and close all channels;
	// use this when you are finished
	Reset()
}

//	type PubSub struct {
//		topics map[string]map[string]chan interface{} // k1: topic k2: subscriber id v: channel
//		mtx    *sync.RWMutex
//		log    *logrus.Entry
//	}
type PubSub struct {
	topics  map[string]map[string]chan interface{} // k1: topic k2: subscriber id v: channel
	closing map[string]bool                        // Track whether a topic is being closed
	mtx     *sync.RWMutex
	log     *logrus.Entry
}

//func New() *PubSub {
//	return &PubSub{
//		topics: make(map[string]map[string]chan interface{}),
//		log:    logrus.WithField("pkg", "pubsub"),
//		mtx:    &sync.RWMutex{},
//	}
//}

func New() *PubSub {
	return &PubSub{
		topics:  make(map[string]map[string]chan interface{}),
		closing: make(map[string]bool),
		log:     logrus.WithField("pkg", "pubsub"),
		mtx:     &sync.RWMutex{},
	}
}

func (ps *PubSub) Listen(topic string, channelID ...string) chan interface{} {
	var id string

	if len(channelID) > 0 {
		id = channelID[0]
	} else {
		id = util.GenerateUUID()
	}

	ch := make(chan interface{}, 1)

	fmt.Println("pubsub.Listen: before lock")

	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	fmt.Println("pubsub.Listen: acquired lock")

	if _, ok := ps.topics[topic]; !ok {
		fmt.Printf("pubsub.Listen: topic '%s' not found, creating\n", topic)
		ps.topics[topic] = make(map[string]chan interface{})
	}

	ps.topics[topic][id] = ch

	fmt.Println("pubsub.Listen: after unlock")

	return ch
}

func (ps *PubSub) Close(topic, channelID string) {
	fmt.Println("pubsub.Close: before lock")

	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	fmt.Println("pubsub.Close: acquired lock")

	ps.closing[topic] = true
	defer func() { ps.closing[topic] = false }()

	ch, ok := ps.topics[topic][channelID]
	if !ok {
		// Nothing to do, asked to cleanup ch that does not exist
		return
	}

	delete(ps.topics[topic], channelID)

	if ch != nil {
		close(ch)
	}

	fmt.Println("pubsub.Close: after unlock")
}

func (ps *PubSub) CloseTopic(topic string) bool {
	fmt.Println("pubsub.CloseTopic: before lock")

	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	fmt.Println("pubsub.CloseTopic: acquired lock")

	ps.closing[topic] = true
	defer func() { ps.closing[topic] = false }()

	channels, ok := ps.topics[topic]
	if !ok {
		return false
	}

	for _, ch := range channels {
		close(ch)
	}

	delete(ps.topics, topic)

	fmt.Println("pubsub.CloseTopic: after unlock")

	return true
}

func (ps *PubSub) Reset() {
	fmt.Println("pubsub.Reset: before lock")

	ps.mtx.Lock()
	defer ps.mtx.Unlock()
	fmt.Println("pubsub.Reset: acquired lock")

	for topic, chs := range ps.topics {
		ps.closing[topic] = true
		for _, ch := range chs {
			close(ch)
		}
		ps.closing[topic] = false
	}

	ps.topics = make(map[string]map[string]chan interface{})
	ps.closing = make(map[string]bool)

	fmt.Println("pubsub.Reset: after unlock")
}

func (ps *PubSub) Publish(topic string, m interface{}) {
	fmt.Println("pubsub.Publish: before lock")

	ps.mtx.RLock()
	defer ps.mtx.RUnlock()
	fmt.Println("pubsub.Publish: acquired lock")

	if _, ok := ps.topics[topic]; !ok {
		return
	}

	if ps.closing[topic] {
		return
	}

	for _, tmpCh := range ps.topics[topic] {
		go func(ch chan interface{}) {
			ps.log.Debugf("pubsub.Publish: before topic '%s' write", topic)
			ch <- m
			ps.log.Debugf("pubsub.Publish: after topic '%s' write", topic)
		}(tmpCh)
	}

	fmt.Println("pubsub.Publish: after unlock")
}

func (ps *PubSub) HaveTopic(topic string) bool {
	fmt.Println("pubsub.HaveTopic: before lock")

	ps.mtx.RLock()
	defer ps.mtx.RUnlock()
	fmt.Println("pubsub.HaveTopic: acquired lock")

	_, ok := ps.topics[topic]

	fmt.Println("pubsub.HaveTopic: after unlock")

	return ok
}
