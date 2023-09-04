package redis

import (
	"api/pkg/ctrader"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/nitishm/go-rejson/v4"
	goredis "github.com/redis/go-redis/v9"
)

var redisAddr string = "redis:6379"
var databaseID int = 0

var positionUpdateKey string = "positionUpdates"
var MonitorUpdateKey string = "monitorUpdates"

type RedisClientWithContext struct {
	RDB *goredis.Client
	Ctx context.Context
	RH  *rejson.Handler
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
	opts := &goredis.Options{
		Addr:     address,
		DB:       database,
		Password: "",
	}
	rdb := goredis.NewClient(opts)
	if len(password) != 0 {
		opts.Password = password[0]

	}
	rh := rejson.NewReJSONHandler()
	ctx := context.Background()
	rh.SetGoRedisClientWithContext(ctx, rdb)
	return &RedisClientWithContext{
		RDB: rdb,
		Ctx: ctx,
		RH:  rh,
	}
}

func (c *RedisClientWithContext) ComparePositions(storageSetName string, currentSet []string) ([]string, []string, error) {
	var diffInStorage []string
	var diffInCurrent []string
	//if there are no positions, then will not return an empty set
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
	if len(diffInCurrent) > 0 || len(diffInStorage) > 0 {
		err := c.overrideSet(storageSetName, currentSet)
		if err != nil {
			// log.Println("diff in current = ", diffInCurrent)
			// log.Println("diff in storage = ", diffInStorage)

			log.Fatal(err)
		}
	}
	return diffInStorage, diffInCurrent, nil
}

func (c *RedisClientWithContext) overrideSet(setKey string, members []string) error {
	// return nil
	log.Println("current set = ", members)
	tx := c.RDB.TxPipeline()
	tx.Del(c.Ctx, setKey)
	if len(members) != 0 {

		tx.SAdd(c.Ctx, setKey, members)
	}

	_, err := tx.Exec(c.Ctx)
	if err != nil {
		return err
	}
	return nil
}

func (c *RedisClientWithContext) PushPositionUpdate(data interface{}) error {
	_, err := c.RDB.LPush(c.Ctx, positionUpdateKey, data).Result()
	if err != nil {
		return err
	}
	return nil
}

func (c *RedisClientWithContext) PopPositionUpdate() (*ctrader.CtraderMonitorMessage, error) {
	res, err := c.RDB.LPop(c.Ctx, positionUpdateKey).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil, nil
		}
		return nil, err
	}
	message := &ctrader.CtraderMonitorMessage{}
	err = json.Unmarshal([]byte(res), message)
	if err != nil {
		log.Fatal(err)
	}

	return message, nil
}

// func SetKey
func (c *RedisClientWithContext) PushUpdate(updateKey string, data interface{}) error {

	_, err := c.RDB.LPush(c.Ctx, updateKey, data).Result()
	if err != nil {
		return err
	}
	return nil
}

func (c *RedisClientWithContext) PopUpdate(updateKey string) ([]byte, error) {

	res, err := c.RDB.LPop(c.Ctx, updateKey).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil, nil
		}
		return nil, err
	}
	return []byte(res), err
}
