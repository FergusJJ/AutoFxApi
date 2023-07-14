package monitormanager

import (
	"api/internal/storage"
	"encoding/json"
	"log"
	"time"
)

type MonitorManagerMessage struct {
	Name   string       `json:"name"`
	Type   MonitorTypes `json:"type"`
	Option int          `json:"option"`
}

var OptionsMap = map[string]int{
	"START": 1,
	"STOP":  0,
}

type MonitorTypes string

var (
	ICMARKETS MonitorTypes = "icmarkets"
)

func Manage(client *storage.RedisClientWithContext) error {
	pollInterval := time.NewTicker(time.Millisecond * 500)
	for {
		select {
		case <-pollInterval.C:
			//pull any updates from redis
			updateBytes, err := client.PopUpdate("monitorQueue")
			if err != nil {
				return err
			}
			if len(updateBytes) == 0 {
				log.Println("no updates")
				continue
			}
			message := &MonitorManagerMessage{}
			json.Unmarshal(updateBytes, message)
			switch message.Option {
			case OptionsMap["START"]:
				NewMonitor(message)
			case OptionsMap["STOP"]:
				CloseMonitor(message)
			default:
				continue
			}
		}
	}
}

func NewMonitor(message *MonitorManagerMessage) {
	log.Printf("got start message: %+v", message)

}

func CloseMonitor(message *MonitorManagerMessage) {
	log.Printf("got stop message: %+v", message)

}
