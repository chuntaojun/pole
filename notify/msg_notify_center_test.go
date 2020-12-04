// Copyright (c) 2020, Conf-Group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package notify

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Conf-Group/pole/utils"
)

// test normal event function

type StringEvent struct {
	content string
}

func (s *StringEvent) Name() string {
	return "StringEvent"
}

func (s *StringEvent) Sequence() int64 {
	return time.Now().Unix()
}

type StringEventSubscriber struct {
	fn func(s string)
}

func (s *StringEventSubscriber) OnEvent(event Event) {
	se := event.(*StringEvent)
	s.fn(se.content)
}

func (s *StringEventSubscriber) SubscribeType() Event {
	return &StringEvent{}
}

func (s *StringEventSubscriber) IgnoreExpireEvent() bool {
	return false
}

func Test_PublishEvent(t *testing.T) {

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Test_PublishEvent has error : %s", err)
			t.FailNow()
		}
	}()

	err := RegisterPublisher(&StringEvent{}, 16)
	assert.Nil(t, err, "register publisher should be success, %s", err)
	isOk, err := PublishEvent(&StringEvent{content: "liaochuntao"})
	assert.Nil(t, err, "publish event should be success, %s", err)
	assert.True(t, isOk, "publish event should be success")
}

func Test_PublishEventAndSubscribe(t *testing.T) {

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Test_PublishEventAndSubscribe has error : %s", err)
			t.FailNow()
		}
	}()

	var reference atomic.Value

	err := RegisterPublisher(&StringEvent{}, 16)
	assert.Nil(t, err, "register publisher should be success, %s", err)
	err = RegisterSubscriber(&StringEventSubscriber{fn: func(s string) {
		fmt.Printf("receive content : %s \n", s)
		reference.Store(s)
	}})
	assert.Nil(t, err, "register publisher should be success, %s", err)
	isOk, err := PublishEvent(&StringEvent{content: "liaochuntao"})
	assert.Nil(t, err, "publish event should be success, %s", err)
	assert.True(t, isOk, "publish event should be success")

	time.Sleep(time.Duration(5) * time.Second)

	v := reference.Load()
	assert.NotNil(t, v, "v must not nil")
	assert.Equal(t, "liaochuntao", v.(string), "StringEvent.content must be liaochuntao")
}

// test expire event function

type ExpireEvent struct {
	sequence int64
	content  string
}

func (s *ExpireEvent) Name() string {
	return "ExpireEvent"
}

func (s *ExpireEvent) Sequence() int64 {
	return s.sequence
}

type IgnoreExpireSubscriber struct {
	fn func(s string)
}

func (s *IgnoreExpireSubscriber) OnEvent(event Event) {
	se := event.(*ExpireEvent)
	s.fn(se.content)
}

func (s *IgnoreExpireSubscriber) SubscribeType() Event {
	return &ExpireEvent{}
}

func (s *IgnoreExpireSubscriber) IgnoreExpireEvent() bool {
	return true
}

func Test_IgnoreExpireEvent(t *testing.T) {

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Test_IgnoreExpireEvent has error : %T", err)
			t.FailNow()
		}
	}()

	var reference atomic.Value
	var cnt int64 = 0

	RegisterOnExpire(func(event Event) {
		e := event.(*ExpireEvent)
		atomic.StoreInt64(&cnt, e.sequence)
	})

	err := RegisterPublisher(&ExpireEvent{}, 16)
	assert.Nil(t, err, "register publisher should be success, %s", err)

	err = RegisterSubscriber(&IgnoreExpireSubscriber{fn: func(s string) {
		fmt.Printf("receive content : %s \n", s)
		reference.Store(s)
	}})
	assert.Nil(t, err, "register subscriber should be success, %s", err)

	isOk, err := PublishEvent(&ExpireEvent{sequence: 10, content: "liaochuntao"})
	assert.Nil(t, err, "publish event should be success, %s", err)
	assert.True(t, isOk, "publish event should be success")

	time.Sleep(time.Duration(5) * time.Second)

	v := reference.Load()
	assert.NotNil(t, v, "v must not be nil")
	assert.Equal(t, "liaochuntao", v.(string), "StringEvent.content must be liaochuntao")

	isOk, err = PublishEvent(&ExpireEvent{sequence: 1, content: "liaochuntao2"})

	time.Sleep(time.Duration(5) * time.Second)
	assert.EqualValues(t, 1, cnt)

}

type SlowEventOne struct {
	content string
}

func (s *SlowEventOne) Name() string {
	return "SlowEventOne"
}

type SlowEventTwo struct {
	content string
}

func (s *SlowEventTwo) Name() string {
	return "SlowEventTwo"
}

type SlowEventThree struct {
	content string
}

func (s *SlowEventThree) Name() string {
	return "SlowEventThree"
}

type SlowEventSubscriber struct {
	f func(s string)
}

func (s *SlowEventSubscriber) OnEvent(event Event) {
	switch e := event.(type) {
	case *SlowEventOne:
		fmt.Printf("receive event %s \n", e.content)
		s.f(e.content)
	case *SlowEventTwo:
		fmt.Printf("receive event %s \n", e.content)
		s.f(e.content)
	}
}

func (s *SlowEventSubscriber) SubscribeTypes() []Event {
	return []Event{&SlowEventOne{}, &SlowEventTwo{}}
}

func (s *SlowEventSubscriber) IgnoreExpireEvent() bool {
	return false
}

func Test_PublishSlowEvent(t *testing.T) {

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Test_PublishSlowEvent has error : %+v, %+v\n", err, utils.PrintStack())
			t.FailNow()
		}
	}()

	var reference atomic.Value

	err := RegisterSharePublisher(&SlowEventOne{})
	assert.Nil(t, err, "register publisher should be success, %s", err)
	err = RegisterSharePublisher(&SlowEventTwo{})
	assert.Nil(t, err, "register publisher should be success, %s", err)
	err = RegisterSharePublisher(&SlowEventThree{})
	assert.Nil(t, err, "register publisher should be success, %s", err)

	err = RegisterSubscriber(&SlowEventSubscriber{
		f: func(s string) {
			reference.Store(s)
		},
	})
	assert.Nil(t, err, "register subscriber should be success, %s", err)

	isOk, err := PublishEvent(&SlowEventOne{content: "SlowEventOne"})
	assert.Nil(t, err, "publish event should be success, %s", err)
	assert.True(t, isOk, "publish event should be success")

	time.Sleep(time.Duration(5) * time.Second)

	assert.EqualValues(t, "SlowEventOne", reference.Load().(string))

	reference.Store("")

	isOk, err = PublishEvent(&SlowEventTwo{content: "SlowEventTwo"})
	assert.Nil(t, err, "publish event should be success, %s", err)
	assert.True(t, isOk, "publish event should be success")

	time.Sleep(time.Duration(5) * time.Second)

	assert.EqualValues(t, "SlowEventTwo", reference.Load().(string))

	reference.Store("")

	isOk, err = PublishEvent(&SlowEventThree{content: "SlowEventThree"})
	assert.Nil(t, err, "publish event should be success, %s", err)
	assert.True(t, isOk, "publish event should be success")

	time.Sleep(time.Duration(5) * time.Second)

	assert.Equal(t, "", reference.Load(), "SlowEventThree must not receive")
}
