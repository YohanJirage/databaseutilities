package coreactions

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"yohan/databaseutilities/logger"
)

func BackupDatabase(dbType, host string, port int, username, password, dbName, outputFile string) error {
	logger.Info(fmt.Sprintf("Starting full backup of database %s", dbName))

	if outputFile == "" {
		timestamp := time.Now().Format("20060102_150405")
		outputFile = fmt.Sprintf("%s_backup_%s.sql", dbName, timestamp)
	}

	dir := filepath.Dir(outputFile)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Error(fmt.Sprintf("Failed to create directory %s: %v", dir, err))
			return err
		}
	}

	var cmd *exec.Cmd

	switch strings.ToLower(dbType) {
	case "mysql", "mariadb":
		cmd = exec.Command(
			"mysqldump",
			fmt.Sprintf("-h%s", host),
			fmt.Sprintf("-P%d", port),
			fmt.Sprintf("-u%s", username),
			fmt.Sprintf("-p%s", password),
			"--single-transaction",
			"--routines",
			"--triggers",
			"--databases",
			dbName,
		)

	case "postgresql", "postgres":
		os.Setenv("PGPASSWORD", password)
		cmd = exec.Command(
			"pg_dump",
			fmt.Sprintf("--host=%s", host),
			fmt.Sprintf("--port=%d", port),
			fmt.Sprintf("--username=%s", username),
			"--format=plain",
			"--create",
			"--clean",
			dbName,
		)

	default:
		err := fmt.Errorf("unsupported database type: %s", dbType)
		logger.Error(err.Error())
		return err
	}

	// Redirect output to file
	outFile, err := os.Create(outputFile)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create output file: %v", err))
		return err
	}
	defer outFile.Close()

	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logger.Error(fmt.Sprintf("Database backup failed: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("Database backup completed successfully to %s", outputFile))
	return nil
}

func BackupDatabaseTables(dbType, host string, port int, username, password, dbName, outputFile string, tables []string) error {
	logger.Info(fmt.Sprintf("Starting backup of selected tables in database %s", dbName))

	if outputFile == "" {
		timestamp := time.Now().Format("20060102_150405")
		outputFile = fmt.Sprintf("%s_tables_backup_%s.sql", dbName, timestamp)
	}

	dir := filepath.Dir(outputFile)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Error(fmt.Sprintf("Failed to create directory %s: %v", dir, err))
			return err
		}
	}

	var cmd *exec.Cmd

	switch strings.ToLower(dbType) {
	case "mysql", "mariadb":
		args := []string{
			fmt.Sprintf("-h%s", host),
			fmt.Sprintf("-P%d", port),
			fmt.Sprintf("-u%s", username),
			fmt.Sprintf("-p%s", password),
			"--single-transaction",
			dbName,
		}
		// Add tables to arguments
		args = append(args, tables...)
		cmd = exec.Command("mysqldump", args...)

	case "postgresql", "postgres":

		os.Setenv("PGPASSWORD", password)
		args := []string{
			fmt.Sprintf("--host=%s", host),
			fmt.Sprintf("--port=%d", port),
			fmt.Sprintf("--username=%s", username),
			"--format=plain",
			dbName,
		}

		for _, table := range tables {
			args = append(args, fmt.Sprintf("--table=%s", table))
		}
		cmd = exec.Command("pg_dump", args...)

	default:
		err := fmt.Errorf("unsupported database type: %s", dbType)
		logger.Error(err.Error())
		return err
	}

	outFile, err := os.Create(outputFile)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create output file: %v", err))
		return err
	}
	defer outFile.Close()

	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logger.Error(fmt.Sprintf("Database tables backup failed: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("Database tables backup completed successfully to %s", outputFile))
	return nil
}
