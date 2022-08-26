package main

import (
	"context"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/rs/xid"
)

const groupName = "gec-slacker"

var (
	// Because all redis errors are of proto.RedisError it's hard to
	// do a proper error comparison.
	//
	// Instead, the best we can do is compare error strings
	busyGroupErr = "BUSYGROUP Consumer Group name already exists"
)

type Redis struct {
	client    *redis.Client
	inStream  string
	outStream string
	id        string
}

func NewRedis(addr, inStream, outStream string) (r Redis, err error) {
	r.inStream = inStream
	r.outStream = outStream

	r.client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	r.id = xid.New().String()

	return r, nil
}

func (r Redis) Process(c chan Message) (err error) {
	ctx := context.Background()

	err = r.client.XGroupCreate(ctx, r.inStream, groupName, "$").Err()
	if err != nil && err.Error() != busyGroupErr {
		return err
	}

	var entries []redis.XStream
	for {
		entries, err = r.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    groupName,
			Consumer: r.id,
			Streams:  []string{r.inStream, ">"},
			Count:    1,
			Block:    0,
			NoAck:    false,
		}).Result()
		if err != nil {
			break
		}

		msg := entries[0].Messages[0].Values

		c <- Message{
			ID:      msg["id"].(string),
			Ts:      msg["ts"].(string),
			Message: msg["msg"].(string),
		}
	}

	return
}

func (r Redis) Produce(id, msg string) (err error) {
	return r.client.XAdd(context.Background(), &redis.XAddArgs{
		Stream: r.outStream,
		Values: map[string]interface{}{
			"id":  id,
			"ts":  time.Now().Unix(),
			"msg": msg,
		},
	}).Err()
}
