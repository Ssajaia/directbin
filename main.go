package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type server struct {
	dataDir string
	newID   func() (string, error)
}

func newServer(dataDir string, newID func() (string, error)) (*server, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	return &server{dataDir: dataDir, newID: newID}, nil
}

func generateID() (string, error) {
	randomBytes := make([]byte, 6)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/health":
		s.health(w, r)
	case r.URL.Path == "/paste":
		s.create(w, r)
	case strings.HasPrefix(r.URL.Path, "/paste/"):
		s.paste(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (s *server) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *server) create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "could not read request body", http.StatusBadRequest)
		return
	}
	id, err := s.save(content)
	if err != nil {
		http.Error(w, "could not store paste", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"id": id})
}

func (s *server) paste(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/paste/"), "/")
	if len(parts) == 0 || !validID(parts[0]) {
		http.NotFound(w, r)
		return
	}

	switch {
	case len(parts) == 1 && r.Method == http.MethodGet:
		s.retrieve(w, r, parts[0])
	case len(parts) == 1 && r.Method == http.MethodDelete:
		s.delete(w, r, parts[0])
	case len(parts) == 2 && parts[1] == "raw" && r.Method == http.MethodGet:
		s.retrieve(w, r, parts[0])
	case len(parts) == 1:
		w.Header().Set("Allow", "GET, DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	case len(parts) == 2 && parts[1] == "raw":
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	default:
		http.NotFound(w, r)
	}
}

func (s *server) save(content []byte) (string, error) {
	for range 10 {
		id, err := s.newID()
		if err != nil {
			return "", err
		}
		if !validID(id) {
			return "", errors.New("generated invalid paste ID")
		}
		file, err := os.OpenFile(s.dataDir+string(os.PathSeparator)+id, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		_, writeErr := file.Write(content)
		closeErr := file.Close()
		if writeErr != nil {
			_ = os.Remove(s.dataDir + string(os.PathSeparator) + id)
			return "", writeErr
		}
		if closeErr != nil {
			_ = os.Remove(s.dataDir + string(os.PathSeparator) + id)
			return "", closeErr
		}
		return id, nil
	}
	return "", errors.New("could not generate a unique paste ID")
}

func (s *server) retrieve(w http.ResponseWriter, r *http.Request, id string) {
	content, err := os.ReadFile(s.dataDir + string(os.PathSeparator) + id)
	if errors.Is(err, os.ErrNotExist) {
		http.Error(w, "paste not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "could not read paste", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(content)
}

func (s *server) delete(w http.ResponseWriter, r *http.Request, id string) {
	err := os.Remove(s.dataDir + string(os.PathSeparator) + id)
	if errors.Is(err, os.ErrNotExist) {
		http.Error(w, "paste not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "could not delete paste", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func validID(id string) bool {
	if len(id) != 8 {
		return false
	}
	for _, char := range id {
		if !(char >= 'a' && char <= 'z') &&
			!(char >= 'A' && char <= 'Z') &&
			!(char >= '0' && char <= '9') &&
			char != '-' && char != '_' {
			return false
		}
	}
	return true
}

func main() {
	address := os.Getenv("ADDR")
	if address == "" {
		address = ":8080"
	}
	app, err := newServer("data", generateID)
	if err != nil {
		log.Fatal(err)
	}
	httpServer := &http.Server{Addr: address, Handler: app}
	shutdown := make(chan error, 1)
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(signals)
		<-signals
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		shutdown <- httpServer.Shutdown(ctx)
	}()
	log.Printf("DirectBin listening on %s", address)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
	if err := <-shutdown; err != nil {
		log.Fatal(err)
	}
}
