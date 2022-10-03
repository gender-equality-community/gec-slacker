package main

import (
	"context"
	"fmt"

	"github.com/gender-equality-community/types"
	redis "github.com/go-redis/redis/v9"
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

type redisClient interface {
	XGroupCreate(context.Context, string, string, string) *redis.StatusCmd
	XReadGroup(context.Context, *redis.XReadGroupArgs) *redis.XStreamSliceCmd
	XAdd(context.Context, *redis.XAddArgs) *redis.StringCmd
}

type Redis struct {
	client    redisClient
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

func (r Redis) Process(c chan types.Message) (err error) {
	ctx := context.Background()

	err = r.client.XGroupCreate(ctx, r.inStream, groupName, "$").Err()
	if err != nil && err.Error() != busyGroupErr {
		return err
	}

	var (
		entries []redis.XStream
		msg     types.Message
	)

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

		msg, err = types.ParseMessage(entries[0].Messages[0].Values)
		if err != nil {
			// log messages, but don't stop the world
			fmt.Printf("An error processing message: %#v", err)

			continue
		}

		c <- msg
	}

	return
}

func (r Redis) Produce(id, msg string) (err error) {
	return r.client.XAdd(context.Background(), &redis.XAddArgs{
		Stream: r.outStream,
		Values: types.NewMessage(types.SourceSlack, id, msg).Map(),
	}).Err()
}
