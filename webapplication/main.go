package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"yohan/databaseutilities/coreactions"
	"yohan/databaseutilities/logger"
	"yohan/databaseutilities/webapplication"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

var ApplicationType string
var DatabaseType string
var DatabaseUsername string
var DatabasePassword string
var DatabaseHost string
var DatabasePort int
var DatabaseName string
var ListOfTables []string
var ActionType string    // Restore\Backup database action
var DateToRestore string // Restore databse till date
var DatabaseRestoreInputFile string
var DatabaseRestoreOutputFile string
var BackupSchedule string

func init() {

	rootCmd.PersistentFlags().StringVarP(&ApplicationType, "applicationtype", "a", ApplicationType, "To Define the type of the application")
	rootCmd.PersistentFlags().StringVarP(&DatabaseType, "dbtype", "d", DatabaseType, "To Define the type of the database")
	rootCmd.PersistentFlags().StringVarP(&DatabaseUsername, "username", "u", DatabaseUsername, "To Define the Username of the database")
	rootCmd.PersistentFlags().StringVarP(&DatabasePassword, "password", "p", DatabasePassword, "To Define the Password of the database")
	rootCmd.PersistentFlags().StringVarP(&DatabaseHost, "host", "H", DatabaseHost, "To Define the Host of the database") // Changed from 'h' to 'H'
	rootCmd.PersistentFlags().IntVarP(&DatabasePort, "port", "o", DatabasePort, "To Define the Port of the database")
	rootCmd.PersistentFlags().StringVarP(&DatabaseName, "dbname", "n", DatabaseName, "To Define the Name of the database")
	rootCmd.PersistentFlags().StringSliceVarP(&ListOfTables, "tables", "t", ListOfTables, "To Define the list of tables to be included in the backup")
	rootCmd.PersistentFlags().StringVarP(&ActionType, "actiontype", "e", ActionType, "To Define the action type (restore or backup)")
	rootCmd.PersistentFlags().StringVarP(&DateToRestore, "date", "r", DateToRestore, "To Define the date to restore the database")
	rootCmd.PersistentFlags().StringVarP(&DatabaseRestoreInputFile, "inputfile", "i", DatabaseRestoreInputFile, "To Define the input file for restore database")
	rootCmd.PersistentFlags().StringVarP(&DatabaseRestoreOutputFile, "outputfile", "y", DatabaseRestoreOutputFile, "To Define the output file for restore database")
	rootCmd.PersistentFlags().StringVarP(&BackupSchedule, "schedule", "s", "", "Cron schedule for automatic backups (e.g., '0 0 * * *')")

	rootCmd.PersistentFlags().BoolP("help", "h", false, "Help for this command")

	rootCmd.MarkFlagRequired("applicationtype")
}

func scheduleBackup(cronExpr string) {
	c := cron.New()
	_, err := c.AddFunc(cronExpr, func() {
		// timestamp := time.Now().Format("20060102_150405")
		// outputFile := "backup_" + timestamp + ".zip"
		logger.Info("Starting scheduled backup...")

		// Triggering the database backup
		coreactions.BackupDatabase(DatabaseType, DatabaseHost, DatabasePort, DatabaseUsername, DatabasePassword, DatabaseName, DatabaseRestoreInputFile)

		logger.Info("Scheduled backup completed and stored at: " + DatabaseRestoreInputFile)

		fmt.Println("Backup job completed. Exiting the application.")
		os.Exit(0)
	})
	if err != nil {
		log.Fatalf("Failed to schedule the backup: %v", err)
	}
	c.Start()
	fmt.Println("Backup job scheduled. Waiting for the next run...")

	// Allow cron jobs to execute by keeping the app running until job is done
	select {}
}

var rootCmd = &cobra.Command{
	Use:   filepath.Base(os.Args[0]),
	Short: "Backup and Restore the database",
	Long:  "Backup and Restore the database",
	Run: func(cmd *cobra.Command, args []string) {
		switch ApplicationType {
		case "application":
			logger.Info("Started As Web Application")
			err := godotenv.Load(".env") // Load .env file
			if err != nil {
				log.Fatalf("Error loading .env file")
			}
			connStr := os.Getenv("DATABASE_URL")
			db, err1 := sql.Open("postgres", connStr)
			if err1 != nil {
				log.Fatalf("Failed to connect to the database: %v", err)
			}
			webapplication.DB = db
			// Start the web application
			webapplication.RunWebApp()
			defer db.Close()
		case "commandline":
			if BackupSchedule != "" && ApplicationType == "commandline" {
				logger.Info("Scheduling automatic backups...")
				scheduleBackup(BackupSchedule)
				select {} // Keep the application running to allow cron jobs to execute
			} else if ActionType == "backup" && len(ListOfTables) == 0 {
				coreactions.BackupDatabase(DatabaseType, DatabaseHost, DatabasePort, DatabaseUsername, DatabasePassword, DatabaseName, DatabaseRestoreOutputFile)
			} else if ActionType == "restore" && len(ListOfTables) == 0 {
				coreactions.RestoreDatabase(DatabaseType, DatabaseHost, DatabasePort, DatabaseUsername, DatabasePassword, DatabaseName, DatabaseRestoreInputFile)
			} else if ActionType == "backup" && len(ListOfTables) > 0 {
				coreactions.BackupDatabaseTables(DatabaseType, DatabaseHost, DatabasePort, DatabaseUsername, DatabasePassword, DatabaseName, DatabaseRestoreOutputFile, ListOfTables)
			} else if ActionType == "restore" && len(ListOfTables) > 0 {
				coreactions.RestoreDatabaseTables(DatabaseType, DatabaseHost, DatabasePort, DatabaseUsername, DatabasePassword, DatabaseName, DatabaseRestoreInputFile, ListOfTables)
			} else if ActionType == "pittest" {
				// Point in time restore
				coreactions.RestoreDatabaseOfSpecificDate(DatabaseType, DatabaseHost, DatabasePort, DatabaseUsername, DatabasePassword, DatabaseName, DatabaseRestoreInputFile, DateToRestore)
			}
			logger.Info("Database operation completed successfully")
		}
	},
}

func main() {
	logger.Init()
	logger.Info("Database Utility Starts")
	if err := rootCmd.Execute(); err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}
