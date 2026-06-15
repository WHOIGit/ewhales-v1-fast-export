// Package main includes memory statistics tracking.
package main

import (
	"encoding/csv"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// StartMemoryStatsRecording starts a background goroutine to record memory statistics
// to the specified CSV file at regular intervals. It captures both the Go runtime's
// internal heap allocation (HeapAlloc) and the total physical Resident Set Size (RSS)
// of the running process and its children.
func StartMemoryStatsRecording(interval int, filePath string) {
	go func() {
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			log.Printf("Failed to open memstats file: %v", err)
			return
		}
		defer f.Close()

		writer := csv.NewWriter(f)
		writer.Write([]string{"Timestamp", "HeapAlloc", "RSS"})
		writer.Flush()

		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()

		for t := range ticker.C {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			var totalRSS uint64
			pid := int32(os.Getpid())
			p, err := process.NewProcess(pid)
			if err == nil {
				mem, err := p.MemoryInfo()
				if err == nil {
					totalRSS = mem.RSS
				}
				children, err := p.Children()
				if err == nil {
					for _, child := range children {
						childMem, err := child.MemoryInfo()
						if err == nil {
							totalRSS += childMem.RSS
						}
					}
				}
			}

			writer.Write([]string{
				t.Format(time.RFC3339),
				strconv.FormatUint(m.Alloc, 10),
				strconv.FormatUint(totalRSS, 10),
			})
			writer.Flush()
		}
	}()
}
