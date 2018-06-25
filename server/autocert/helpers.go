package autocert

import (
	"strconv"
	"strings"
)

func splitHostPort(hostPort string, defaultPort int) (host string, port int, err error) {

	colonIndex := strings.Index(hostPort, ":")
	if colonIndex == -1 {
		host = hostPort
		port = defaultPort
		return
	}

	host = hostPort[:colonIndex]
	portString := hostPort[colonIndex+1:]
	port, err = strconv.Atoi(portString)
	if err != nil {
		return
	}

	return
}
