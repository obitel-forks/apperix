package apperix

import (
	"fmt"
	"net"
	"strconv"
)

/*
	newHttpListener creates a new listener for incoming HTTP requests
	and sets up a passive parallel goroutine
	which stops the listener when shutdownSignal is approached
*/
func newHttpListener(
	addr string,
	port uint16,
	shutdownSignal chan int,
) (
	newListener net.Listener,
	err error,
) {
	newListener, err = net.Listen(
		"tcp",
		net.JoinHostPort(addr, strconv.Itoa(int(port))),
	)
	if err != nil {
		return nil, fmt.Errorf("Could not create listener: %s", err)
	}
	go func() {
		//block until shutdown signal's received
		<- shutdownSignal
		//close listener, stop accepting new connections
		newListener.Close()
	}()
	return newListener, nil
}