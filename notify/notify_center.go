package notify

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"nacos-go/utils"
)

type void struct{}

var member void

type PublisherCenter struct {
	sharePublisher   *Publisher
	Publishers       sync.Map
	smartSubscribers utils.SyncSet
	hasSubscriber    bool

	// just for test
	onExpire func(event Event)
}

var instance *PublisherCenter

var once sync.Once

var defaultFastRingBufferSize = utils.GetInt64FromEnvOptional("conf.notify.fast-event-buffer.size", 1024)
var defaultSlowRingBufferSize = utils.GetInt64FromEnvOptional("conf.notify.slow-event-buffer.size", 16384)

func Init() {
	once.Do(func() {
		instance = &PublisherCenter{
			Publishers:       sync.Map{},
			smartSubscribers: utils.NewSyncSet(),
			hasSubscriber:    false,
		}

		instance.sharePublisher = &Publisher{
			queue: make(chan Event, defaultSlowRingBufferSize),
			topic: "00--0-SlowEvent-0--00",
		}

		instance.sharePublisher.start()

	})
}

func RegisterDefaultPublisher(event Event) (*Publisher, error) {
	return RegisterPublisher(event, defaultFastRingBufferSize)
}

func RegisterSharePublisher(event Event) (*Publisher, error) {
	return RegisterPublisher(event, 0)
}

func RegisterPublisher(event Event, ringBufferSize int64) (*Publisher, error) {
	if ringBufferSize <= 32 {
		ringBufferSize = 128
	}
	switch t := event.(type) {
	case FastEvent:
		topic := t.Name()

		instance.Publishers.LoadOrStore(topic, &Publisher{
			queue:        make(chan Event, ringBufferSize),
			topic:        topic,
			subscribers:  sync.Map{},
			lastSequence: -1,
		})

		p, ok := instance.Publishers.Load(topic)

		if ok {
			publisher := p.(*Publisher)
			publisher.start()

			return publisher, nil
		}

		return nil, errors.New("register event publisher failed")
	case SlowEvent:
		return instance.sharePublisher, nil
	default:
		_ = t
		return nil, errors.New("this event not support, just support notify/notify_center.Event or notify/notify_center.SlowEvent")
	}
}

func PublishEvent(event Event) (bool, error) {
	switch t := event.(type) {
	case FastEvent:
		if p, ok := instance.Publishers.Load(event.Name()); ok {
			p.(*Publisher).PublishEvent(event)
			return true, nil
		}
		return false, errors.New("the event was not registered")
	case SlowEvent:
		instance.sharePublisher.PublishEvent(event)
		return true, nil
	default:
		_ = t
		return false, errors.New("this event not support, just support notify/notify_center.Event or notify/notify_center.SlowEvent")
	}
}

func RegisterSubscriber(s Subscriber) error {
	switch t := s.(type) {
	case SingleSubscriber:

		topic := t.SubscribeType()

		if v, ok := instance.Publishers.Load(topic); ok {
			p := v.(*Publisher)
			(*p).AddSubscriber(s)
			return nil
		}

		return fmt.Errorf("this topic [%s] no publisher", topic)

	case MultiSubscriber:
		instance.smartSubscribers.Add(s)
		instance.hasSubscriber = true
		return nil
	default:
		_ = t
		return errors.New("wrong subscriber type")
	}
}

func DeregisterSubscriber(s Subscriber) error {
	switch t := s.(type) {
	case SingleSubscriber:

		topic := t.SubscribeType()

		if v, ok := instance.Publishers.Load(topic); ok {
			p := v.(*Publisher)
			(*p).RemoveSubscriber(s)
			return nil
		}

		return fmt.Errorf("this topic [%s] no publisher", topic)

	case MultiSubscriber:
		instance.smartSubscribers.Remove(s)
		return nil
	default:
		_ = t
		return errors.New("wrong subscriber type")
	}
}

func Shutdown() {

	instance.sharePublisher.shutdown()

	instance.Publishers.Range(func(key, value interface{}) bool {
		p := key.(*Publisher)
		(*p).shutdown()
		return true
	})

}

func RegisterOnExpire(f func(event Event)) {
	instance.onExpire = f
}

// Event interface
type Event interface {
	// Topic of the event
	Name() string
}

// That means it happen very fast
type FastEvent interface {
	Event
	// The sequence number of the event
	Sequence() int64
}

// That means it doesn't happen very often
type SlowEvent interface {
	Event
}

type Subscriber interface {

	OnEvent(event Event)

	IgnoreExpireEvent() bool
}

type SingleSubscriber interface {

	Subscriber

	SubscribeType() string
}

type MultiSubscriber interface {

	Subscriber

	CanNotify(name string) bool
}

type Publisher struct {
	queue        chan Event
	topic        string
	subscribers  sync.Map
	canOpen      bool
	init         sync.Once
	lastSequence int64
}

func (p *Publisher) start() {
	p.init.Do(func() {
		go p.openHandler()
	})
}

func (p *Publisher) PublishEvent(event Event) {
	p.queue <- event
}

func (p *Publisher) AddSubscriber(s Subscriber) {
	p.subscribers.Store(s, member)
	p.canOpen = true
}

func (p *Publisher) RemoveSubscriber(s Subscriber) {
	p.subscribers.Delete(s)
}

func (p *Publisher) shutdown() {
	close(p.queue)
}

func (p *Publisher) openHandler() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("dispose event has error : %s", err)
		}
	}()

	for {
		if p.canOpen || instance.hasSubscriber {
			break
		}
		time.Sleep(time.Duration(100) * time.Millisecond)
	}

	for e := range p.queue {

		fmt.Printf("handler receive event : %s\n", e)

		name := e.Name()

		switch t := e.(type) {

		case FastEvent:

			currentSequence := t.Sequence()

			p.subscribers.Range(func(key, value interface{}) bool {

				defer func() {
					if err := recover(); err != nil {
						fmt.Printf("notify subscriber has error : %s \n", err)
					}
				}()

				subscriber := key.(SingleSubscriber)

				// shared message channels need to handle event types
				if strings.Compare(name, subscriber.SubscribeType()) != 0 {
					return true
				}

				if subscriber.IgnoreExpireEvent() && currentSequence < p.lastSequence {

					// just for test
					if instance.onExpire != nil {
						instance.onExpire(e)
					}

					return true
				}

				subscriber.OnEvent(e)
				return true
			})

			p.lastSequence = int64(math.Max(float64(currentSequence), float64(p.lastSequence)))

		case SlowEvent:

			instance.smartSubscribers.Range(func(value interface{}) {

				defer func() {
					if err := recover(); err != nil {
						fmt.Printf("notify multi subscriber has error : %s \n", err)
					}
				}()

				subscriber := value.(MultiSubscriber)

				if subscriber.CanNotify(name) {
					subscriber.OnEvent(e)
				}

			})
		}

	}
}
