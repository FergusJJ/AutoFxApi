package shutdown

import (
	"os"
	"os/signal"
	"syscall"
)

func Gracefully() {
	quit := make(chan os.Signal, 1)
	defer close(quit)

	//if app stops, channel blcoks
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	//execution blocks
	<-quit
}
