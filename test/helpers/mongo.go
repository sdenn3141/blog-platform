package helpers

import (
	"blog-platform/internal/database"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestDatabase struct {
	Repository *database.MongoBlogRepository
	DbAddress  string
	Container  testcontainers.Container
}

func SetupTestDatabase() *TestDatabase {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	container, repository, dbAddr, err := createMongoContainer(ctx)
	if err != nil {
		log.Fatal("failed to setup test", err)
	}

	return &TestDatabase{
		Container:  container,
		Repository: repository,
		DbAddress:  dbAddr,
	}
}

func (tdb *TestDatabase) TearDown() {
	_ = tdb.Container.Terminate(context.Background())
}

func createMongoContainer(ctx context.Context) (testcontainers.Container, *database.MongoBlogRepository, string, error) {
	env := map[string]string{
		"MONGO_INITDB_ROOT_USERNAME": "root",
		"MONGO_INITDB_ROOT_PASSWORD": "pass",
		"MONGO_INITDB_DATABASE":      "blogs",
	}
	port := "27017:27017"

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mongo",
			ExposedPorts: []string{port},
			Env:          env,
			WaitingFor:   wait.ForListeningPort("27017/tcp"),
		},
		Started: true,
	}
	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return container, nil, "", fmt.Errorf("failed to start container: %v", err)
	}

	p, err := container.MappedPort(ctx, "27017")
	if err != nil {
		return container, nil, "", fmt.Errorf("failed to get container external port: %v", err)
	}

	log.Println("mongo container ready and running at port: ", p.Port())

	uri := fmt.Sprintf("mongodb://localhost:%s", p.Port())
	dbSettings := database.Settings{
		HostName:   os.Getenv("DB_HOST"),
		Port:       os.Getenv("DB_PORT"),
		Username:   os.Getenv("DB_USERNAME"),
		Password:   os.Getenv("DB_PASSWORD"),
		DbName:     os.Getenv("DB_DATABASE"),
		AuthSource: os.Getenv("DB_AUTHSOURCE"),
	}
	db, err := database.New(dbSettings)
	if err != nil {
		return container, db, uri, fmt.Errorf("failed to establish database connection: %v", err)
	}

	return container, db, uri, nil
}
