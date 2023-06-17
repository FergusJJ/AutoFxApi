package storage_test

import (
	"api/internal/storage"
	"testing"
)

var address = "127.0.0.1:6379"
var databaseID = 0

func Test_redis_initialise(t *testing.T) {
	client := storage.RedisGetClient(address, databaseID)
	client.RDB.Do(client.Ctx)
	pong, err := client.RDB.Ping(client.Ctx).Result()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(pong)

}

func Test_compare_positions(t *testing.T) {

	//testStoragePositions = 1, 2, 3
	client := storage.RedisGetClient(address, databaseID)
	closedPositions, newPositions, err := client.ComparePositions("testStoragePositions", []string{"1", "2"})
	if err != nil {
		t.Fatal(err)
	}
	if len(closedPositions) == 0 {
		t.Fail()
	}
	if closedPositions[0] != "3" {
		t.Fail()
	}
	if len(newPositions) != 0 {
		t.Fail()
	}
	closedPositions, newPositions, err = client.ComparePositions("testStoragePositions", []string{"4"})
	if err != nil {
		t.Fatal(err)
	}
	if len(closedPositions) == 0 || len(newPositions) == 0 {
		t.Fail()
	}
	if closedPositions[0] != "1" || closedPositions[1] != "2" || closedPositions[2] != "3" {
		t.Fail()
	}
	if newPositions[0] != "4" {
		t.Fail()
	}

}
