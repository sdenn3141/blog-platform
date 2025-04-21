package integration_test

import (
	"blog-platform/internal/database"
	"blog-platform/internal/server"
	"blog-platform/test/helpers"
	"encoding/json"
	"fmt"
	"log"

	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	repository   *database.MongoBlogRepository
	testDatabase *helpers.TestDatabase
}

type SetupTestSuite interface {
	SetupSuite()
}

type TearDownTestSuite interface {
	TearDownSuite()
}

func (suite *IntegrationTestSuite) SetupSuite() {
	suite.testDatabase = helpers.SetupTestDatabase()
	suite.repository = suite.testDatabase.Repository
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	err := suite.testDatabase.Container.Terminate(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) TestSuite() {

	createBlog := func(b string) (*httptest.ResponseRecorder, error) {
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
		if err != nil {
			return nil, err
		}
		return rec, nil
	}

	suite.Run("Creates blog and is able to read and delete the same blog", func() {
		rec, err := createBlog(`{
			"title": "My Test Blog",
			"category": "Tech",
			"content": "This is some blog content.",
			"tags": ["go", "echo"]
		}`)
		suite.NoError(err)
		suite.Equal(http.StatusCreated, rec.Code)

		var res map[string]string
		err = json.NewDecoder(rec.Body).Decode(&res)
		suite.NoError(err)

		id, ok := res["data"]

		suite.True(ok, "expected ID in response")
		suite.NotEmpty(id)

		blog, err := suite.repository.GetBlog(context.Background(), id)
		suite.NoError(err)
		suite.Equal("My Test Blog", blog.Title)
		suite.Equal("Tech", blog.Category)

		blog, err = suite.repository.DeleteBlog(context.Background(), id)
		suite.NoError(err)
		suite.Equal("My Test Blog", blog.Title)
		suite.Equal("Tech", blog.Category)

		blog, err = suite.repository.GetBlog(context.Background(), id)
		suite.Error(err)
	})

	suite.Run("Creates blog and is able to read and delete the same blog", func() {
		rec, err := createBlog(`{
			"title": "My Test Blog",
			"category": "Tech",
			"content": "This is some blog content.",
			"tags": ["go", "echo"]
		}`)
		suite.NoError(err)
		suite.Equal(http.StatusCreated, rec.Code)

		var res map[string]string
		err = json.NewDecoder(rec.Body).Decode(&res)
		suite.NoError(err)

		id, ok := res["data"]

		suite.True(ok, "expected ID in response")
		suite.NotEmpty(id)

		blog, err := suite.repository.GetBlog(context.Background(), id)
		suite.NoError(err)
		suite.Equal("My Test Blog", blog.Title)
		suite.Equal("Tech", blog.Category)

		blog, err = suite.repository.DeleteBlog(context.Background(), id)
		suite.NoError(err)
		suite.Equal("My Test Blog", blog.Title)
		suite.Equal("Tech", blog.Category)

		blog, err = suite.repository.GetBlog(context.Background(), id)
		suite.Error(err)
	})
	//
	suite.Run("Creates multiple blogs and is able to search one by term", func() {
		_, err := createBlog(`{
			"title": "My Test Blog 1",
			"category": "Tech",
			"content": "This is some blog content.",
			"tags": ["case", "one"]
		}`)

		_, err = createBlog(`{
			"title": "My Test Blog 2",
			"category": "Example",
			"content": "foo bar baz",
			"tags": ["go", "two"]
		}`)

		blogs, err := suite.repository.GetBlogs(context.Background())
		suite.NoError(err)
		suite.Equal(2, len(blogs))

		blogs, err = suite.repository.GetBlogsByTerm(context.Background(), "foo")
		fmt.Println(err)
		suite.Equal("My Test Blog 2", (*blogs[0]).Title)
		suite.Equal("Example", (*blogs[0]).Category)
		suite.Equal("foo bar baz", (*blogs[0]).Content)
	})
}
