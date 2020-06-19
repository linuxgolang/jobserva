package main

func main(){
	go watch()
	server := &Server{
		Client: &Client{
			Sip:   "127.0.0.1",
			Sport: 8882,
		},
		Ip:     "127.0.0.1",
		Port:   8881,
	}
	server.Run()
}
