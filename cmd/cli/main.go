package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: autoforge [submit|status|list]")
		os.Exit(1)
	}

	command := os.Args[1]
	supervisorURL := getEnv("SUPERVISOR_URL", "http://localhost:8081")

	switch command {
	case "submit":
		if len(os.Args) < 4 {
			fmt.Println("Usage: autoforge submit <repo-url> <commit-sha>")
			os.Exit(1)
		}
		submit(supervisorURL, os.Args[2], os.Args[3])

	case "status":
		if len(os.Args) < 3 {
			fmt.Println("Usage: autoforge status <saga-id>")
			os.Exit(1)
		}
		status(supervisorURL, os.Args[2])

	case "list":
		list(supervisorURL)

	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func submit(url, repo, commit string) {
	payload := map[string]string{"repo_url": repo, "commit_sha": commit}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(url+"/saga", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Printf("Submitted saga: %s\n", result["id"])
	fmt.Printf("Current state: %s\n", result["current_state"])
	fmt.Printf("Track with: autoforge status %s\n", result["id"])
}

func status(url, id string) {
	resp, err := http.Get(fmt.Sprintf("%s/saga?id=%s", url, id))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var saga map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&saga)

	fmt.Printf("Saga: %s\n", saga["id"])
	fmt.Printf("State: %s\n", saga["current_state"])
	fmt.Printf("Repo: %s\n", saga["repo_url"])
	fmt.Printf("Created: %s\n", saga["created_at"])

	if events, ok := saga["events"].([]interface{}); ok {
		fmt.Printf("\nEvents (%d):\n", len(events))
		for _, e := range events {
			if event, ok := e.(map[string]interface{}); ok {
				fmt.Printf("  %s -> %s (%s)\n",
					event["from_state"], event["to_state"], event["timestamp"])
			}
		}
	}
}

func list(url string) {
	resp, err := http.Get(url + "/sagas")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var sagas []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&sagas)

	fmt.Printf("Active sagas: %d\n\n", len(sagas))
	for _, s := range sagas {
		fmt.Printf("%s | %s | %s | %s\n",
			s["id"], s["current_state"], s["repo_url"], s["updated_at"])
	}
}
