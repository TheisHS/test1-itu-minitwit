package main

import (
	"database/sql"
	"log"
	"os"
	"strings"
)

var (
  db *sql.DB
	dbPath = "./data/minitwit.db"
)


func initDB() {
  log.Println("Initialising the database...")

  os.Create(dbPath)
  db, err := sql.Open("sqlite3", dbPath)
  if err != nil {
    log.Println(err)
  }
  
  schema, err := os.ReadFile("./schema.sql")
  if err != nil {
    log.Println(err) 
  }
  
  _, err = db.Exec(string(schema))
  if err != nil {
    log.Println(err) 
  }
  db.Close()
}


func connectDB() (*sql.DB, error) {
  promtailClient := getPromtailClient("connectDB")

  if env == "test" {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
      return nil, err
    }
    return db, nil
  }
  if env == "dev" {
    path, ok := os.LookupEnv("DATABASE_URL")
    if ok {
      connStr, err := os.ReadFile(path)
      if err != nil {
        db, err := sql.Open("postgres", strings.Trim(path, "\n"))
        if err != nil {
          promtailClient.Errorf("Could not read from filepath %s", path)
        }
        return db, nil
      }
      db, err := sql.Open("postgres", strings.Trim(string(connStr), "\n"))
      if err != nil {
        return nil, err
      }
      return db, nil
    }  
    promtailClient.Errorf("DATABASE_URL not set")
    panic("DATABASE_URL not set!")
  }
  if env == "prod" {
    path, ok := os.LookupEnv("DATABASE_URL")
    if ok {
      connStr, err := os.ReadFile(path)
      if err != nil {
        db, err := sql.Open("postgres", strings.Trim(path, "\n"))
        if err != nil {
          promtailClient.Errorf("Could not read from filepath %s", path)
        }
        return db, nil
      }
      db, err := sql.Open("postgres", strings.Trim(string(connStr), "\n"))
      if err != nil {
        return nil, err
      }
      return db, nil
    }  
    promtailClient.Errorf("DATABASE_URL not set")
    panic("DATABASE_URL not set!")
  }
  panic("Unknown environment")
}