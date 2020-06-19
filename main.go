package main

func main(){
	go watch()
	server := &Server{}
	server.Run()
}
