package main

import (
	"log"
	"os/exec"
)

func main() {
	monitorExec()

	// time.Sleep(10 * time.Second)
}

func monitorExec() {
	cmd := exec.Command("dist/exec_monitor.exe", "--headless=true", "--user=7venWwvj")
	if err := cmd.Start(); err != nil {
		log.Fatalf("cmd.Start: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			log.Printf("Exit Status: %d", exiterr.ExitCode())
		} else {
			log.Fatalf("cmd.Wait: %v", err)
		}
	}

}
