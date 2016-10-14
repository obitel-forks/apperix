package apperix

import (
	"fmt"
	"net"
	"strconv"
	"crypto/tls"
)

/*
	newHttpsListener creates a new listener for incoming HTTPS requests
	and sets up a passive parallel goroutine
	which stops the listener when shutdownSignal is approached
*/
func newHttpsListener(
	addr string,
	port uint16,
	shutdownSignal chan int,
	tlsConfig *tls.Config,
) (
	newListener net.Listener,
	err error,
) {
	newListener, err = tls.Listen(
		"tcp",
		net.JoinHostPort(addr, strconv.Itoa(int(port))),
		tlsConfig,
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