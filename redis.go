package main

import (
	"context"

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
	client *redis.Client
	stream string
	id     string
}

func NewRedis(addr, stream string) (r Redis, err error) {
	r.stream = stream
	r.client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	r.id = xid.New().String()

	return r, nil
}

func (r Redis) Process(c chan Message) error {
	ctx := context.Background()

	err := r.client.XGroupCreate(ctx, r.stream, groupName, "$").Err()
	if err != nil && err.Error() != busyGroupErr {
		return err
	}

	for {
		entries, err := r.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    groupName,
			Consumer: r.id,
			Streams:  []string{r.stream, ">"},
			Count:    1,
			Block:    0,
			NoAck:    false,
		}).Result()
		if err != nil {
			return err
		}

		msg := entries[0].Messages[0].Values

		c <- Message{
			ID:      msg["id"].(string),
			Ts:      msg["ts"].(string),
			Message: msg["msg"].(string),
		}
	}

	return nil
}
