# Database Utility Tool Documentation

## Overview
The Database Utility Tool is a command-line and web-based application designed to handle backup and restore operations for databases. It supports PostgreSQL, MySQL, and MariaDB, providing functionalities such as full backups, table-specific backups, scheduled backups, and point-in-time restores.

---

## Features
- **Backup and Restore** for entire databases or specific tables.
- **Command-line** and **Web Application** modes.
- **Cron-based scheduling** for automatic backups.
- **Supports PostgreSQL, MySQL, MariaDB**.
- **Point-in-time Restore (PITR)** for specific date recovery.

---

## Installation

### Clone the repository
```bash
git clone https://github.com/yohan/databaseutilities.git
cd databaseutilities
```

### Install dependencies
```bash
go mod tidy
```

### Build the application
```bash
go build -o dbutility
```

---

## Command-line Usage

### Flags
- `-a`, `--applicationtype`: **(Required)** `application` or `commandline`
- `-d`, `--dbtype`: Database type (`postgres`, `mysql`, `mariadb`)
- `-u`, `--username`: Database username
- `-p`, `--password`: Database password
- `-H`, `--host`: Database host
- `-o`, `--port`: Database port
- `-n`, `--dbname`: Database name
- `-t`, `--tables`: List of tables (optional)
- `-e`, `--actiontype`: `backup` or `restore`
- `-r`, `--date`: Date for point-in-time restore
- `-i`, `--inputfile`: Input file for restore
- `-y`, `--outputfile`: Output file for backup
- `-s`, `--schedule`: Cron expression for scheduled backups

---

## Examples

### Full Backup
```bash
dbutility -a commandline -d postgres -u user -p pass -H localhost -o 5432 -n mydb -e backup -y backup.sql
```

### Restore Backup
```bash
dbutility -a commandline -d postgres -u user -p pass -H localhost -o 5432 -n mydb -e restore -i backup.sql
```

### Backup Specific Tables
```bash
dbutility -a commandline -d mysql -u root -p pass -H localhost -o 3306 -n salesdb -e backup -t users,orders -y tables_backup.sql
```

### Scheduled Backup (Daily at Midnight)
```bash
dbutility -a commandline -d postgres -u user -p pass -H localhost -o 5432 -n mydb -e backup -s "0 0 * * *"
```

### Point-in-Time Restore
```bash
dbutility -a commandline -d postgres -u user -p pass -H localhost -o 5432 -n mydb -e pittest -r "2024-03-15 10:30:00"
```

---

## Web Application Mode

### Start the web application
```bash
dbutility -a application -d postgres -u user -p pass -H localhost -o 5432 -n mydb
```

### Access the application
[http://localhost:8080](http://localhost:8080)

---

## Error Handling
- **Invalid credentials**: Ensure username/password are correct.
- **Connection issues**: Verify host and port.
- **File errors**: Ensure input/output file paths are valid.
- **Permission denied**: Run with appropriate permissions or check directory paths.

---

## Logging
Logs are saved to `logs/databaseutility.log`. Example log entries:
```bash
INFO: Database Utility Starts
INFO: Starting full backup of database mydb
INFO: Database backup completed successfully to backup.sql
ERROR: Failed to connect to the database: connection refused
```

---

## Future Enhancements
- Support for more databases (e.g., SQLite, Oracle).
- Incremental backups.
- Parallel table backups.
- Encryption for backup files.

---

## License
This project is licensed under the **MIT License**.

---

## Author
- **Yohan**  
- Email: yjirage@gmail.com
- GitHub: https://github.com/YohanJirage

---
