package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/agl/ewhales-v1-exporter/internal/config"
	"github.com/agl/ewhales-v1-exporter/internal/db"
	"github.com/agl/ewhales-v1-exporter/internal/metrics"
	"github.com/agl/ewhales-v1-exporter/internal/serialize"
	_ "github.com/go-sql-driver/mysql"
	"github.com/schollz/progressbar/v3"
	"github.com/tebeka/strftime"
	"time"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to the configuration file")
	progressFlag := flag.Bool("progress", false, "Enable progress bars for querying and serialization")
	helpFlag := flag.Bool("h", false, "Print help info and exit")
	helpFlagLong := flag.Bool("help", false, "Print help info and exit")

	memStatsFlag := flag.Bool("memstats", false, "Enable recording of memory statistics")
	memStatsIntervalFlag := flag.Int("memstats-interval", 1, "Interval in seconds to record memory statistics")
	memStatsFileFlag := flag.String("memstats-file", "memstats.csv", "File to write memory statistics to")

	flag.Usage = func() {
		fmt.Println("eWHALES v1 Exporter Tool")
		fmt.Println("Usage: exporter [options]")
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

	baseName, err := strftime.Format(cfg.CSVBaseName, time.Now())
	if err != nil {
		log.Printf("Error formatting filename with strftime: %v. Using original config string.", err)
		baseName = cfg.CSVBaseName
	}

	fmt.Printf("  - Database Host : %s\n", cfg.Host)
	fmt.Printf("  - Database Port : %d\n", cfg.Port)
	fmt.Printf("  - Database Name : %s\n", cfg.Database)
	fmt.Printf("  - Database User : %s\n", cfg.Username)
	fmt.Printf("  - Target Table  : %s\n", cfg.Table)
	fmt.Printf("  - Logbooks CSV  : logbooks_%s\n", baseName)
	fmt.Printf("  - Entries CSV   : %s\n", baseName)

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

	// 3. Querying Phase
	fmt.Println("\nStep 3: Querying Phase")
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
	pivotData, err := db.QueryPivotData(dbConn, cfg, queryProgressCallback)
	if err != nil {
		log.Fatalf("Error querying pivot data: %v", err)
	}

	// 4. Serialization Phase
	fmt.Println("\nStep 4: Serialization Phase")
	serializer := &serialize.CSVSerializer{
		LogbooksFile:       "logbooks_" + baseName,
		LogbookEntriesFile: baseName,
		IdsToFields:        cfg.IdsToFields,
	}

	var serializeProgressCallback func(int, int)
	if *progressFlag {
		var serializeBar *progressbar.ProgressBar
		serializeProgressCallback = func(processed int, total int) {
			if serializeBar == nil {
				serializeBar = progressbar.Default(int64(total), "Writing to CSV   ")
			}
			serializeBar.Set(processed)
		}
	}
	err = serializer.Serialize(*pivotData, serializeProgressCallback)
	if err != nil {
		log.Fatalf("Error during serialization: %v", err)
	}

	fmt.Printf("\nSuccessfully exported data to %s and %s\n", serializer.LogbooksFile, serializer.LogbookEntriesFile)
}
