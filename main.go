package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type DBConfig struct {
	BackupTargetPath string   `json:"backup_target_path"`
	EnableLogging    bool     `json:"enable_logging"`
	DBLists          []DBInfo `json:"dblists"`
}

type DBInfo struct {
	DBNumber   string `json:"db_number"`
	DBName     string `json:"db_name"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBHost     string `json:"db_host"`
	Remark     string `json:"remark"`
}

func main() {
	// Define command-line arguments
	action := flag.String("action", "", "Specify 'backup' to start backup or 'list' to list backed-up files")
	concurrency := flag.String("concurrency", "3", "Specify the number of concurrent backups (default is 3)")
	flag.Parse()

	if *action == "" {
		log.Fatalf("Please specify an action: 'backup' or 'list'")
	}

	// Parse concurrency value
	maxConcurrency, err := strconv.Atoi(*concurrency)
	if err != nil {
		log.Fatalf("Invalid concurrency value: %v", err)
	}

	switch *action {
	case "backup":
		startBackup(maxConcurrency)
	case "list":
		listBackups()
	default:
		log.Fatalf("Invalid action: %s", *action)
	}
}

func setupLogging(enableLogging bool) {
	if enableLogging {
		logFileName := time.Now().Format("20060102") + ".log"
		logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		log.SetOutput(logFile)
	} else {
		log.SetOutput(os.Stdout)
	}
}

func startBackup(maxConcurrency int) {
	// Read and parse JSON configuration file
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer file.Close()

	byteValue, _ := ioutil.ReadAll(file)

	var dbConfig DBConfig
	if err := json.Unmarshal(byteValue, &dbConfig); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	// Setup logging based on config
	setupLogging(dbConfig.EnableLogging)

	// Ensure the backup target path exists
	if err := os.MkdirAll(dbConfig.BackupTargetPath, os.ModePerm); err != nil {
		log.Fatalf("Failed to create backup target path: %v", err)
	}

	// Concurrently backup databases
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrency) // Control concurrency

	for _, db := range dbConfig.DBLists {
		wg.Add(1)
		go func(db DBInfo) {
			defer wg.Done()
			sem <- struct{}{} // Acquire a semaphore
			backupDatabase(db, dbConfig.BackupTargetPath)
			<-sem // Release the semaphore
		}(db)
	}

	wg.Wait()
	fmt.Println("All database backups completed")
}

func backupDatabase(db DBInfo, backupPath string) {
	// Get the current date and time for the filename
	timestamp := time.Now().Format("20060102_150405")
	sqlFilename := fmt.Sprintf("%s_%s_%s.sql", db.DBNumber, db.DBName, timestamp)
	gzFilename := filepath.Join(backupPath, fmt.Sprintf("%s.gz", sqlFilename))

	cmd := exec.Command("mysqldump", "-h", db.DBHost, "-u", db.DBUser, "-p"+db.DBPassword, db.DBName)
	gzipCmd := exec.Command("gzip")

	// Create a pipeline: mysqldump | gzip > file
	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()

	cmd.Stdout = pipeWriter
	gzipCmd.Stdin = pipeReader

	outfile, err := os.Create(gzFilename)
	if err != nil {
		log.Printf("Failed to create backup file %s: %v", gzFilename, err)
		return
	}
	defer outfile.Close()
	gzipCmd.Stdout = outfile

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start mysqldump for database %s: %v", db.DBName, err)
		return
	}

	if err := gzipCmd.Start(); err != nil {
		log.Printf("Failed to start gzip for database %s: %v", db.DBName, err)
		return
	}

	if err := cmd.Wait(); err != nil {
		log.Printf("Failed to complete mysqldump for database %s: %v", db.DBName, err)
		return
	}
	pipeWriter.Close() // Close the writer to signal gzip to finish

	if err := gzipCmd.Wait(); err != nil {
		log.Printf("Failed to complete gzip for database %s: %v", db.DBName, err)
	} else {
		log.Printf("Successfully backed up and compressed database %s to %s", db.DBName, gzFilename)
	}
}

func listBackups() {
	// Read and parse JSON configuration file
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer file.Close()

	byteValue, _ := ioutil.ReadAll(file)

	var dbConfig DBConfig
	if err := json.Unmarshal(byteValue, &dbConfig); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	// Setup logging based on config
	setupLogging(dbConfig.EnableLogging)

	files, err := ioutil.ReadDir(dbConfig.BackupTargetPath)
	if err != nil {
		log.Fatalf("Failed to read backup target path: %v", err)
	}

	if len(files) == 0 {
		fmt.Println("No backup files found")
		return
	}

	fmt.Println("Backed-up files:")
	for _, file := range files {
		fmt.Println(file.Name())
	}
}
