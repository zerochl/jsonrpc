package main

import (
	"fmt"
	//	"io"
	"jsonrpc/manager"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const (
	SEND_MSG_CODE_OK = 1
	INTERVAL_END     = "###$$$%%%$$$###"
	//CMD
	START_CONNECT      = "START_CONNECT"
	HEART_BEAT         = "HEART_BEAT"
	HEART_TIME         = 20
	HEART_RECEIVE_TIME = 8
)

//与browser相关的conn
type server struct {
	conn         net.Conn
	er           chan bool
	writ         chan bool
	reconnect    chan bool
	closeWrite   chan bool
	closeHandler chan bool
	recv         chan string
	send         chan string
}

var quitSemaphore chan bool
var closeCountDown int

//var writeStr chan int

func main() {
	newConnect()
}

func newConnect() {
connect:
	tcpAddr, _ := net.ResolveTCPAddr("tcp", "104.224.174.229:8082")
	//	tcpAddr, _ := net.ResolveTCPAddr("tcp", "192.168.0.253:8085")
	conn, _ := net.DialTCP("tcp", nil, tcpAddr)
	defer conn.Close()
	recv := make(chan string)
	send := make(chan string)
	er := make(chan bool, 1)
	writ := make(chan bool)
	reconnect := make(chan bool)
	closeWrite := make(chan bool)
	closeHandler := make(chan bool)
	server := &server{conn, er, writ, reconnect, closeWrite, closeHandler, recv, send}
	//告诉服务器，开始建立连接了
	sendToServer(conn, START_CONNECT)
	go server.newRead()
	go server.newHandler()
	go server.newWrite()
	if <-reconnect {
		goto connect
	}
}

func (self server) newRead() {
	//isheart与timeout共同判断是不是自己设定的SetReadDeadline
	var isheart bool = false
	//20秒发一次心跳包
	self.conn.SetReadDeadline(time.Now().Add(time.Second * HEART_TIME))
	for {
		data := make([]byte, 0) //此处做一个输入缓冲以免数据过长读取到不完整的数据
		buf := make([]byte, 128)
		for {
			//分段读取
			n, err := self.conn.Read(buf)
			if err != nil {
				if strings.Contains(err.Error(), "timeout") && !isheart {
					//					log.Println("发送心跳包")
					sendToServer(self.conn.(*net.TCPConn), HEART_BEAT)
					//4秒时间收心跳包
					self.conn.SetReadDeadline(time.Now().Add(time.Second * HEART_RECEIVE_TIME))
					isheart = true
					break
				}
				self.doReconnect()
				return
			}
			data = append(data, buf[:n]...)
			//		log.Println("data:" + string(data))
			if strings.HasSuffix(string(data), INTERVAL_END) {
				break
			}
		}
		if len(data) == 0 {
			continue
		}
		//		log.Println("收到数据", string(data))
		readDataList := strings.Split(string(data[0:len(data)-len(INTERVAL_END)]), INTERVAL_END)
		for _, readDataItem := range readDataList {
			//			log.Println("读取到一条数据：", string(readDataItem))
			if strings.Compare(HEART_BEAT, string(readDataItem)) == 0 {
				//属于心跳
				log.Println("receive heart")
				self.conn.SetReadDeadline(time.Now().Add(time.Second * HEART_TIME))
				isheart = false
				continue
			}
			self.recv <- readDataItem
		}
	}
}

func (self server) newWrite() {
	for {
		var send string

		select {
		case send = <-self.send:
			sendToServer(self.conn.(*net.TCPConn), manager.GetTextByJson(send))
		case <-self.closeWrite:
			//fmt.Println("写入server进程关闭")
			log.Println("close write")
			break
		}

	}
}

func (self server) newHandler() {
	for {
		var send string
		select {
		case send = <-self.recv:
			self.send <- manager.GetTextByJson(send)
		case <-self.closeHandler:
			//fmt.Println("写入server进程关闭")
			log.Println("close handler")
			break
		}

	}
}

func (self server) doReconnect() {
	self.reconnect <- true
	self.closeWrite <- true
	self.closeHandler <- true
}

func sendToServer(conn *net.TCPConn, msg string) {
	//	log.Println("开始写入数据")
	_, errW := conn.Write([]byte(msg + INTERVAL_END))
	if errW != nil {
		log.Fatalln("往服务端发送数据失败", errW)
	}
	//	time.Sleep(time.Second)
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
