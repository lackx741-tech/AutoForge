package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type SagaState string

const (
	StatePlan     SagaState = "PLAN"
	StateCodegen  SagaState = "CODEGEN"
	StateBuild    SagaState = "BUILD"
	StateTest     SagaState = "TEST"
	StateSecurity SagaState = "SECURITY"
	StatePackage  SagaState = "PACKAGE"
	StateRelease  SagaState = "RELEASE"
	StateComplete SagaState = "COMPLETE"
	StateRetry    SagaState = "RETRY"
	StateBlocked  SagaState = "BLOCKED"
	StateFatal    SagaState = "FATAL"
)

type SagaEvent struct {
	Timestamp  time.Time `json:"timestamp"`
	FromState  SagaState `json:"from_state"`
	ToState    SagaState `json:"to_state"`
	Reason     string    `json:"reason"`
	RetryCount int       `json:"retry_count"`
	SandboxRef string    `json:"sandbox_ref"`
	LogsURL    string    `json:"logs_url"`
}

type Saga struct {
	ID           string      `json:"id"`
	RepoURL      string      `json:"repo_url"`
	CommitSha    string      `json:"commit_sha"`
	CurrentState SagaState   `json:"current_state"`
	Events       []SagaEvent `json:"events"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	mu           sync.RWMutex
}

var sagas = make(map[string]*Saga)
var sagasMu sync.RWMutex

func createSaga(repoURL, commitSha string) *Saga {
	id := fmt.Sprintf("saga-%d", time.Now().UnixNano())
	saga := &Saga{
		ID:           id,
		RepoURL:      repoURL,
		CommitSha:    commitSha,
		CurrentState: StatePlan,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Events:       []SagaEvent{},
	}
	sagasMu.Lock()
	sagas[id] = saga
	sagasMu.Unlock()

	// Start async processing
	go processSaga(saga)

	return saga
}

func processSaga(s *Saga) {
	states := []SagaState{StatePlan, StateCodegen, StateBuild, StateSecurity, StateTest, StatePackage, StateRelease, StateComplete}

	for _, nextState := range states {
		time.Sleep(2 * time.Second) // Simulate work

		s.mu.Lock()
		event := SagaEvent{
			Timestamp:  time.Now(),
			FromState:  s.CurrentState,
			ToState:    nextState,
			Reason:     "automated_transition",
			RetryCount: 0,
			SandboxRef: fmt.Sprintf("sandbox-%s-%s", s.ID, nextState),
		}
		s.Events = append(s.Events, event)
		s.CurrentState = nextState
		s.UpdatedAt = time.Now()
		s.mu.Unlock()

		log.Printf("Saga %s transitioned: %s -> %s", s.ID, event.FromState, event.ToState)

		if nextState == StateComplete {
			break
		}
	}
}

func sagaHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var req struct {
			RepoURL   string `json:"repo_url"`
			CommitSha string `json:"commit_sha"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		saga := createSaga(req.RepoURL, req.CommitSha)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(saga)

	case "GET":
		id := r.URL.Query().Get("id")
		sagasMu.RLock()
		saga, exists := sagas[id]
		sagasMu.RUnlock()

		if !exists {
			http.Error(w, "Not found", 404)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(saga)
	}
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	sagasMu.RLock()
	defer sagasMu.RUnlock()

	result := make([]*Saga, 0, len(sagas))
	for _, s := range sagas {
		result = append(result, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	http.HandleFunc("/saga", sagaHandler)
	http.HandleFunc("/sagas", listHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	log.Printf("Saga Supervisor on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
