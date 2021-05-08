// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package notify

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"

	polerpc "github.com/pole-group/pole-rpc"

	"github.com/pole-group/pole/common"
	"github.com/pole-group/pole/utils"
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
	log            polerpc.Logger
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
	init         sync.Once
	lastSequence int64
}

func (p *Publisher) start() {
	p.init.Do(func() {
		ctx := common.NewCtxPole()
		polerpc.Go(ctx, func(arg interface{}) {
			p.openHandler(arg.(*common.ContextPole))
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
			p.owner.log.Error("%s : dispose fast event has error : %s", p.topic, err)
		}
	}()
	for e := range p.queue {
		p.owner.log.Debug("%s : handler receive fast event : %#v", p.topic, e)
		p.notifySubscriber(cxt, e)
	}
}

func (p *Publisher) notifySubscriber(ctx context.Context, event Event) {
	currentSequence := event.(FastEvent).Sequence()
	p.subscribers.Range(func(key, value interface{}) bool {

		defer func() {
			if err := recover(); err != nil {
				p.owner.log.Error("%s : notify subscriber has error : %s", p.topic, err)
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
		ctx := common.NewCtxPole()
		polerpc.Go(ctx, func(arg interface{}) {
			sp.openHandler(arg.(*common.ContextPole))
		})
	})
}

func (sp *SharePublisher) AddSubscriber(s Subscriber) {
	switch t := s.(type) {
	case SingleSubscriber:
		topic := t.SubscribeType()
		sp.listeners.LoadOrStore(topic.Name(), polerpc.NewSyncSet())
		if set, exist := sp.listeners.Load(topic.Name()); exist {
			set.(*polerpc.ConcurrentSet).Add(s)
			return
		}
		panic(ErrorAddSubscriber)
	case MultiSubscriber:
		for _, topic := range t.SubscribeTypes() {
			switch t := topic.(type) {
			case SlowEvent:
				sp.listeners.LoadOrStore(t.Name(), polerpc.NewSyncSet())
				if set, exist := sp.listeners.Load(t.Name()); exist {
					set.(*polerpc.ConcurrentSet).Add(s)
				}
			}
		}
	}
}

func (sp *SharePublisher) RemoveSubscriber(s Subscriber) {
	switch t := s.(type) {
	case SingleSubscriber:
		topic := t.SubscribeType()
		if set, exist := sp.listeners.Load(topic.Name()); exist {
			set.(*polerpc.ConcurrentSet).Remove(s)
			return
		}
		panic(ErrorAddSubscriber)
	case MultiSubscriber:
		for _, topic := range t.SubscribeTypes() {
			if set, exist := sp.listeners.Load(topic.Name()); exist {
				set.(*polerpc.ConcurrentSet).Remove(s)
			}
		}
	}
}

func (sp *SharePublisher) openHandler(ctx context.Context) {

	defer func() {
		if err := recover(); err != nil {
			sp.owner.log.Debug("%s : dispose slow event has error : %s", sp.topic, err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case e := <-sp.queue:
			sp.owner.log.Debug("%s : handler receive slow event : %#v", sp.topic, e)
			sp.notifySubscriber(e)
		}
	}
}

func (sp *SharePublisher) notifySubscriber(event Event) {
	topic := event.Name()
	if set, exist := sp.listeners.Load(topic); exist {
		set.(*polerpc.ConcurrentSet).Range(func(value interface{}) {
			value.(Subscriber).OnEvent(event)
		})
	}

}
