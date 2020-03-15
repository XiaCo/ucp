package server

import (
	"github.com/XiaCo/ucp/protocol"
	"github.com/XiaCo/ucp/utils/log"
	"net"
)

var (
	addrMap       = make(map[string]*protocol.ServerTask)
	ClientMessage = make(chan Package, 128) // 多余的包会被丢弃
)

func handle(conn *net.UDPConn, addr *net.UDPAddr, b []byte) {
	remoteAddr := addr.String()
	if _, ok := addrMap[remoteAddr]; !ok {
		serverTask := protocol.NewServerTask(conn, addr)
		addrMap[remoteAddr] = serverTask
		log.Printf("request addr:%s\n", remoteAddr)
		go serverTask.WritePackage()
	}
	buf := make([]byte, len(b))
	copy(buf, b)
	select {
	case ClientMessage <- Package{buf, addr}:
	default:
		break
	}
}

func ReadAndDealPackage() {
	for {
		select {
		case pk := <-ClientMessage:
			pServerTask := addrMap[pk.Addr.String()]
			pServerTask.DealBuffer(pk.Msg)
		}
	}
}

func UDPServer(addr string) {
	log.Println("udpserver start")
	go ReadAndDealPackage() // 读取接收到的udp包
	udpAddr, resolveErr := net.ResolveUDPAddr("udp4", addr)
	if resolveErr != nil {
		log.Fatalln(resolveErr)
	}
	//监听端口
	udpConn, listenErr := net.ListenUDP("udp", udpAddr)
	if listenErr != nil {
		log.Fatalln(listenErr)
	}
	buf := make([]byte, 1024*32)
	for {
		n, udpRemoteAddr, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			log.Fatalln(err)
		}
		handle(udpConn, udpRemoteAddr, buf[:n])
	}
}
