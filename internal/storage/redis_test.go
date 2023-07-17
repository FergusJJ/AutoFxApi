package storage_test

import (
	"api/internal/storage"
	"encoding/json"
	"log"
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
	log.Println(closedPositions)
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

func Test_set_json(t *testing.T) {

	type Object struct {
		Name     string `json:"name"`
		Age      int    `json:"age"`
		Verified bool   `json:"verified"`
	}

	client := storage.RedisGetClient(address, databaseID)
	// data1 := Object{
	// 	Name:     "fergus1",
	// 	Age:      19,
	// 	Verified: true,
	// }
	data1 := Object{
		Name:     "fergus1as",
		Age:      192,
		Verified: false,
	}

	jsonBytes, err := json.Marshal(data1)
	if err != nil {
		t.Fatal(err)
	}
	// jsonBytes2, err := json.Marshal(data2)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// err = client.PushToList("positionUpdates", jsonBytes)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	err = client.PushPositionUpdate(jsonBytes)
	if err != nil {
		t.Fatal(err)
	}
	// res, err := client.PopPositionUpdate()
	// if err != nil && err.Error() != "redis: nil" {
	// 	t.Fatal(err)
	// }
	// log.Printf("Result:\n---\n%s\n---\n", res)
}

func Test_Push_Pop_Position_Update(t *testing.T) {

	type MonitorTypes string
	type Object struct {
		Name   string       `json:"name"`
		Type   MonitorTypes `json:"type"`
		Option int          `json:"option"`
	}

	var (
		ICMARKETS MonitorTypes = "icmarkets"
	)

	client := storage.RedisGetClient(address, databaseID)

	data1 := &Object{
		Name:   "test_test",
		Type:   ICMARKETS,
		Option: 1,
	}

	jsonBytes, err := json.Marshal(data1)
	if err != nil {
		t.Fatal(err)
	}

	err = client.PushUpdate(storage.MonitorUpdateKey, jsonBytes)
	if err != nil {
		t.Fatal(err)
	}
	resBytes, err := client.PopUpdate(storage.MonitorUpdateKey)
	if err != nil && err.Error() != "redis: nil" {
		t.Fatal(err)
	}
	obj := &Object{}
	json.Unmarshal(resBytes, obj)
	// resStr := string(resBytes)

	log.Printf("Result:\n---\n%+v\n---\n", obj)

}
