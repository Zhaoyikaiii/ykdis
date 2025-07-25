package main

func main() {
	// Initialize the server
	server, err := NewServer(DefaultAddr)
	if err != nil {
		panic(err)
	}

	server.RunAndServe()
}
