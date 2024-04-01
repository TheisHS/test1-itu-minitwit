package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)


func updateLatest(w http.ResponseWriter, r *http.Request) {
  parsedCommandID, err := strconv.Atoi(r.URL.Query().Get("latest"))
  if err != nil {
    return
  }

  if parsedCommandID != -1 {
    file, err := os.Create("./latest_processed_sim_action_id.txt")
    if err != nil {
      totalErrors.Inc()
      http.Error(w, "Failed to open file", http.StatusInternalServerError)
      return
    }
    defer file.Close()

    _, err = fmt.Fprintf(file, "%d", parsedCommandID)
    if err != nil {
      totalErrors.Inc()
      http.Error(w, "Failed to write to file", http.StatusInternalServerError)
      return
    }
  }    
}


func getLatestHandler(w http.ResponseWriter, r *http.Request) {
  _, err = os.Stat("./latest_processed_sim_action_id.txt")
  if err != nil {
    os.Create("./latest_processed_sim_action_id.txt")
  }

  file, err := os.ReadFile("./latest_processed_sim_action_id.txt")
  if err != nil {
    totalErrors.Inc()
    http.Error(w, "Failed to open file", http.StatusInternalServerError)
    return
  }

  content, err := strconv.Atoi(string(file)) 
  if err != nil {
    io.WriteString(w, `{"latest":-1}`)
  } else {
    io.WriteString(w, fmt.Sprintf(`{"latest":%d}`, content))
  }
}