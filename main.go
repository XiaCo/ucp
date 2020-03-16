package main

import (
	"flag"
	"fmt"
	"github.com/XiaCo/ucp/protocol"
	"github.com/XiaCo/ucp/server"
	cmd "github.com/XiaCo/ucp/utils/flag"
	logging "github.com/XiaCo/ucp/utils/log"
	"os"
	"strings"
)

func init() {
	cmd.Init()
	logging.Init(*cmd.LogPath) // init logger
}

func main() {

	if *cmd.H || *cmd.Help {
		flag.Usage()
		return
	}

	if *cmd.S {
		server.UDPServer(*cmd.ServerAddr)
		return
	}

	if *cmd.CP != "" {
		s := strings.Split(*cmd.CP, " ")
		remoteAddr, filePath, savePath := s[0], s[1], s[2]
		downloadTask, newErr := protocol.NewClientTask(remoteAddr, savePath, filePath)
		if newErr != nil {
			fmt.Fprint(os.Stderr, newErr)
			return
		}
		downloadTask.Pull(*cmd.Speed)
		return
	}

	fmt.Fprintf(os.Stderr, "Parameter parsing error\n")
	flag.Usage()

}
