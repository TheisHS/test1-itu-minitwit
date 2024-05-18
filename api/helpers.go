package main

import (
	"fmt"
	"time"

	"github.com/afiskon/promtail-client/promtail"
)


func devLog(str string) {
  if env == "dev" {
    fmt.Println(str)
  }
}

func getPromtailClient(f string) (promtail.Client) {
  conf := promtail.ClientConfig{
		PushURL:            "http://loki:3100/api/prom/push",
		Labels:             fmt.Sprintf(`{source="minitwit-api", function="%s"}`, f),
		BatchWait:          5 * time.Second,
		BatchEntriesNumber: 10000,
		SendLevel: 			promtail.INFO,
		PrintLevel: 		promtail.ERROR,
	}
  promtailClient, err := promtail.NewClientJson(conf)
  if err != nil {
    return nil
  }
  return promtailClient
}