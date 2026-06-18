package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/agl/ewhales-v1-exporter/internal/config"
	"github.com/agl/ewhales-v1-exporter/internal/server"
)

func main() {
	configPath := flag.String("config", "server_config.json", "Path to the configuration file")
	helpFlag := flag.Bool("h", false, "Print help info and exit")
	helpFlagLong := flag.Bool("help", false, "Print help info and exit")

	flag.Usage = func() {
		fmt.Println("eWHALES v1 Server")
		fmt.Println("Usage: server [options]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *helpFlag || *helpFlagLong {
		flag.Usage()
		return
	}

	fmt.Println("Loading configuration...")
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if cfg.ListenPort == 0 {
		cfg.ListenPort = 8443
	}

	if err := server.RunServer(cfg); err != nil {
		log.Fatalf("Server exited with error: %v", err)
	}
}
