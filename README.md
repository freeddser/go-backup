# Database Backup Program

This is a concurrent database backup program written in Go. It reads a local JSON configuration file to get the database connection details, backs up the databases concurrently, and compresses the backup files using gzip.

## Features

1. Reads a local JSON configuration file for database details.
2. Backs up multiple databases concurrently.
3. Compresses backup files using gzip.
4. Allows specifying the number of concurrent backups.
5. Logs backup status to a file.
6. Lists backed-up files.

## Configuration

Create a `config.json` file in the same directory as the program with the following structure:

```json
{
  "backup_target_path": "/path/to/backup",
  "dblists": [
    {
      "db_number": "101",
      "db_name": "abc1",
      "db_user": "backupuser",
      "db_password": "password",
      "db_host": "127.0.0.1:3306",
      "remark": ""
    },
    {
      "db_number": "102",
      "db_name": "abc1",
      "db_user": "backupuser",
      "db_password": "password",
      "db_host": "127.0.0.1:3306",
      "remark": ""
    }
  ]
}
```

## Usage

##### Backup Databases

To start the backup process, use the following command:

```go run main.go -action backup```

By default, the program runs with a concurrency of 3. You can specify a different concurrency level using the -concurrency flag:

```go run main.go -action backup -concurrency 5```

##### List Backups
To list the backed-up files, use the following command:

```go run main.go -action list```


## Logging
The program logs backup status and errors to backup.log in the same directory as the program.

##  Dependencies
Ensure you have mysqldump and gzip installed on your system, as they are used for database backup and compression.

## Example
##### Here's an example of how to use the program:
    1.Create a config.json file with your database details and backup target path.
      create the backup path: /path/to/backup
      add your mysql db account in the config.json file
    2.Run the backup command:
        go run main.go -action backup -concurrency 3
    3. Check the backup.log file for backup status and errors.
    4. List the backed-up files:
        go run main.go -action list


