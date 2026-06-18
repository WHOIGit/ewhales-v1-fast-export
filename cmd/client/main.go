package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/agl/ewhales-v1-exporter/internal/client"
	"github.com/agl/ewhales-v1-exporter/internal/config"
	"github.com/agl/ewhales-v1-exporter/internal/db"
	"github.com/agl/ewhales-v1-exporter/internal/metrics"
	_ "github.com/go-sql-driver/mysql"
	"github.com/schollz/progressbar/v3"
)

func main() {
	configPath := flag.String("config", "client_config.json", "Path to the configuration file")
	progressFlag := flag.Bool("progress", false, "Enable progress bars for querying and streaming")
	helpFlag := flag.Bool("h", false, "Print help info and exit")
	helpFlagLong := flag.Bool("help", false, "Print help info and exit")

	memStatsFlag := flag.Bool("memstats", false, "Enable recording of memory statistics")
	memStatsIntervalFlag := flag.Int("memstats-interval", 1, "Interval in seconds to record memory statistics")
	memStatsFileFlag := flag.String("memstats-file", "memstats_client.csv", "File to write memory statistics to")

	flag.Usage = func() {
		fmt.Println("eWHALES v1 Client Exporter Tool")
		fmt.Println("Usage: client [options]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *helpFlag || *helpFlagLong {
		flag.Usage()
		return
	}

	if *memStatsFlag {
		metrics.StartMemoryStatsRecording(*memStatsIntervalFlag, *memStatsFileFlag)
	}

	// 1. Configuration Phase
	fmt.Println("Step 1: Configuration Phase")
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if cfg.Port == 0 {
		cfg.Port = 3306
	}

	fmt.Printf("  - Database Host : %s\n", cfg.Host)
	fmt.Printf("  - Database Port : %d\n", cfg.Port)
	fmt.Printf("  - Database Name : %s\n", cfg.Database)
	fmt.Printf("  - Server Address: %s\n", cfg.ServerAddress)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	// 2. Database Connection Phase
	fmt.Println("\nStep 2: Database Connection Phase")
	dbConn, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer dbConn.Close()

	if err := dbConn.Ping(); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	fmt.Println("Successfully connected to MySQL database.")

	// 3. Connect to Server
	fmt.Println("\nStep 3: Connecting to Server")
	apiClient, err := client.NewAPIClient(cfg)
	if err != nil {
		log.Fatalf("Error connecting to server: %v", err)
	}
	defer apiClient.Close()
	fmt.Println("Successfully connected to gRPC server.")

	// 4. Querying and Streaming Phase
	fmt.Println("\nStep 4: Querying and Streaming Phase")
	var queryProgressCallback func(int, int)
	if *progressFlag {
		var queryBar *progressbar.ProgressBar
		queryProgressCallback = func(processed int, total int) {
			if queryBar == nil {
				queryBar = progressbar.Default(int64(total), "Querying database")
			}
			queryBar.Set(processed)
		}
	}
	
	// Right now we query everything into memory and then send it as one batch
	// In the future this should be fully streamed from the database directly
	pivotData, err := db.QueryPivotData(dbConn, cfg, queryProgressCallback)
	if err != nil {
		log.Fatalf("Error querying pivot data: %v", err)
	}

	fmt.Println("\nStep 5: Sending Data to Server")
	err = apiClient.StartStream(len(pivotData.Logbooks), len(pivotData.LogbookEntries))
	if err != nil {
		log.Fatalf("Error starting stream: %v", err)
	}

	err = apiClient.SendBatch(pivotData)
	if err != nil {
		log.Fatalf("Error sending batch: %v", err)
	}

	err = apiClient.FinishStream()
	if err != nil {
		log.Fatalf("Error finishing stream: %v", err)
	}

	fmt.Println("\nSuccessfully exported data to server.")
}
