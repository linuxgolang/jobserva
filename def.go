package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"gopkg.in/snksoft/crc.v1"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	HEADER = "250"//数据包开始标志
	PAYLOAD_SIZE = 1//数据长度占用字节,这里其实只保存20这个值,所以1个字节就够了
)

var (
	ErrLoginData = errors.New("错误的登陆数据")
)

var (
	chServerbChannel = make(chan []byte)//来自serverb的数据
	timeoutBroken = make(map[string]int)//客户端连接超时断开标记
	broken = make(map[string]int)//客户端连接断开标记
)

/**
 * 读取客户端的数据,并做超时标记
 */
func readSomething(conn *net.Conn, clientId string)  {
	buffer := make([]byte, 512)
	for{
		err := (*conn).SetReadDeadline(time.Now().Add(40*time.Second))
		if isErrAPrint(err){
			broken[clientId] = 1 //连接错误断开了
			return
		}
		_,err = (*conn).Read(buffer)
		if isTimeout(err) {
			timeoutBroken[clientId] = 1 //读取超时了
			return
		}
		if isErrAPrint(err){
			broken[clientId] = 1 //连接错误断开了
			return
		}
	}
}

/**
 * 为客户端提供服务的函数
 */
func writeSomething(conn *net.Conn,clientId string) bool {
	time.Sleep(1*time.Second)//此处模拟需要大量计算的场景
	if _,ok := timeoutBroken[clientId];ok {
		tick := time.NewTicker(time.Minute)
		//读取超时了,暂停1分钟,重连就继续,没重连就退出了
		for{
			select {
			case <-tick.C:
				tick.Stop()
				return false
			default:
				if _,ok := timeoutBroken[clientId];ok {
					fmt.Println("继续执行")
					break
				}
			}
		}
	}

	_, err := (*conn).Write([]byte("这是服务a"))
	if isErrAPrint(err) {
		//连接出现问题,退出.
		broken[clientId] = 1
		return false
	}
	return true
}

func checkLoginData(data []byte)(bool,error){
	total := len(data)//总长(26)
	headerLen := len(HEADER)//固定头部占用字节数(3)
	headerPayloadLen := headerLen+PAYLOAD_SIZE//固定头部占用字节数和payload长度数字占用字节(4)
	if len(data) <= headerPayloadLen {
		return false,nil
	}
	payloadAndCrc := data[headerPayloadLen:]//payload和crc所有数据(22)
	payloadLen := data[headerLen:headerPayloadLen][0]//payload数据应该占用的字节数(20)
	if string(data[:headerLen]) != HEADER {
		return false,ErrLoginData
	}
	if (uint64(len(payloadAndCrc)) - uint64(payloadLen)) != 2{
		return false,nil
	}
	payload := data[headerPayloadLen:total-2]
	rck := data[total-2:]

	ck := make([]byte, 2)
	hash := crc.NewHash(crc.X25)
	x25Crc := hash.CalculateCRC(payload)
	binary.LittleEndian.PutUint16(ck, uint16(x25Crc))
	if ck[0] == rck[0] && ck[1] == rck[1] {
		return true,nil
	}

	return false,ErrLoginData
}

func cReadSomething(conn *net.Conn,more bool) []byte {
	buffer := make([]byte, 512)
	for{
		_, err := (*conn).Read(buffer)
		if err != nil{
			//连接出现问题或服务器关闭连接,退出.
			fmt.Printf("Read error: %s", err)
		}
		chServerbChannel <- buffer
		//这里处理服务器返回的服务数据
		fmt.Println(buffer)

		if !more {return nil}
	}
}

func cWriteSomething(conn *net.Conn, data []byte) {
	_, err := (*conn).Write(data)
	if isErrAPrint(err) {
		//连接出现问题,退出.
		panic(fmt.Sprintf("Write error: %s", err))
	}
}

func isTimeout(err error) bool {
	if e, ok := err.(net.Error);ok && e.Timeout() {
		return true
	}else {
		return false
	}
}

func isErrAPrint(err error) bool {
	if err != nil{
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		return true
	}
	return false
}

func watch() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs,syscall.SIGINT)
	select {
	case sig := <- sigs:
		if sig == syscall.SIGINT{
			os.Exit(0)
		}
	}
}
