package integration_test

import (
	"blog-platform/internal/database"
	"blog-platform/internal/server"
	"blog-platform/test/helpers"
	"encoding/json"
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

	createBlog := func(payload string) (*httptest.ResponseRecorder, error) {
		e := echo.New()

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

	deleteBlog := func(id string) (*httptest.ResponseRecorder, error) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodDelete, "/posts/:id", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(id)

		s := &server.Server{
			DB: suite.repository,
		}
		err := s.DeleteBlogHandler(c)
		if err != nil {
			return nil, err
		}
		return rec, nil
	}

	updateBlog := func(payload, id string) (*httptest.ResponseRecorder, error) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPut, "/posts/:id", strings.NewReader(payload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(id)

		s := &server.Server{
			DB: suite.repository,
		}
		err := s.UpdateBlogHandler(c)
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

		recDelete, err := deleteBlog(id)
		var resDelete map[string]interface{}
		err = json.NewDecoder(recDelete.Body).Decode(&resDelete)
		suite.NoError(err)
		title, ok := resDelete["title"]

		suite.True(ok, "expected title in response")
		suite.NotEmpty(title)
		suite.Equal("My Test Blog", title)

		blog, err = suite.repository.GetBlog(context.Background(), id)
		suite.Error(err)
	})

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
		suite.Equal("My Test Blog 2", (*blogs[0]).Title)
		suite.Equal("Example", (*blogs[0]).Category)
		suite.Equal("foo bar baz", (*blogs[0]).Content)
	})
	//
	suite.Run("Creates a blog and is able to update it", func() {
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

		recUpdate, err := updateBlog(`{"title": "My Test Blog 2"}`, id)
		var resUpdate map[string]interface{}
		err = json.NewDecoder(recUpdate.Body).Decode(&resUpdate)
		suite.NoError(err)
		title, ok := resUpdate["title"]

		suite.True(ok, "expected title in response")
		suite.NotEmpty(title)
		suite.Equal("My Test Blog 2", title)

		blog, err = suite.repository.GetBlog(context.Background(), id)
		suite.NoError(err)
		suite.Equal("My Test Blog 2", blog.Title)
	})
}
