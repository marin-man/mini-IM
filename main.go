package main

func main() {
	server := NewServier("127.0.0.1", 8888)
	server.Start()
}
