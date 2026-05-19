package main

import (
	"log"
	"my-redis/tcp"
)

func main() {
	// 创建TCP服务器，使用EchoHandler测试
	server := tcp.EchoHandler{}
	log.Println("Starting Redis server...")
	if err := tcp.ListenAndServeWithSignal(&tcp.Config{Address: "localhost:6379"}, &server); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}