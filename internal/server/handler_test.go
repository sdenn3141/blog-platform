package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"blog-platform/internal/database"
	"blog-platform/internal/server"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTest() (*echo.Echo, *mockDB, time.Time) {
	e := echo.New()
	mockDB := new(mockDB)
	mockDate := time.Date(2025, time.April, 15, 10, 0, 0, 0, time.UTC)

	return e, mockDB, mockDate
}

func TestHealthHandler(t *testing.T) {
	e, mockDB, _ := setupTest()
	mockDB.On("Health", mock.Anything).Return(nil)
	s := &server.Server{
		DB: mockDB,
	}
	t.Run("Health Check runs properly", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		err := s.HealthHandler(c)
		assert.NoError(t, err)

		var res map[string]interface{}
		err = json.NewDecoder(rec.Body).Decode(&res)
		if err != nil {
			t.Fatal("error decoding response")
		}
		assert.Equal(t, "healthy", res["status"])
	})
}

func TestCreateBlogHandler(t *testing.T) {
	e, mockDB, _ := setupTest()
	createBlogId := "123"
	mockDB.On("CreateBlog", mock.Anything, mock.Anything).Return(&createBlogId, nil)
	s := &server.Server{
		DB: mockDB,
	}
	t.Run("Valid Blog Creation", func(t *testing.T) {
		payload := `{
			"title": "My Test Blog",
			"category": "Tech",
			"content": "This is some blog content.",
			"tags": ["go", "echo"]
		}`

		req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(payload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := s.CreateBlogHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var res map[string]interface{}
		err = json.NewDecoder(rec.Body).Decode(&res)
		if err != nil {
			t.Fatalf("error decoding response: %s", err)
		}
		assert.Equal(t, "123", res["data"])
	})

	t.Run("Invalid Payload", func(t *testing.T) {
		payload := `{ "title": 123 }`
		req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(payload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := s.CreateBlogHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var res map[string]interface{}
		err = json.NewDecoder(rec.Body).Decode(&res)
		if err != nil {
			t.Fatalf("error decoding response: %s", err)
		}
		assert.Equal(t, res["error"], "invalid request body")
	})
}

func TestGetBlogHandler(t *testing.T) {
	e, mockDB, mockDate := setupTest()
	mockGetResponse := database.Blog{Title: "Blog Title", Content: "My First Blog", Category: "Example", Tags: []string{"example"}, CreatedAt: mockDate, UpdatedAt: mockDate}
	mockDB.On("GetBlog", mock.Anything, mock.Anything).Return(&mockGetResponse, nil)
	s := &server.Server{
		DB: mockDB,
	}

	t.Run("Returns Valid Blog", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/posts/1234", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := s.GetBlogHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var res map[string]interface{}
		err = json.NewDecoder(rec.Body).Decode(&res)
		if err != nil {
			t.Fatalf("error decoding response: %s", err)
		}
		assert.Equal(t, "Blog Title", res["title"])
	})
}

func TestGetBlogsHandler(t *testing.T) {
	e, mockDB, mockDate := setupTest()
	mockDB.On("GetBlogs", mock.Anything).Return(
		[]*database.Blog{
			{Title: "Blog Title 3", Content: "My First Blog", Category: "Example", Tags: []string{"example"}, CreatedAt: mockDate, UpdatedAt: mockDate},
			{Title: "Blog Title 4", Content: "My First Blog", Category: "Example", Tags: []string{"example"}, CreatedAt: mockDate, UpdatedAt: mockDate},
		}, nil)
	mockDB.On("GetBlogsByTerm", mock.Anything, mock.Anything).Return(
		[]*database.Blog{
			{Title: "Blog Title 1", Content: "My First Blog", Category: "Example", Tags: []string{"example"}, CreatedAt: mockDate, UpdatedAt: mockDate},
			{Title: "Blog Title 2", Content: "My Second Blog", Category: "Example", Tags: []string{"example"}, CreatedAt: mockDate, UpdatedAt: mockDate},
		}, nil)
	s := &server.Server{
		DB: mockDB,
	}
	t.Run("Gets all blogs", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/posts/1234", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := s.GetBlogsHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var res []map[string]interface{}
		err = json.NewDecoder(rec.Body).Decode(&res)
		if err != nil {
			t.Fatalf("error decoding response: %s", err)
		}

		assert.Equal(t, "Blog Title 3", res[0]["title"])
		assert.Equal(t, "Blog Title 4", res[1]["title"])
	})

	t.Run("Search by Query is called when using a query parameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/posts/1234?term=example", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := s.GetBlogsHandler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var res []map[string]interface{}
		err = json.NewDecoder(rec.Body).Decode(&res)
		if err != nil {
			t.Fatalf("error decoding response: %s", err)
		}

		assert.Equal(t, "Blog Title 1", res[0]["title"])
		assert.Equal(t, "Blog Title 2", res[1]["title"])
	})
}
