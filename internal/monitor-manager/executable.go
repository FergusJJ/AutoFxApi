package monitormanager

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

type Executable struct {
	Path        string
	Args        []string
	cmd         *exec.Cmd
	statusMutex sync.Mutex
	running     bool
}

func NewExecutable(path string, args ...string) *Executable {
	cmd := exec.CommandContext(context.Background(), path, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return &Executable{
		Path:        path,
		Args:        args,
		cmd:         cmd,
		statusMutex: sync.Mutex{},
		running:     false,
	}

}

func (e *Executable) Start() error {
	e.statusMutex.Lock()
	defer e.statusMutex.Unlock()
	if e.running {
		return fmt.Errorf("executable is already running")
	}
	err := e.cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start the executable: %w", err)
	}
	e.running = true
	go func() {
		err := e.cmd.Wait()
		if err != nil {
			if err != nil {
				log.Printf("Executable at '%s' exited with error: %s\n", e.Path, err)
			} else {
				log.Printf("Executable at '%s' completed successfully\n", e.Path)
			}
		}
		e.statusMutex.Lock()
		defer e.statusMutex.Unlock()
		e.running = false
	}()

	return nil
}

func (e *Executable) Restart() error {
	err := e.Stop()
	if err != nil {
		return err
	}
	return e.Start()
}

func (e *Executable) Stop() error {
	e.statusMutex.Lock()
	defer e.statusMutex.Unlock()

	if !e.running {
		return fmt.Errorf("executable is not running")
	}

	err := e.cmd.Process.Signal(syscall.SIGTERM)
	if err != nil {
		return fmt.Errorf("failed to stop the executable: %w", err)
	}
	e.running = false

	return nil
}
