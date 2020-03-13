package server

import "net"

type Package struct {
	Msg  []byte
	Addr *net.UDPAddr
}
