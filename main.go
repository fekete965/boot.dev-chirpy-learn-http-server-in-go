package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

const DEFAULT_PORT = 8080
const STATIC_DIR = "./static"

type ServerConfig struct {
	Port int
}

func getServerPort() (int, error) {
	portEnv := os.Getenv("CHIRPY_PORT")

	if portEnv == "" {
		return -1, fmt.Errorf("no port defined")
	}

	port, err := strconv.Atoi(portEnv)
	if err != nil {
		return -1, fmt.Errorf("invalid port \"%s\": %w", portEnv, err)
	}

	return port, nil
}

var handleHome = http.FileServer(http.Dir(STATIC_DIR))

func main() {
	port, err := getServerPort()
	if err != nil {
		fmt.Printf("error getting server port: %v\nfalling back to default port: %d\n", err, DEFAULT_PORT)
		port = DEFAULT_PORT
	}

	cfg := ServerConfig{
		Port: port,
	}

	mux := http.NewServeMux()
	mux.Handle("/", handleHome)

	addr := fmt.Sprintf(":%d", cfg.Port)
	server := http.Server{
		Addr: addr,
		Handler: mux,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("could not start the server: %v", err)
	}

	fmt.Printf("server start on port: %d\n", port)
}
