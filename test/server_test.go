package test

import (
	"github.com/XiaCo/ucp/protocol"
	"github.com/XiaCo/ucp/server"
	"github.com/XiaCo/ucp/utils/log"
	"os"
	"testing"
)

func TestPullFile(t *testing.T) {
	log.Println("init") // 初始化log，否则报空指针错误
	go server.UDPServer("0.0.0.0:56789")
	downloadTask, newErr := protocol.NewClientTask("127.0.0.1:56789", "./pulled.log", "./pulltest.tar")
	if newErr != nil {
		t.Error(newErr)
	}
	downloadTask.Pull(10)
	os.Remove("./pulled.log")
}
