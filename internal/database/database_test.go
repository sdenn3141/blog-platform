package database_test

import (
	"blog-platform/internal/database"
	"blog-platform/internal/dto"
	"blog-platform/test/helpers"
	"context"

	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/assert"

	"os"
	"testing"
)

func TestDatabaseLogic(t *testing.T) {
	t.Run("Test Create Blog", func(t *testing.T) {
		testDb := helpers.SetupTestDatabase()
		ctx := context.Background()
		settings := database.Settings{
			HostName:   os.Getenv("DB_HOST"),
			Port:       os.Getenv("DB_PORT"),
			Username:   os.Getenv("DB_USERNAME"),
			Password:   os.Getenv("DB_PASSWORD"),
			DbName:     os.Getenv("DB_NAME"),
			AuthSource: os.Getenv("DB_AUTHSOURCE"),
		}

		repository, err := database.New(settings)
		if err != nil {
			t.Logf("failed to create database %s", err)
		}

		b := dto.BlogCreateDto{
			Title:    "Test Blog",
			Category: "Test Category",
			Content:  "Test Blog",
			Tags:     []string{"Test Blog"},
		}

		id, err := repository.CreateBlog(ctx, b)
		result := *id

		assert.NoError(t, err)
		assert.IsType(t, "string", result)

		err = testDb.Container.Terminate(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	})
}
