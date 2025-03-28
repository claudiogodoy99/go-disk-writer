package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func openFile(path string, currentFile int) (*os.File, error) {
	fileName := fmt.Sprintf("%03d", currentFile)
	fullName := filepath.Join(path, fileName)
	file, err := os.OpenFile(fullName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0744)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func writeLoop(path string, maxFileSize int, count int) error {
	buffer := make([]byte, 10*1024)
	for i := 0; i < len(buffer); i++ {
		buffer[i] = byte(i)
	}

	currentFile := 0

	for {
		file, err := openFile(path, currentFile)
		if err != nil {
			return err
		}

		var fileSize uint64 = 0

		for {
			written, err := file.Write(buffer)
			if err != nil {
				panic(err)
			}
			fileSize += uint64(written)
			if fileSize >= uint64(maxFileSize) {
				currentFile++
				if currentFile >= (count) {
					return nil
				}
				break
			}
		}
	}
}

func main() {
	go serveMetrics()

	var pathFlag = flag.String("p", "", "Directory to write files to")
	var sizeFlag = flag.Int("s", 10*1024*1024*1024, "Size of each file (default 10GB)")
	var countFlag = flag.Int("n", 3, "Number of files to write (default 3)")
	flag.Parse()

	if *pathFlag == "" {
		os.Stderr.WriteString("Path (p) argument is required\n")
		os.Exit(1)
	}
	if *sizeFlag <= 0 {
		os.Stderr.WriteString("Size (s) argument should be greater than zero\n")
		os.Exit(1)
	}
	if *countFlag <= 0 {
		os.Stderr.WriteString("File count (n) argument should be greater than zero\n")
		os.Exit(1)
	}

	err := writeLoop(*pathFlag, *sizeFlag, *countFlag)
	if err != nil {
		panic(err)
	}
	time.Sleep(1 * time.Hour)
}

func serveMetrics() {
	fmt.Printf("serving metrics at localhost:2223/metrics")
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":9090", nil) //nolint:gosec // Ignoring G114: Use of net/http serve function that has no support for setting timeouts.
	if err != nil {
		fmt.Printf("error serving http: %v", err)
		return
	}
}
