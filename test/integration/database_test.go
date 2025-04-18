package integration_test

import (
	"blog-platform/internal/database"
	"blog-platform/internal/server"
	"blog-platform/test/helpers"
	"log"

	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type BlogRepositorySuite struct {
	suite.Suite
	repository   *database.MongoBlogRepository
	testDatabase *helpers.TestDatabase
}

type SetupAllSuite interface {
	SetupSuite()
}

type TearDownAllSuite interface {
	TearDownSuite()
}

func (suite *BlogRepositorySuite) SetupSuite() {
	suite.testDatabase = helpers.SetupTestDatabase()
	suite.repository = suite.testDatabase.Repository
}

func (suite *BlogRepositorySuite) TearDownSuite() {
	err := suite.testDatabase.Container.Terminate(context.Background())
	log.Fatal(err)
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(BlogRepositorySuite))
}

func (suite *BlogRepositorySuite) TestCreateBlog() {
	suite.Run("Creates blog and is able to read the same blog", func() {
		e := echo.New()

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

		s := &server.Server{
			DB: suite.repository,
		}

		err := s.CreateBlogHandler(c)
		suite.NoError(err)
		suite.Equal(http.StatusCreated, rec.Code)

		var res map[string]string
		err = json.NewDecoder(rec.Body).Decode(&res)
		suite.NoError(err)

		fmt.Println(res)

		id, ok := res["data"]

		suite.True(ok, "expected ID in response")
		suite.NotEmpty(id)

		blog, err := suite.repository.GetBlog(context.Background(), id)
		suite.NoError(err)
		suite.Equal("My Test Blog", blog.Title)
		suite.Equal("Tech", blog.Category)
	})
}
