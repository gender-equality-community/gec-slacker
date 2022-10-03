package main

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/gender-equality-community/types"
	"github.com/go-redis/redis/v9"
)

var (
	validMessage = map[string]interface{}{
		"id":  "foo",
		"ts":  0,
		"msg": "hi there",
	}

	emptyMessage = map[string]interface{}{}

	badDataTypes = map[string]interface{}{
		"id":  []byte("foo"),
		"ts":  0,
		"msg": "hi there",
	}
)

type dummyRedis struct {
	msg              map[string]interface{}
	stream           string
	readGroupError   bool
	busyError        bool
	createGroupError bool
	addError         bool
	callCount        int
}

func (d *dummyRedis) XGroupCreate(context.Context, string, string, string) *redis.StatusCmd {
	var err error
	if d.createGroupError {
		err = fmt.Errorf("an error")
	} else if d.busyError {
		err = fmt.Errorf("BUSYGROUP Consumer Group name already exists")
	}

	return redis.NewStatusResult("1", err)
}

func (d *dummyRedis) XReadGroup(context.Context, *redis.XReadGroupArgs) *redis.XStreamSliceCmd {
	var err error
	if d.readGroupError {
		err = fmt.Errorf("an error")
	}

	d.callCount++

	// Hang after first call, otherwise we might just end up sending
	// the same message a gazillion times
	if d.callCount > 1 {
		for {
		}
	}

	return redis.NewXStreamSliceCmdResult([]redis.XStream{
		{
			Messages: []redis.XMessage{
				{
					Values: d.msg,
				},
			},
		},
	}, err)
}

func (d *dummyRedis) XAdd(_ context.Context, m *redis.XAddArgs) *redis.StringCmd {
	var err error
	if d.addError {
		err = fmt.Errorf("an error")
	} else {
		d.msg = m.Values.(map[string]interface{})
	}

	d.stream = m.Stream

	return redis.NewStringResult("", err)
}

func TestNewRedis(t *testing.T) {
	_, err := NewRedis("example.com:6379", "test-in", "test-out")
	if err != nil {
		t.Errorf("unexpected error: %#v", err)
	}
}

func TestRedis_Process(t *testing.T) {
	for _, test := range []struct {
		name        string
		r           redisClient
		expect      *types.Message
		expectError bool
	}{
		{"Creating XGroup errors bubble up", &dummyRedis{createGroupError: true}, nil, true},
		{"Reading XGroup errors bubble up", &dummyRedis{readGroupError: true}, nil, true},
		{"Parse errors bubble up", &dummyRedis{readGroupError: true}, nil, true},
		{"Messages with empty fields should be skipped", &dummyRedis{msg: emptyMessage}, nil, false},
		{"Messages with unepected field types should error", &dummyRedis{msg: badDataTypes}, nil, true},
		{"Valid messages, on a new XGroup, should succeed", &dummyRedis{msg: validMessage}, &types.Message{ID: "foo", Timestamp: 0, Message: "hi there"}, false},
		{"Valid messages, on a existing XGroup, should succeed", &dummyRedis{msg: validMessage, busyError: true}, &types.Message{ID: "foo", Timestamp: 0, Message: "hi there"}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := Redis{
				client:    test.r,
				inStream:  "test-in",
				outStream: "test-out",
				id:        "foo",
			}

			c := make(chan types.Message)

			go func() {
				err := r.Process(c)
				if err != nil && !test.expectError {
					t.Errorf("unexpected error: %v", err)
				} else if err == nil && test.expectError {
					t.Error("expected error")
				}
			}()

			// This is a little gross, but fine for now.
			// Ideally we'd use a waitgroup and some kind of
			// shutdown signal in r.Process, but adding a load
			// of complex logic to support tests to shave milliseconds
			// isn't the kind of efficiency gain we need right now
			time.Sleep(100 * time.Millisecond)

			if test.expect != nil {
				m := <-c
				t.Run("message content", func(t *testing.T) {
					if !reflect.DeepEqual(*test.expect, m) {
						t.Errorf("expected\n%#v\nreceived\n%#v", *test.expect, m)
					}
				})
			}
		})
	}
}
