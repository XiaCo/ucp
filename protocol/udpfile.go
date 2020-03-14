package protocol

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"net"
	"os"
	"sync"
	"time"
)

const (
	UDPBufferSize int64 = 1024 * 32 // 单个udp包大小
	SplitFileSize int64 = 1024 * 1  // 将文件切割成小块，每块的大小  todo 暂测16k无法接受到

	RequestFlag      = 1 // 请求文件包
	SupplyFlag       = 2 // 请求补充包
	CloseFlag        = 3 // 关闭包
	ReplyNumbersFlag = 4 // 回复客户端文件编号包
	ReplyConfirmFlag = 5 // 编号包确认收到
	FileDataFlag     = 6 // 文件数据包
)

type ServerTask struct {
	ReadConnMsg  *UDPFilePackage
	WriteConnMsg *UDPFilePackage
	Buffer       []byte     // 用来读取文件中的数据
	Numbers      chan int64 // 存储当前待传输任务的编号
	Path         string
	Conn         *net.UDPConn
	SendSpeed    int64 // 单位:kb/s
	RemoteAddr   *net.UDPAddr
	File         *os.File
	Ready        chan struct{}
	Over         bool
}

func NewServerTask(conn *net.UDPConn, remoteAddr *net.UDPAddr) *ServerTask {
	st := &ServerTask{
		ReadConnMsg:  &UDPFilePackage{},
		WriteConnMsg: &UDPFilePackage{},
		Buffer:       make([]byte, SplitFileSize),
		Numbers:      nil,
		Path:         "",
		Conn:         conn,
		SendSpeed:    0,
		RemoteAddr:   remoteAddr,
		File:         nil,
		Ready:        make(chan struct{}),
		Over:         false,
	}
	return st
}

func (s *ServerTask) Close() {
	if s.Over { // 防止多次调用
		return
	}
	close(s.Ready)
	closeErr := s.File.Close()
	if closeErr != nil {
		Logger.Println(closeErr)
	}
	s.Over = true
}

func (s *ServerTask) requestInit() {
	// 设置编号并发送文件编号包
	defer func() {
		e := recover()
		if e != nil { // 出错记录并关闭下载任务
			Logger.Println(e)
			s.Close()
		}
	}()
	var numbersLength int64
	if s.Path != "" {
		goto sendNumbers // 已经初始化过的话，直接发送编号包
	}
	{ // 初始化下载任务信息
		s.Path = s.ReadConnMsg.GetPath()
		f, openErr := os.Open(s.Path)
		if openErr != nil {
			Logger.Fatalln(openErr)
		} else {
			s.File = f
		}
		stat, statErr := s.File.Stat() // 初始化请求信息
		if statErr != nil {
			Logger.Fatalln(statErr)
		}
		s.SendSpeed = s.ReadConnMsg.Speed
		numbersLength = int64(len(SplitFile(stat.Size())))
		Logger.Printf("包数：%d\n", numbersLength)
		s.Numbers = make(chan int64, numbersLength)
		for i := int64(0); i < numbersLength; i++ {
			s.Numbers <- i
		}
	}
sendNumbers:
	{ // 给客户端发送编号回复包
		replyMsg := &UDPFilePackage{Ack: ReplyNumbersFlag, Path: s.Path, Number: []int64{int64(len(s.Numbers))}}
		m, marshalErr := proto.Marshal(replyMsg)
		if marshalErr != nil {
			panic(marshalErr)
		} else {
			_, writeUDPErr := s.Conn.WriteToUDP(m, s.RemoteAddr) // 写入需要接收的编号
			if writeUDPErr != nil {
				Logger.Fatalln(writeUDPErr)
			}
		}
	}
}

func (s *ServerTask) dealMsg() {
	// 处理客户端的包信息
	switch s.ReadConnMsg.Ack {
	case RequestFlag:
		s.requestInit()
	case SupplyFlag:
		for _, number := range s.ReadConnMsg.Number {
			s.Numbers <- number
		}
	case ReplyConfirmFlag:
		s.Ready <- struct{}{} // 就绪位，表示文件数据允许写入
	case CloseFlag:
		s.Close()
	}
}

func (s *ServerTask) WritePackage() {
	// 当over有信号或是超过10s待发送区没有编号时，停止
	delay := time.NewTimer(time.Second * 10)
	defer delay.Stop()
	<-s.Ready
	sleep := SleepAfterSendPackage(250, s.SendSpeed)
	for !s.Over {
		select {
		case fileNumber := <-s.Numbers: // 取一个待发编号，取到文件对应数据，并发送
			offset, seekErr := s.File.Seek(SplitFileSize*fileNumber, 0)
			if seekErr != nil {
				Logger.Println(seekErr)
			}
			readSize, readErr := s.File.Read(s.Buffer)
			if readErr != nil {
				Logger.Println(readErr)
			}

			s.WriteConnMsg.Ack = FileDataFlag
			s.WriteConnMsg.Start = offset
			s.WriteConnMsg.Data = s.Buffer[:readSize]
			msg, marshalErr := proto.Marshal(s.WriteConnMsg)
			if marshalErr != nil {
				Logger.Println(marshalErr)
			}
			_, writeUDPErr := s.Conn.WriteToUDP(msg, s.RemoteAddr)
			if writeUDPErr != nil {
				Logger.Println(writeUDPErr)
			}
			sleep()
		case <-delay.C: // 一定时间后，都没有收到补充请求包，待写区一直为空
			s.Close()
			return
		}
		delay.Reset(time.Second * 10)
	}
}

func (s *ServerTask) DealBuffer(buf []byte) {
	err := proto.Unmarshal(buf, s.ReadConnMsg) // 读到的内容转换为Msg结构体
	if err != nil {
		Logger.Println(err)
	}
	s.dealMsg()
}

type ClientTask struct {
	ReadConnMsg   *UDPFilePackage
	WriteConnMsg  *UDPFilePackage
	Buffer        []byte             // 读取udp
	ReadSemaphore chan struct{}      // 每读取一个包，发送一个信号
	RTO           time.Duration      // 客户端到服务端来回一轮的时间，单位纳秒
	Numbers       map[int64]struct{} // 用以判断 1.数据包是否已经写入过了 2.数据包剩余量
	NumbersLock   *sync.Mutex
	SpeedCal      SpeedCalculator
	Conn          *net.UDPConn
	File          *os.File
	RemotePath    string
	Ready         chan struct{}
	Over          bool
}

func NewClientTask(addr string, savePath string, remotePath string) (*ClientTask, error) {
	udpRemoteAddr, resolveUDPAddrErr := net.ResolveUDPAddr("udp", addr)
	if resolveUDPAddrErr != nil {
		return nil, resolveUDPAddrErr
	}
	conn, dialErr := net.DialUDP("udp", nil, udpRemoteAddr)
	if dialErr != nil {
		return nil, dialErr
	}

	if !SavePathIsValid(savePath) {
		return nil, errors.New("error with save path")
	}
	fp, fpErr := os.OpenFile(savePath, os.O_WRONLY|os.O_CREATE, 0444)
	if fpErr != nil {
		return nil, fpErr
	}
	ct := &ClientTask{
		ReadConnMsg:  &UDPFilePackage{},
		WriteConnMsg: &UDPFilePackage{},
		Buffer:       make([]byte, UDPBufferSize),
		Numbers:      nil,
		NumbersLock:  new(sync.Mutex),
		SpeedCal:     NewSpeedCalculator(time.Second),
		Conn:         conn,
		File:         fp,
		RemotePath:   remotePath,
		Ready:        make(chan struct{}),
		Over:         false,
	}
	return ct, nil
}

func (c *ClientTask) requestInit(speedKbs uint64) error {
	// 发送请求文件信息
	pkg := &UDPFilePackage{Ack: RequestFlag, Path: c.RemotePath, Speed: int64(speedKbs)}
	buf, marshalErr := proto.Marshal(pkg)
	if marshalErr != nil {
		return marshalErr
	}
	c.RTO = time.Duration(time.Now().UnixNano()) // 注册起始时间
	_, writeConnErr := c.Conn.Write(buf)
	if writeConnErr != nil {
		return writeConnErr
	}
	return nil
}

func (c *ClientTask) sendReplyConfirm() {
	c.WriteConnMsg.Ack = ReplyConfirmFlag
	err := c.flushWrite()
	if err != nil {
		Logger.Println(err)
	}
}

func (c *ClientTask) dealMsg() {
	switch c.ReadConnMsg.Ack {
	case ReplyNumbersFlag:
		fmt.Printf("the total package of file: %d\n", c.ReadConnMsg.Number[0])
		c.Numbers = make(map[int64]struct{}, c.ReadConnMsg.Number[0])
		c.RTO = time.Duration(time.Now().UnixNano()) - c.RTO // 算出发出请求到接收到编号包的时间
		//fmt.Printf("rto为：%f秒\n", float64(c.RTO)/1e9)
		for i := int64(0); i < c.ReadConnMsg.Number[0]; i++ {
			c.Numbers[i] = struct{}{}
		}
		c.ReadSemaphore = make(chan struct{}, 1024)
		c.Ready <- struct{}{}
		c.sendReplyConfirm()
	case FileDataFlag:
		num := c.ReadConnMsg.Start / SplitFileSize
		if _, exist := c.Numbers[num]; !exist {
			return
		}
		_, seekErr := c.File.WriteAt(c.ReadConnMsg.Data, c.ReadConnMsg.Start) // 写入文件
		if seekErr != nil {
			Logger.Println(seekErr)
		} else {
			c.NumbersLock.Lock()
			delete(c.Numbers, num)
			c.NumbersLock.Unlock()
			c.ReadSemaphore <- struct{}{}
			c.SpeedCal.AddFlow(SplitFileSize / 1024)
			if len(c.Numbers) == 0 {
				c.close()
			}
		}
	}
}

func (c *ClientTask) readBuffer() {
	n, readErr := c.Conn.Read(c.Buffer)
	if readErr != nil {
		Logger.Println(readErr)
	}
	unmarshalErr := proto.Unmarshal(c.Buffer[:n], c.ReadConnMsg)
	if unmarshalErr != nil {
		Logger.Println(unmarshalErr)
	}
}

func (c *ClientTask) readAndDealMsg() {
	// 读取服务端信息并进行处理
	c.readBuffer()
	c.dealMsg()
}

func (c *ClientTask) flushWrite() (err error) {
	defer func() {
		if e := recover(); e != nil {
			switch t := e.(type) {
			case string:
				err = errors.New(t)
			case error:
				err = t
			default:
				err = nil
			}
		}
	}()
	buf, marshalErr := proto.Marshal(c.WriteConnMsg)
	if marshalErr != nil {
		panic(marshalErr)
	} else {
		_, writeErr := c.Conn.Write(buf)
		if writeErr != nil {
			panic(writeErr)
		}
	}
	return nil
}

func (c *ClientTask) sendAllWillDownloadNumbers() {
	// 发送未接收成功的包的编号

	// copy old numbers
	c.NumbersLock.Lock()
	copyNumbers := make(map[int64]struct{}, len(c.Numbers))
	for n := range c.Numbers {
		copyNumbers[n] = struct{}{}
	}
	c.NumbersLock.Unlock()

	numbers := [1000]int64{}
	s := numbers[:0]
	for fileNumber := range copyNumbers {
		if len(s) == 1000 { // 分组并发送
			c.WriteConnMsg.Ack = SupplyFlag
			c.WriteConnMsg.Number = s
			if err := c.flushWrite(); err != nil {
				Logger.Println(err)
			}
			time.Sleep(time.Millisecond)
			s = numbers[:0]
		} else {
			s = append(s, fileNumber)
		}
	}
	if len(s) != 0 { // 发送最后一个分组
		c.WriteConnMsg.Ack = SupplyFlag
		c.WriteConnMsg.Number = s
		if err := c.flushWrite(); err != nil {
			Logger.Println(err)
		}
	}
}

func (c *ClientTask) sendOver() {
	// 发送关闭包
	c.WriteConnMsg.Ack = CloseFlag
	buf, err := proto.Marshal(c.WriteConnMsg)
	if err != nil {
		Logger.Println(err)
	}
	_, writeErr := c.Conn.Write(buf)
	if writeErr != nil {
		Logger.Println(writeErr)
	}
}

func (c *ClientTask) close() {
	// 关闭下载，释放资源
	if c.Over {
		return
	}
	c.sendOver()
	c.Over = true
	close(c.ReadSemaphore)
	close(c.Ready)
	c.SpeedCal.Close()
	fileCloseErr := c.File.Close()
	if fileCloseErr != nil {
		Logger.Println(fileCloseErr)
	}
	udpCloseErr := c.Conn.Close()
	if udpCloseErr != nil {
		Logger.Println(udpCloseErr)
	}
}

func (c *ClientTask) GetSpeed() int64 {
	return c.SpeedCal.GetSpeed()
}

func (c *ClientTask) GetProgress() int64 {
	c.NumbersLock.Lock()
	n := len(c.Numbers)
	c.NumbersLock.Unlock()
	return int64(n)
}

func (c *ClientTask) PrintDownloadProgress(wg *sync.WaitGroup) {
	delay := time.NewTicker(time.Second)
	defer delay.Stop()
	clear := "\r                                                                       "
	for !c.Over {
		select {
		case <-delay.C:
			speed := c.GetSpeed()
			willDownload := c.GetProgress()
			fmt.Print(clear)
			fmt.Printf("\rcurrent speed: %d kb/s\t\twill download: %d kb", speed, willDownload)
		}
	}
	fmt.Print(clear)
	fmt.Printf("\rfile was downloaded, save path: %s\n", c.File.Name())
	wg.Done()
}

func (c *ClientTask) Pull(speedKBS uint64) {
	// 向服务端请求数据，并接收数据
	wg := sync.WaitGroup{}
	wg.Add(1)
	go c.readAndDealMsg()
	{
		timeout := time.After(time.Second * 10)
		requestRetry := time.NewTicker(time.Second * 2)
		defer requestRetry.Stop()
	loop:
		for i := 0; i < 5; i++ {
			select {
			case <-c.Ready:
				fmt.Println("Ready to receive data")
				break loop
			case <-requestRetry.C:
				reqErr := c.requestInit(speedKBS)
				if reqErr != nil {
					Logger.Println(reqErr)
				}
			case <-timeout:
				fmt.Println("Request timed out, please request the task from the server again")
				c.close()
				return
			}
		}
	}
	go c.PrintDownloadProgress(&wg)

	go func() { // 从udp一直读取
		for !c.Over {
			c.readAndDealMsg()
		}
	}()

	{ // 控制udp超时，在规定时间内未读到服务端的包
		rtt := c.RTO * 2
		delay := time.NewTimer(rtt)
		for !c.Over {
			select {
			case <-delay.C:
				c.sendAllWillDownloadNumbers()
			case <-c.ReadSemaphore:
			}
			delay.Reset(rtt)
		}
	}
	wg.Wait()
}
