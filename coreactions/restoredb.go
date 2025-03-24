package coreactions

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
	"yohan/databaseutilities/logger"
)

func RestoreDatabase(dbType, host string, port int, username, password, dbName, inputFile string) error {
	logger.Info(fmt.Sprintf("Starting restore of database %s from %s", dbName, inputFile))

	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		err := fmt.Errorf("input file does not exist: %s", inputFile)
		logger.Error(err.Error())
		return err
	}

	var cmd *exec.Cmd

	switch strings.ToLower(dbType) {
	case "mysql", "mariadb":
		cmd = exec.Command(
			"mysql",
			fmt.Sprintf("-h%s", host),
			fmt.Sprintf("-P%d", port),
			fmt.Sprintf("-u%s", username),
			fmt.Sprintf("-p%s", password),
			dbName,
		)

		inFile, err := os.Open(inputFile)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to open input file: %v", err))
			return err
		}
		defer inFile.Close()
		cmd.Stdin = inFile

	case "postgresql", "postgres":

		os.Setenv("PGPASSWORD", password)
		cmd = exec.Command(
			"psql",
			fmt.Sprintf("--host=%s", host),
			fmt.Sprintf("--port=%d", port),
			fmt.Sprintf("--username=%s", username),
			fmt.Sprintf("--dbname=%s", dbName),
			"-f", inputFile,
		)

	default:
		err := fmt.Errorf("unsupported database type: %s", dbType)
		logger.Error(err.Error())
		return err
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logger.Error(fmt.Sprintf("Database restore failed: %v", err))
		return err
	}

	logger.Info(fmt.Sprintf("Database restore completed successfully from %s", inputFile))
	return nil
}

func RestoreDatabaseTables(dbType, host string, port int, username, password, dbName, inputFile string, tables []string) error {
	logger.Info(fmt.Sprintf("Starting restore of selected tables to database %s from %s", dbName, inputFile))

	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		err := fmt.Errorf("input file does not exist: %s", inputFile)
		logger.Error(err.Error())
		return err
	}

	if strings.ToLower(dbType) == "mysql" || strings.ToLower(dbType) == "mariadb" {
		// Create a temporary file to hold the filtered backup
		tempFile, err := os.CreateTemp("", "filtered_backup_*.sql")
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to create temporary file: %v", err))
			return err
		}
		defer os.Remove(tempFile.Name())

		data, err := os.ReadFile(inputFile)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to read input file: %v", err))
			return err
		}

		content := string(data)
		var filteredContent strings.Builder

		headerEnd := strings.Index(content, "CREATE TABLE")
		if headerEnd > 0 {
			filteredContent.WriteString(content[:headerEnd])
		}

		for _, table := range tables {
			tableStart := strings.Index(content, fmt.Sprintf("CREATE TABLE `%s`", table))
			if tableStart == -1 {
				logger.Warning(fmt.Sprintf("Table %s not found in backup file", table))
				continue
			}

			nextTableStart := strings.Index(content[tableStart+1:], "CREATE TABLE")
			if nextTableStart == -1 {

				filteredContent.WriteString(content[tableStart:])
			} else {
				filteredContent.WriteString(content[tableStart : tableStart+nextTableStart+1])
			}
		}

		if err := os.WriteFile(tempFile.Name(), []byte(filteredContent.String()), 0644); err != nil {
			logger.Error(fmt.Sprintf("Failed to write to temporary file: %v", err))
			return err
		}

		inputFile = tempFile.Name()
	}

	return RestoreDatabase(dbType, host, port, username, password, dbName, inputFile)
}

func RestoreDatabaseOfSpecificDate(dbType, host string, port int, username, password, dbName, inputFile, date string) error {
	logger.Info(fmt.Sprintf("Starting point-in-time restore of database %s to date %s", dbName, date))

	// Check if input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		err := fmt.Errorf("input file does not exist: %s", inputFile)
		logger.Error(err.Error())
		return err
	}

	// Parse the date string
	targetDate, err := time.Parse("2006-01-02T15:04:05", date)
	if err != nil {
		logger.Error(fmt.Sprintf("Invalid date format: %v", err))
		return err
	}

	switch strings.ToLower(dbType) {
	case "mysql", "mariadb":

		if err := RestoreDatabase(dbType, host, port, username, password, dbName, inputFile); err != nil {
			return err
		}

		// Execute mysqlbinlog to apply logs up to the target date
		cmd := exec.Command(
			"mysqlbinlog",
			fmt.Sprintf("--stop-datetime='%s'", targetDate.Format("2006-01-02 15:04:05")),
			"--database", dbName,
			"/var/log/mysql/mysql-bin.*", // Adjust path to binlogs as needed
		)

		// Pipe output to mysql command
		mysqlCmd := exec.Command(
			"mysql",
			fmt.Sprintf("-h%s", host),
			fmt.Sprintf("-P%d", port),
			fmt.Sprintf("-u%s", username),
			fmt.Sprintf("-p%s", password),
			dbName,
		)

		pipe, err := cmd.StdoutPipe()
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to create pipe: %v", err))
			return err
		}
		mysqlCmd.Stdin = pipe

		// Start both commands
		if err := cmd.Start(); err != nil {
			logger.Error(fmt.Sprintf("Failed to start mysqlbinlog: %v", err))
			return err
		}
		if err := mysqlCmd.Start(); err != nil {
			logger.Error(fmt.Sprintf("Failed to start mysql: %v", err))
			return err
		}

		// Wait for completion
		if err := cmd.Wait(); err != nil {
			logger.Error(fmt.Sprintf("mysqlbinlog failed: %v", err))
			return err
		}
		if err := mysqlCmd.Wait(); err != nil {
			logger.Error(fmt.Sprintf("mysql failed: %v", err))
			return err
		}

	case "postgresql", "postgres":
		os.Setenv("PGPASSWORD", password)
		cmd := exec.Command(
			"pg_restore",
			fmt.Sprintf("--host=%s", host),
			fmt.Sprintf("--port=%d", port),
			fmt.Sprintf("--username=%s", username),
			fmt.Sprintf("--dbname=%s", dbName),
			fmt.Sprintf("--recovery-target-time=%s", targetDate.Format("2006-01-02 15:04:05")),
			inputFile,
		)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			logger.Error(fmt.Sprintf("Point-in-time recovery failed: %v", err))
			return err
		}

	default:
		err := fmt.Errorf("unsupported database type for point-in-time recovery: %s", dbType)
		logger.Error(err.Error())
		return err
	}

	logger.Info(fmt.Sprintf("Point-in-time recovery completed successfully to %s", date))
	return nil
}
