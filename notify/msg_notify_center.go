// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package notify

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/Conf-Group/pole/utils"
)

var (
	ErrorEventNotSupport     = errors.New("this event not support, just support notify/notify_center.Event or notify/notify_center.SlowEvent")
	ErrorEventNotRegister    = errors.New("the event was not registered")
	ErrorEventRegister       = errors.New("register event publisher failed")
	ErrorWrongSubscriberType = errors.New("wrong subscriber type")
	ErrorAddSubscriber       = errors.New("add subscriber failed")

	member          void
	publisherCenter *msgPublisherCenter
	pOnce           sync.Once

	defaultFastRingBufferSize = utils.GetInt64FromEnvOptional("conf.notify.fast-event-buffer.size", 1024)
	defaultSlowRingBufferSize = utils.GetInt64FromEnvOptional("conf.notify.slow-event-buffer.size", 16384)
)

type void struct{}

type msgPublisherCenter struct {
	sharePublisher *SharePublisher
	Publishers     sync.Map
	hasSubscriber  bool
	log            utils.Logger
	// just for test
	onExpire func(event Event)
}

// 事件中心初始化
func init() {
	pOnce.Do(func() {
		// 构建事件中心管理者
		publisherCenter = &msgPublisherCenter{
			Publishers:    sync.Map{},
			hasSubscriber: false,
		}

		// 注册共享的事件队列
		publisherCenter.sharePublisher = &SharePublisher{
			listeners: sync.Map{},
		}
		// 设置共享事件通道的长度
		publisherCenter.sharePublisher.queue = make(chan Event, defaultSlowRingBufferSize)
		// 设置共享topic
		publisherCenter.sharePublisher.topic = "00--0-SlowEvent-0--00"
		// 设置共享事件发布者开启
		publisherCenter.sharePublisher.start()
		// 还不允许事件通知消费者，为了避免消息丢失
		publisherCenter.sharePublisher.canOpen = make(chan bool)

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

		publisherCenter.Publishers.LoadOrStore(topic, &Publisher{
			owner:        publisherCenter,
			queue:        make(chan Event, ringBufferSize),
			topic:        topic,
			subscribers:  sync.Map{},
			lastSequence: -1,
		})

		p, ok := publisherCenter.Publishers.Load(topic)

		if ok {
			publisher := p.(*Publisher)
			publisher.start()
			return nil
		}

		return ErrorEventRegister
	case SlowEvent:
		return nil
	default:
		_ = t
		return ErrorEventNotSupport
	}
}

func PublishEventWait(ctx context.Context, event Event) (bool, error) {
	switch t := event.(type) {
	case FastEvent:
		if p, ok := publisherCenter.Publishers.Load(event.Name()); ok {
			return p.(*Publisher).PublishEvent(ctx, event)
		}
		return false, ErrorEventNotRegister
	case SlowEvent:
		return publisherCenter.sharePublisher.PublishEvent(ctx, event)
	default:
		_ = t
		return false, ErrorEventNotSupport
	}
}

var noWaitContext = context.Background()

func PublishEvent(event Event) (bool, error) {
	switch t := event.(type) {
	case FastEvent:
		if p, ok := publisherCenter.Publishers.Load(event.Name()); ok {
			return p.(*Publisher).PublishEvent(noWaitContext, event)
		}
		return false, ErrorEventNotRegister
	case SlowEvent:
		return publisherCenter.sharePublisher.PublishEvent(noWaitContext, event)
	default:
		_ = t
		return false, ErrorEventNotSupport
	}
}

func RegisterSubscriber(s Subscriber) error {
	switch t := s.(type) {
	case SingleSubscriber:
		topic := t.SubscribeType()
		switch e := topic.(type) {
		case FastEvent:
			if v, ok := publisherCenter.Publishers.Load(e.Name()); ok {
				p := v.(*Publisher)
				(*p).AddSubscriber(s)
				return nil
			}

			return fmt.Errorf("this topic [%s] no publisher", topic)
		case SlowEvent:
			publisherCenter.sharePublisher.AddSubscriber(s)
			return nil
		default:
			return ErrorWrongSubscriberType
		}

	case MultiSubscriber:
		names := t.SubscribeTypes()
		for _, topic := range names {
			switch e := topic.(type) {
			case FastEvent:
				if v, ok := publisherCenter.Publishers.Load(e.Name()); ok {
					p := v.(*Publisher)
					(*p).AddSubscriber(s)
				} else {
					return fmt.Errorf("this topic [%s] no publisher", topic)
				}
			case SlowEvent:
				// do nothing
			}
		}
		publisherCenter.sharePublisher.AddSubscriber(s)
		return nil
	default:
		_ = t
		return ErrorWrongSubscriberType
	}
}

func DeregisterSubscriber(s Subscriber) error {
	switch t := s.(type) {
	case SingleSubscriber:
		topic := t.SubscribeType()
		publisherCenter.sharePublisher.RemoveSubscriber(s)
		if v, ok := publisherCenter.Publishers.Load(topic); ok {
			p := v.(*Publisher)
			(*p).RemoveSubscriber(s)
		}
		return nil
	case MultiSubscriber:
		names := t.SubscribeTypes()
		for _, topic := range names {
			if v, ok := publisherCenter.Publishers.Load(topic); ok {
				p := v.(*Publisher)
				(*p).RemoveSubscriber(s)
			}
		}
		publisherCenter.sharePublisher.RemoveSubscriber(s)
		return nil
	default:
		_ = t
		return ErrorWrongSubscriberType
	}
}

func Shutdown() {
	publisherCenter.sharePublisher.shutdown()

	publisherCenter.Publishers.Range(func(key, value interface{}) bool {
		p := key.(*Publisher)
		(*p).shutdown()
		return true
	})

}

func RegisterOnExpire(f func(event Event)) {
	publisherCenter.onExpire = f
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
	owner        *msgPublisherCenter
	queue        chan Event
	topic        string
	subscribers  sync.Map
	canOpen      chan bool
	init         sync.Once
	lastSequence int64
}

func (p *Publisher) start() {
	p.init.Do(func() {
		ctx := context.WithValue(context.Background(), utils.TraceIDKey, p.topic)
		utils.Go(ctx, func(cxt context.Context) {
			p.openHandler(ctx)
		})
	})
}

func (p *Publisher) PublishEvent(ctx context.Context, event Event) (bool, error) {
	select {
	case p.queue <- event:
		return true, nil
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

func (p *Publisher) AddSubscriber(s Subscriber) {
	p.subscribers.Store(s, member)
	p.canOpen <- true
}

func (p *Publisher) RemoveSubscriber(s Subscriber) {
	p.subscribers.Delete(s)
}

func (p *Publisher) shutdown() {
	close(p.queue)
}

func (p *Publisher) openHandler(cxt context.Context) {

	defer func() {
		if err := recover(); err != nil {
			p.owner.log.Debug(cxt, "dispose fast event has error : %s", err)
		}
	}()

	<-p.canOpen

	for e := range p.queue {
		p.owner.log.Debug(cxt, "handler receive fast event : %#v", e)
		p.notifySubscriber(cxt, e)
	}
}

func (p *Publisher) notifySubscriber(ctx context.Context, event Event) {
	currentSequence := event.(FastEvent).Sequence()
	p.subscribers.Range(func(key, value interface{}) bool {

		defer func() {
			if err := recover(); err != nil {
				p.owner.log.Error(ctx, "notify subscriber has error : %s", err)
			}
		}()

		subscriber := key.(Subscriber)

		if subscriber.IgnoreExpireEvent() && currentSequence < p.lastSequence {
			// just for test
			if publisherCenter.onExpire != nil {
				publisherCenter.onExpire(event)
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
	listeners sync.Map
}

func (sp *SharePublisher) start() {
	sp.init.Do(func() {
		ctx := context.WithValue(context.Background(), utils.TraceIDKey, sp.topic)
		utils.Go(ctx, func(ctx context.Context) {
			sp.openHandler(ctx)
		})
	})
}

func (sp *SharePublisher) AddSubscriber(s Subscriber) {
	switch t := s.(type) {
	case SingleSubscriber:
		topic := t.SubscribeType()
		sp.listeners.LoadOrStore(topic.Name(), utils.NewSyncSet())
		if set, exist := sp.listeners.Load(topic.Name()); exist {
			set.(*utils.SyncSet).Add(s)
			return
		}
		panic(ErrorAddSubscriber)
	case MultiSubscriber:
		for _, topic := range t.SubscribeTypes() {
			switch t := topic.(type) {
			case SlowEvent:
				sp.listeners.LoadOrStore(t.Name(), utils.NewSyncSet())
				if set, exist := sp.listeners.Load(t.Name()); exist {
					set.(*utils.SyncSet).Add(s)
				}
			}
		}
	}
	sp.canOpen <- true
}

func (sp *SharePublisher) RemoveSubscriber(s Subscriber) {
	switch t := s.(type) {
	case SingleSubscriber:
		topic := t.SubscribeType()
		if set, exist := sp.listeners.Load(topic.Name()); exist {
			set.(*utils.SyncSet).Remove(s)
			return
		}
		panic(ErrorAddSubscriber)
	case MultiSubscriber:
		for _, topic := range t.SubscribeTypes() {
			if set, exist := sp.listeners.Load(topic.Name()); exist {
				set.(*utils.SyncSet).Remove(s)
			}
		}
	}
}

func (sp *SharePublisher) openHandler(ctx context.Context) {

	defer func() {
		if err := recover(); err != nil {
			sp.owner.log.Debug(ctx, "dispose slow event has error : %s", err)
		}
	}()

	<-sp.canOpen

	for e := range sp.queue {
		sp.owner.log.Debug(ctx, "handler receive slow event : %#v", e)
		sp.notifySubscriber(e)
	}
}

func (sp *SharePublisher) notifySubscriber(event Event) {
	topic := event.Name()
	if set, exist := sp.listeners.Load(topic); exist {
		set.(*utils.SyncSet).Range(func(value interface{}) {
			value.(Subscriber).OnEvent(event)
		})
	}

}
