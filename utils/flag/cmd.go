package flag

import (
	"flag"
	"fmt"
)

// 实际中应该用更好的变量名
var (
	Help *bool
	H    *bool

	S          *bool
	ServerAddr *string

	CP    *string
	Speed *uint64

	LogPath *string
)

func Init() {
	Help = flag.Bool("help", false, "help for ucp")
	H = flag.Bool("h", false, "help for ucp")

	S = flag.Bool("s", false, "run in server")
	ServerAddr = flag.String("saddr", "0.0.0.0:56789", "file sender listen to binding address")

	CP = flag.String("cp", "", "format: \"remoteIP:port filePath savePath\"\nexample: \"22.22.22.22:56789 /home/test.zip ./test.zip\"")
	Speed = flag.Uint64("speed", 1024, "It is recommended to fill in the minimum bandwidth download / upload speed at both ends\nunit: KB/s\n")

	LogPath = flag.String("o", "", "log path")

	flag.Usage = func() {
		fmt.Println("ucp version: 1.0.0\nA transport tool using UDP underlying protocol")
		flag.PrintDefaults()
	}

	flag.Parse()
}
