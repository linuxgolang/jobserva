package main

import (
	"net"
	"os"
	"strconv"
	"time"
)

type Client struct {
	Sip string
	Sport uint64
}

func (client *Client)Run(loginData []byte)  {
	conn, err := net.DialTimeout("tcp",client.Sip+":"+strconv.FormatUint(client.Sport, 10),3*time.Second)
	if isErrAPrint(err) { os.Exit(1) }
	defer conn.Close()
	clientId := string(loginData[len(HEADER)+PAYLOAD_SIZE:len(loginData)-2])

	if !client.login(&conn,loginData) {
		tick := time.NewTicker(5*time.Second)
		for{
			select {
			case <-tick.C:
				_,ok1 := timeoutBroken[clientId]
				_,ok2 := broken[clientId]
				if !ok1 && !ok2 {
					//客户端没有断开才登陆
					if client.login(&conn,loginData) {
						break
					}
				}else {
					return
				}
			}
		}
	}

	go func() {
		for{
			_,ok1 := timeoutBroken[clientId]
			_,ok2 := broken[clientId]
			if ok1 || ok2{
				//客户端已经断开,所以就不需要再连接serverb了
				conn.Close()
			}
		}
	}()

	client.todoSomething(&conn,clientId)
}

func (client *Client)login(conn *net.Conn, loginData []byte) bool {
	cWriteSomething(conn,loginData)
	isLogin := true
	defer func() {
		if err:=recover();err!=nil{
			isLogin = false
		}
	}()
	cReadSomething(conn, false)//如果这个函数没有panic,就说明登陆成功了,登陆失败服务器会关闭连接,报panic.
	return isLogin
}

func (client *Client)todoSomething(conn *net.Conn, clientId string)  {
	cReadSomething(conn,true)
	//for{
	//	cWriteSomething(conn,[]byte("abc"))
	//	time.Sleep(time.Second)
	//}
}