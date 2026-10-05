package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testServer(t *testing.T, newID func() (string, error)) *server {
	t.Helper()
	app, err := newServer(t.TempDir(), newID)
	if err != nil {
		t.Fatal(err)
	}
	return app
}

func TestCreateRetrieveRawAndDeletePaste(t *testing.T) {
	app := testServer(t, func() (string, error) {
		return "a8K29xQ_", nil
	})
	content := "hello from another computer\n"

	createRequest := httptest.NewRequest(http.MethodPost, "/paste", strings.NewReader(content))
	createResponse := httptest.NewRecorder()
	app.ServeHTTP(createResponse, createRequest)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createResponse.Code, http.StatusCreated)
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(createResponse.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.ID != "a8K29xQ_" {
		t.Fatalf("created ID = %q, want %q", created.ID, "a8K29xQ_")
	}

	for _, path := range []string{"/paste/" + created.ID, "/paste/" + created.ID + "/raw"} {
		response := httptest.NewRecorder()
		app.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want %d", path, response.Code, http.StatusOK)
		}
		if response.Body.String() != content {
			t.Errorf("GET %s body = %q, want %q", path, response.Body.String(), content)
		}
	}

	deleteResponse := httptest.NewRecorder()
	app.ServeHTTP(deleteResponse, httptest.NewRequest(http.MethodDelete, "/paste/"+created.ID, nil))
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d", deleteResponse.Code, http.StatusNoContent)
	}

	missingResponse := httptest.NewRecorder()
	app.ServeHTTP(missingResponse, httptest.NewRequest(http.MethodGet, "/paste/"+created.ID, nil))
	if missingResponse.Code != http.StatusNotFound {
		t.Fatalf("retrieval after delete status = %d, want %d", missingResponse.Code, http.StatusNotFound)
	}
}

func TestPasteNotFound(t *testing.T) {
	app := testServer(t, generateID)
	for _, method := range []string{http.MethodGet, http.MethodDelete} {
		response := httptest.NewRecorder()
		app.ServeHTTP(response, httptest.NewRequest(method, "/paste/a8K29xQ_", nil))
		if response.Code != http.StatusNotFound {
			t.Errorf("%s missing paste status = %d, want %d", method, response.Code, http.StatusNotFound)
		}
	}
}

func TestCreateDoesNotOverwriteOnIDCollision(t *testing.T) {
	dataDir := t.TempDir()
	existingContent := []byte("keep this paste")
	if err := os.WriteFile(filepath.Join(dataDir, "a8K29xQ_"), existingContent, 0600); err != nil {
		t.Fatal(err)
	}
	ids := []string{"a8K29xQ_", "f72LmQ2_"}
	app, err := newServer(dataDir, func() (string, error) {
		id := ids[0]
		ids = ids[1:]
		return id, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()
	app.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/paste", strings.NewReader("new paste")))
	if response.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", response.Code, http.StatusCreated)
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.ID != "f72LmQ2_" {
		t.Fatalf("created ID = %q, want %q", created.ID, "f72LmQ2_")
	}
	storedContent, err := os.ReadFile(filepath.Join(dataDir, "a8K29xQ_"))
	if err != nil {
		t.Fatal(err)
	}
	if string(storedContent) != string(existingContent) {
		t.Fatalf("collision changed existing content to %q", storedContent)
	}
}

func TestHealth(t *testing.T) {
	app := testServer(t, generateID)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", response.Code, http.StatusOK)
	}
	if got, want := strings.TrimSpace(response.Body.String()), `{"status":"ok"}`; got != want {
		t.Fatalf("health body = %q, want %q", got, want)
	}
}

func TestUnsupportedMethod(t *testing.T) {
	app := testServer(t, generateID)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, httptest.NewRequest(http.MethodPut, "/paste", nil))

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unsupported method status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
	if got := response.Header().Get("Allow"); got != http.MethodPost {
		t.Fatalf("Allow = %q, want %q", got, http.MethodPost)
	}
}

func TestGeneratedIDIsURLSafe(t *testing.T) {
	id, err := generateID()
	if err != nil {
		t.Fatal(err)
	}
	if !validID(id) {
		t.Fatalf("generated ID %q is not valid", id)
	}
}
