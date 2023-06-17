package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var redisAddr string = "127.0.0.1:6379"
var databaseID int = 0

type RedisClientWithContext struct {
	RDB *redis.Client
	Ctx context.Context
}

func RedisInitialise() (*RedisClientWithContext, func(), error) {
	//if no errors, return function that will close the connection
	Client := RedisGetClient(redisAddr, databaseID)

	return Client, func() {
		err := Client.RDB.Close()
		if err != nil {
			log.Println("Error closing Redis connection:", err)
		} else {
			log.Println("Redis connection closed successfully")
		}
	}, nil
}

func RedisGetClient(address string, database int, password ...string) *RedisClientWithContext {
	opts := &redis.Options{
		Addr:     address,
		DB:       database,
		Password: "",
	}
	if len(password) != 0 {
		opts.Password = password[0]
	}

	return &RedisClientWithContext{
		RDB: redis.NewClient(opts),
		Ctx: context.Background(),
	}
}

func (c *RedisClientWithContext) ComparePositions(storageSetName string, currentSet []string) ([]string, []string, error) {
	var diffInStorage []string
	var diffInCurrent []string

	// Retrieve all members of the storage set
	membersCmd := c.RDB.SMembers(context.TODO(), storageSetName)
	storageMembers, err := membersCmd.Result()
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving set members: %v", err)
	}

	// Check members in storage set that are not in current set
	for _, member := range storageMembers {
		found := false
		for _, currentMember := range currentSet {
			if member == currentMember {
				found = true
				break
			}
		}
		if !found {
			diffInStorage = append(diffInStorage, member)
		}
	}

	// Check members in current set that are not in storage set
	for _, member := range currentSet {
		found := false
		for _, storageMember := range storageMembers {
			if member == storageMember {
				found = true
				break
			}
		}
		if !found {
			diffInCurrent = append(diffInCurrent, member)
		}
	}

	return diffInStorage, diffInCurrent, nil
}

// func SetKey
