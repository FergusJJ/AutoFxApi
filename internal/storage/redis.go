package storage

import "log"

func RedisInitialise() (func(), error) {
	//if no errors, return function that will close the connection
	return func() {
		log.Println("closing connection")
	}, nil
}
