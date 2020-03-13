package main

import (
	"flag"
	"fmt"
	"github.com/XiaCo/ucp/protocol"
	"github.com/XiaCo/ucp/server"
	"os"
	"strings"
)

// 实际中应该用更好的变量名
var (
	help *bool
	h    *bool

	s          *bool
	serverAddr *string

	cp    *string
	speed *uint64
)

func init() {
	help = flag.Bool("help", false, "help for ucp")
	h = flag.Bool("h", false, "help for ucp")

	s = flag.Bool("s", false, "run in server")
	serverAddr = flag.String("saddr", "0.0.0.0:56789", "file sender listen to binding address")

	cp = flag.String("cp", "", "format: \"remoteIP:port filePath savePath\"\nexample: \"22.22.22.22:56789 /home/test.zip ./test.zip\"")
	speed = flag.Uint64("speed", 1024, "It is recommended to fill in the minimum bandwidth download / upload speed at both ends\nunit: Kb/s\n")

	flag.Usage = func() {
		fmt.Println("ucp version: 1.0.0\nA transport tool using UDP underlying protocol")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if *h || *help {
		flag.Usage()
		return
	}

	if *s {
		server.UDPServer(*serverAddr)
		return
	}

	if *cp != "" {
		s := strings.Split(*cp, " ")
		remoteAddr, filePath, savePath := s[0], s[1], s[2]
		downloadTask, newErr := protocol.NewClientTask(remoteAddr, savePath, filePath)
		if newErr != nil {
			fmt.Fprint(os.Stderr, newErr)
			return
		}
		downloadTask.Pull(*speed)
		return
	}

	fmt.Fprintf(os.Stderr, "Parameter parsing error\n")
	flag.Usage()

}
