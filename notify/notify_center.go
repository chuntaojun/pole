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
	sharePublisher *SharePublisher
	Publishers     sync.Map
	hasSubscriber  bool

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
			Publishers:    sync.Map{},
			hasSubscriber: false,
		}

		instance.sharePublisher = &SharePublisher{}
		instance.sharePublisher.queue = make(chan Event, defaultSlowRingBufferSize)
		instance.sharePublisher.topic = "00--0-SlowEvent-0--00"
		instance.sharePublisher.start()

	})
}

func RegisterDefaultPublisher(event Event) error {
	return RegisterPublisher(event, defaultFastRingBufferSize)
}

func RegisterSharePublisher(event Event) error {
	return RegisterPublisher(event, 0)
}

func RegisterPublisher(event Event, ringBufferSize int64) error {
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

			return nil
		}

		return errors.New("register event publisher failed")
	case SlowEvent:
		return nil
	default:
		_ = t
		return errors.New("this event not support, just support notify/notify_center.Event or notify/notify_center.SlowEvent")
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

		switch e := topic.(type) {
		case FastEvent:
			if v, ok := instance.Publishers.Load(e.Name()); ok {
				p := v.(*Publisher)
				(*p).AddSubscriber(s)
				return nil
			}

			return fmt.Errorf("this topic [%s] no publisher", topic)
		case SlowEvent:
			instance.sharePublisher.AddSubscriber(s)
			return nil
		default:
			return errors.New("wrong subscriber type")
		}

	case MultiSubscriber:
		names := t.SubscribeTypes()
		for _, topic := range names {
			switch e := topic.(type) {
			case FastEvent:
				if v, ok := instance.Publishers.Load(e.Name()); ok {
					p := v.(*Publisher)
					(*p).AddSubscriber(s)
					return nil
				}

				return fmt.Errorf("this topic [%s] no publisher", topic)
			case SlowEvent:
				instance.sharePublisher.AddSubscriber(s)
				return nil;
			}
		}
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
		names := t.SubscribeTypes()
		for _, topic := range names {
			if v, ok := instance.Publishers.Load(topic); ok {
				p := v.(*Publisher)
				(*p).RemoveSubscriber(s)
			}
		}
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

	SubscribeType() Event
}

type MultiSubscriber interface {
	Subscriber

	SubscribeTypes() []Event
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
		if p.canOpen {
			break
		}
		time.Sleep(time.Duration(100) * time.Millisecond)
	}

	for e := range p.queue {
		fmt.Printf("handler receive event : %s\n", e)
		p.notifySubscriber(e)
	}
}

func (p *Publisher) notifySubscriber(event Event) {
	currentSequence := event.(FastEvent).Sequence()
	p.subscribers.Range(func(key, value interface{}) bool {

		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("notify subscriber has error : %s \n", err)
			}
		}()

		subscriber := key.(Subscriber)

		if subscriber.IgnoreExpireEvent() && currentSequence < p.lastSequence {
			// just for test
			if instance.onExpire != nil {
				instance.onExpire(event)
			}
			return true
		}

		subscriber.OnEvent(event)
		return true
	})

	p.lastSequence = int64(math.Max(float64(currentSequence), float64(p.lastSequence)))
}

type SharePublisher struct {
	Publisher
}

func (sp *SharePublisher) notifySubscriber(event Event) {
	topic := event.Name()
	sp.subscribers.Range(func(key, value interface{}) bool {

		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("notify subscriber has error : %s \n", err)
			}
		}()

		switch s := key.(type) {
		case SingleSubscriber:

			if strings.Compare(topic, s.SubscribeType().Name()) == 0 {
				s.OnEvent(event)
			}

		case MultiSubscriber:
			for _, watch := range s.SubscribeTypes() {
				if strings.Compare(watch.Name(), topic) == 0 {
					s.OnEvent(event)
					break
				}
			}
		}

		return true
	})

}
