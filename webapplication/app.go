package webapplication

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"
	"yohan/databaseutilities/coreactions"
	"yohan/databaseutilities/logger"

	"github.com/gorilla/mux"
)

var DB *sql.DB

type BackupRecord struct {
	ID       int
	Action   string
	FilePath string
	Date     time.Time
	Tables   string
	Status   string
}

func RunWebApp() {
	r := mux.NewRouter()
	r.HandleFunc("/", homePage)
	r.HandleFunc("/backup", backupHandler).Methods("POST")
	r.HandleFunc("/restore", restoreHandler).Methods("POST")
	r.HandleFunc("/logs", viewLogsHandler).Methods("GET")

	fs := http.FileServer(http.Dir("./static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	log.Println("Starting Web Application on http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/index.html")
	tmpl.Execute(w, nil)
}

func getStatus(err error) string {
	if err != nil {
		logger.Error("Operation failed:", err)
		return "Failed"
	}
	logger.Info("Operation succeeded")
	return "Success"
}

// func backupHandler(w http.ResponseWriter, r *http.Request) {
// 	filePath := "backup.sql"
// 	err := coreactions.BackupDatabase("psql", "localhost", 5432, "postgres", "postgres", "mydb", filePath)
// 	status := "success"
// 	if err != nil {
// 		status = "failed"
// 	}
// 	logBackupRestore("backup", filePath, "", status)
// 	http.Redirect(w, r, "/", http.StatusSeeOther)
// }

// func restoreHandler(w http.ResponseWriter, r *http.Request) {
// 	filePath := "backup.sql"
// 	err := coreactions.RestoreDatabase("psql", "localhost", 5432, "postgres", "postgres", "mydb", filePath)
// 	status := "success"
// 	if err != nil {
// 		status = "failed"
// 	}
// 	logBackupRestore("restore", filePath, "", status)
// 	http.Redirect(w, r, "/", http.StatusSeeOther)
// }

func viewLogsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := DB.Query("SELECT id, action, file_path, date, tables, status FROM backup_restore_logs ORDER BY date DESC")
	if err != nil {
		http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []BackupRecord
	for rows.Next() {
		var logRecord BackupRecord
		err := rows.Scan(&logRecord.ID, &logRecord.Action, &logRecord.FilePath, &logRecord.Date, &logRecord.Tables, &logRecord.Status)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		logs = append(logs, logRecord)
	}

	tmpl, _ := template.ParseFiles("templates/logs.html")
	tmpl.Execute(w, logs)
}

func LogBackupRestore(action, filePath, tables, status string) {
	_, err := DB.Exec(
		"INSERT INTO backup_restore_logs (action, file_path, tables, status) VALUES ($1, $2, $3, $4)",
		action, filePath, tables, status,
	)
	if err != nil {
		log.Printf("Failed to log %s action: %v", action, err)
	}
}
func parseTableList(tableList string) []string {
	if tableList == "" {
		return []string{} // Return an empty slice if no tables are provided
	}
	return strings.Split(tableList, ",") // Split the input string by commas and return the slice
}

func backupHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// Extract input values from form
	dbHost := r.FormValue("dbHost")
	dbPort, _ := strconv.Atoi(r.FormValue("dbPort"))
	dbUsername := r.FormValue("dbUsername")
	dbPassword := r.FormValue("dbPassword")
	backupFile := r.FormValue("backupFile")
	tables := r.FormValue("tables")
	databasename := r.FormValue("databasename")

	var status string
	var err error

	// Backup all tables or selected tables
	if tables != "" {
		err = coreactions.BackupDatabaseTables("postgres", dbHost, dbPort, dbUsername, dbPassword, databasename, backupFile, parseTableList(tables))
	} else {
		err = coreactions.BackupDatabase("postgres", dbHost, dbPort, dbUsername, dbPassword, databasename, backupFile)
	}

	status = getStatus(err)
	LogBackupRestore("backup", backupFile, tables, status)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func restoreHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// Extract input values from form
	dbHost := r.FormValue("dbHost")
	dbPort, _ := strconv.Atoi(r.FormValue("dbPort"))
	dbUsername := r.FormValue("dbUsername")
	dbPassword := r.FormValue("dbPassword")
	restoreFile := r.FormValue("restoreFile")
	restoreDate := r.FormValue("restoreDate")
	databasename := r.FormValue("databasename")

	var status string
	var err error

	// Perform point-in-time restore if date is provided
	if restoreDate != "" {
		err = coreactions.RestoreDatabaseOfSpecificDate("postgres", dbHost, dbPort, dbUsername, dbPassword, databasename, restoreFile, restoreDate)
	} else {
		err = coreactions.RestoreDatabase("postgres", dbHost, dbPort, dbUsername, dbPassword, databasename, restoreFile)
	}

	status = getStatus(err)
	LogBackupRestore("restore", restoreFile, "", status)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
