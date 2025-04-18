package e2e_test

import (
	"blog-platform/internal/server"
	"blog-platform/test/helpers"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/assert"
)

func TestE2E_CreateBlog(t *testing.T) {
	mongoContainer := helpers.SetupTestDatabase()

	fmt.Println("Created Database")
	s := server.NewServer()
	t.Log(s.Addr)
	fmt.Println("Created Server")

	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	waitForServer := func() {
		for i := 0; i < 10; i++ {
			resp, err := http.Get("http://localhost:8000/health")
			if err == nil && resp.StatusCode == 200 {
				return
			}
			time.Sleep(time.Second)
		}
		t.Fatal("Server did not become ready in time")
	}

	waitForServer()

	payload := `{"title":"My Blog","content":"Hello","tags":["go"]}`
	resp, err := http.Post("http://localhost:8000/posts", "application/json", strings.NewReader(payload))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var res map[string]string
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		t.Fatal("error decoding response")
	}
	assert.NotEmpty(t, res["data"])

	err = s.Shutdown(context.Background())
	if err != nil {
		log.Fatal("Error shutting down server")
	}
	err = mongoContainer.Container.Terminate(context.Background())
	if err != nil {
		log.Fatal("Error terminating mongo container. Should be cleaned up by Ryuk.")
	}

}
