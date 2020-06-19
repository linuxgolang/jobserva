package main

import (
	"net"
	"os"
	"strconv"
)

type Server struct {
	*Client
	Ip string
	Port uint64
}

func (server *Server)Run() {
	listen, err := net.Listen("tcp", server.Ip+":"+strconv.FormatUint(server.Port, 10))
	if isErrAPrint(err) {os.Exit(1)}
	for {
		conn, err := listen.Accept()
		if isErrAPrint(err) {os.Exit(1)}
		server.handle(&conn)
	}
}

func (server *Server)handle(conn *net.Conn)  {
	defer (*conn).Close()
	isLogin,data := server.checkLogin(conn)
	if !isLogin {return}
	clientId := string(data[len(HEADER)+PAYLOAD_SIZE:len(data)-2])
	defer func() {
		delete(broken,clientId)
		delete(timeoutBroken,clientId)
	}()
	//转发服务器b的数据
	go server.forwardData(conn, data)
	//这是a服务器提供的服务
	server.todoSomething(conn,clientId)
}

/**
 * 接收到的登陆数据转给serverb
 * 同时转发serverb的数据给客户端
 */
func (server *Server)forwardData(conn *net.Conn, loginData []byte)  {
	//转发serverb的数据
	go func() {
		for{
			select {
			case byts := <-chServerbChannel:
				_,err := (*conn).Write(byts)
				if isErrAPrint(err){
					return
				}
			}
		}
	}()
	//接收到的登陆数据转给serverb
	//接收serverb的数据
	server.Client.Run(loginData)
}

func (server *Server)todoSomething(conn *net.Conn,clientId string)  {
	//写数据,并做超时暂停,连接继续
	go func() {
		for{
			if !writeSomething(conn,clientId){
				return
			}
		}
	}()

	//读取客户端的数据,并做超时标记
	readSomething(conn,clientId)
}

func (server *Server)checkLogin(conn *net.Conn) (bool,[]byte) {
	buffer := make([]byte,512)
	for{
		n, err := (*conn).Read(buffer)
		if isErrAPrint(err) {return false,nil}
		data := make([]byte,n)
		copy(data,buffer[:n+1])
		isRightComplete,err := checkLoginData(data)
		if isErrAPrint(err) {return false,nil}
		if isRightComplete {
			return true,data
		}
	}
}