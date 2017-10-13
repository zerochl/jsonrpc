package main

import (
	"fmt"
	"io"
	"jsonrpc/manager"
	"log"
	"net"
	"os"
	"strings"
	//	"time"
)

const (
	SEND_MSG_CODE_OK = 1
	INTERVAL_END     = "###$$$%%%$$$###"
	//CMD
	START_CONNECT = "START_CONNECT"
	HEART_BEAT    = "HEART_BEAT"
)

var quitSemaphore chan bool
var closeCountDown int

//var writeStr chan int

func main() {
	//	response, e := http.Get("http://www.baidu.com")
	//	if e != nil {
	//		log.Println("http get error:", e)
	//	}
	//	defer response.Body.Close()
	//	var by []byte
	//	by, _ = ioutil.ReadAll(response.Body)
	//	log.Println("result:", manager.GetTextByJson("{\"url\":\"http://www.baidu.com\"}"))
	connect()
}

func connect() {
	closeCountDown = 10
	//	go monitorHeartBeat()
	var tcpAddr *net.TCPAddr
	tcpAddr, _ = net.ResolveTCPAddr("tcp", "104.224.174.229:8082")
	//	tcpAddr, _ = net.ResolveTCPAddr("tcp", "192.168.0.253:8085")

	conn, _ := net.DialTCP("tcp", nil, tcpAddr)
	defer conn.Close()
	fmt.Println("connected!")
	//告诉服务器，开始建立连接了
	sendToServer(conn, START_CONNECT)
	onMessageRecived(conn)
}

func onMessageRecived(conn *net.TCPConn) {
	//	reader := bufio.NewReader(conn)
	for {
		log.Println("before read")
		//		msg, err := reader.ReadString('\n')
		//		data, _, err := reader.ReadLine()
		readData := readFromServer(conn)
		//		_, err := conn.Read(readData)
		//		fmt.Println(msg)
		log.Println("end read", string(readData))
		//		if err != nil {
		//			quitSemaphore <- true
		//			break
		//		}
		//		time.Sleep(time.Second)
		log.Println("before write")
		if strings.Compare(HEART_BEAT, string(readData)) == 0 {
			sendToServer(conn, HEART_BEAT)
			closeCountDown = 10
			continue
		}

		readDataList := strings.Split(string(readData), INTERVAL_END)
		for _, readDataItem := range readDataList {
			log.Println("开始写入数据")
			if strings.Compare(HEART_BEAT, string(readData)) == 0 {
				sendToServer(conn, HEART_BEAT)
				closeCountDown = 10
				continue
			}
			sendToServer(conn, manager.GetTextByJson(readDataItem))
		}
		//		sendToServer(conn, manager.GetTextByJson(string(readData)))
		//		go sendToServer2(conn, string(readData))
	}
}

func readFromServer(conn *net.TCPConn) []byte {
	data := make([]byte, 0) //此处做一个输入缓冲以免数据过长读取到不完整的数据
	buf := make([]byte, 128)
	for {
		//		log.Println("开始分段读")
		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			checkErr(err)
		}
		data = append(data, buf[:n]...)
		//		log.Println("data:" + string(data))
		if strings.HasSuffix(string(data), INTERVAL_END) {
			break
		}
	}
	return data[0 : len(data)-len(INTERVAL_END)]
}

func sendToServer2(conn *net.TCPConn, readData string) {
	sendToServer(conn, manager.GetTextByJson(readData))
}

func sendToServer(conn *net.TCPConn, msg string) {
	//	log.Println("开始写入数据")
	_, errW := conn.Write([]byte(msg + INTERVAL_END))
	if errW != nil {
		log.Fatalln("往服务端发送数据失败", errW)
	}
	//	time.Sleep(time.Second)
}

//func monitorHeartBeat() {
//	for {
//		time.Sleep(time.Second)
//		closeCountDown--
//		if closeCountDown <= 0 {
//			//break to do reconnect
//			break
//		}
//	}
//	connect()
//}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
