package monitormanager

import (
	"api/internal/storage"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Manager struct {
	signalCh    chan os.Signal
	Executables map[string]*Executable
}

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

func NewManager() *Manager {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	return &Manager{
		Executables: make(map[string]*Executable),
		signalCh:    signalCh,
	}
}

func (m *Manager) StopAll() {
	for id, executable := range m.Executables {
		err := executable.Stop()
		if err != nil {
			log.Printf("Failed to stop executable: %s\n", err)
			continue
		}
		delete(m.Executables, id)
	}
}

func (m *Manager) Manage(client *storage.RedisClientWithContext) error {
	pollInterval := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-pollInterval.C:
			//pull any updates from redis
			updateBytes, err := client.PopUpdate(storage.MonitorUpdateKey)
			if err != nil && err.Error() != "redis: nil" {
				return err
			}
			if len(updateBytes) == 0 {
				log.Print("no updates")
				continue
			}
			message := &MonitorManagerMessage{}
			json.Unmarshal(updateBytes, message)
			switch message.Option {
			case OptionsMap["START"]:
				_, ok := m.Executables[message.Name]
				if ok {
					log.Print("already monitoring: ", message.Name)
					continue
				}
				exe := NewExecutable("/usr/src/fxapi/cmd/monitor.exe", message.Name)
				m.Executables[message.Name] = exe
				err = exe.Start()
				if err != nil {
					log.Print(err)
				}

			case OptionsMap["STOP"]:
				_, ok := m.Executables[message.Name]
				if !ok {
					log.Print("monitor does not exist, cannot close: ", message.Name)
					continue
				}
				err = m.Executables[message.Name].Stop()
				if err != nil {
					log.Print(err)
				}
				if !m.Executables[message.Name].running {
					delete(m.Executables, message.Name)
					log.Print("successfully stopped", message.Name)
				}

			default:
				continue
			}
		case <-m.signalCh:
			//stop all monitors
			log.Println("got signal, stopping all monitors")
			m.StopAll()
			return nil
		}
	}
}
