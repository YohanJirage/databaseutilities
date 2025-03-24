// logger.go - Advanced logger with handling for multiple arguments, including maps and mixed types
package logger

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

const (
	logFilePath   = "databaseutilities.log"
	maxLogSize    = 10 * 1024 * 1024 // 10 MB
	compressedExt = ".gz"
)

var (
	debugLogger   *log.Logger
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
)

func Init() {
	rotateIfNeeded()

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	debugLogger = log.New(file, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	warningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func rotateIfNeeded() {
	fileInfo, err := os.Stat(logFilePath)
	if err == nil && fileInfo.Size() >= maxLogSize {
		compressLogFile()
	}
}

func compressLogFile() {
	timestamp := time.Now().Format("20060102_150405")
	compressedFileName := fmt.Sprintf("%s_%s%s", logFilePath, timestamp, compressedExt)

	inputFile, err := os.Open(logFilePath)
	if err != nil {
		log.Printf("Failed to open log file for compression: %v", err)
		return
	}
	defer inputFile.Close()

	outputFile, err := os.Create(compressedFileName)
	if err != nil {
		log.Printf("Failed to create compressed log file: %v", err)
		return
	}
	defer outputFile.Close()

	gzipWriter := gzip.NewWriter(outputFile)
	defer gzipWriter.Close()

	if _, err := gzipWriter.Write([]byte(readFileContent(inputFile))); err != nil {
		log.Printf("Failed to write compressed log data: %v", err)
		return
	}
	log.Printf("Compressed log file created: %s", compressedFileName)

	if err := os.Truncate(logFilePath, 0); err != nil {
		log.Printf("Failed to truncate log file: %v", err)
	}
}

func readFileContent(file *os.File) string {
	file.Seek(0, 0)
	content, _ := os.ReadFile(file.Name())
	return string(content)
}

// Handles concatenating and formatting multiple arguments
func formatArgs(args ...interface{}) string {
	var sb strings.Builder
	for _, arg := range args {
		if arg == nil {
			sb.WriteString("<nil> ")
		} else {
			switch v := arg.(type) {
			case string:
				if v == "" {
					sb.WriteString("<empty_string> ")
				} else {
					sb.WriteString(v + " ")
				}
			case map[string]interface{}:
				jsonData, _ := json.Marshal(v)
				sb.WriteString(string(jsonData) + " ")
			case []map[string]interface{}:
				jsonData, _ := json.Marshal(v) // Convert map to JSON string
				sb.WriteString(string(jsonData) + " ")
			default:
				sb.WriteString(fmt.Sprintf("%v ", v))
			}
		}
	}
	return sb.String()
}

// Enhanced logging functions that handle multiple types and arguments
func Debug(args ...interface{}) {
	debugLogger.Println(formatArgs(args...))
}

func Info(args ...interface{}) {
	infoLogger.Println(formatArgs(args...))
}

func Warning(args ...interface{}) {
	warningLogger.Println(formatArgs(args...))
}

func Error(args ...interface{}) {
	errorLogger.Println(formatArgs(args...))
}
