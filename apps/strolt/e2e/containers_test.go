package e2e_test

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ContainerManager manages all testcontainers for e2e tests.
type ContainerManager struct {
	ctx               context.Context
	network           testcontainers.Network
	postgresContainer testcontainers.Container
	mongoContainer    testcontainers.Container
	mariadbContainer  testcontainers.Container
	mysqlContainer    testcontainers.Container
	minioContainer    testcontainers.Container
	stroltContainer   testcontainers.Container
}

// NewContainerManager creates a new container manager.
func NewContainerManager(ctx context.Context) (*ContainerManager, error) {
	return &ContainerManager{
		ctx: ctx,
	}, nil
}

// SetupNetwork creates a Docker network for containers.
func (cm *ContainerManager) SetupNetwork() error {
	network, err := testcontainers.GenericNetwork(cm.ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name:           "strolt",
			CheckDuplicate: true,
		},
	})
	if err != nil {
		return err
	}

	cm.network = network

	return nil
}

// StartStrolt starts the Strolt container.
func (cm *ContainerManager) StartStrolt() error {
	absConfigPath, err := filepath.Abs("./strolt.yml")
	if err != nil {
		return err
	}

	absStroltPath, err := filepath.Abs("./.strolt")
	if err != nil {
		return err
	}

	absTempPath, err := filepath.Abs("./.temp/input")
	if err != nil {
		return err
	}

	req := testcontainers.ContainerRequest{
		Image: "strolt/strolt:development",
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      absConfigPath,
				ContainerFilePath: "/strolt/config.yml",
				FileMode:          0644,
			},
		},
		Mounts: testcontainers.Mounts(
			testcontainers.BindMount(absStroltPath, "/strolt/.strolt"),
			testcontainers.BindMount(absTempPath, "/e2e/input"),
		),
		Networks:       []string{"strolt"},
		NetworkAliases: map[string][]string{"strolt": {"strolt"}},
		Entrypoint:     []string{"/bin/sh"},
		Cmd:            []string{"-c", "sleep infinity"},
		WaitingFor:     wait.ForExec([]string{"sh", "-c", "test -d /e2e/input"}).WithExitCodeMatcher(func(exitCode int) bool { return exitCode == 0 }),
	}

	container, err := testcontainers.GenericContainer(cm.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return err
	}

	cm.stroltContainer = container

	return nil
}

// GetPostgresPort returns the mapped port for PostgreSQL.
func (cm *ContainerManager) GetPostgresPort() (string, error) {
	if cm.postgresContainer == nil {
		return "", errors.New("postgres container not started")
	}

	mappedPort, err := cm.postgresContainer.MappedPort(cm.ctx, "5432")
	if err != nil {
		return "", err
	}

	return mappedPort.Port(), nil
}

// GetMongoPort returns the mapped port for MongoDB.
func (cm *ContainerManager) GetMongoPort() (string, error) {
	if cm.mongoContainer == nil {
		return "", errors.New("mongo container not started")
	}

	mappedPort, err := cm.mongoContainer.MappedPort(cm.ctx, "27017")
	if err != nil {
		return "", err
	}

	return mappedPort.Port(), nil
}

// GetMariaDBPort returns the mapped port for MariaDB.
func (cm *ContainerManager) GetMariaDBPort() (string, error) {
	if cm.mariadbContainer == nil {
		return "", errors.New("mariadb container not started")
	}

	mappedPort, err := cm.mariadbContainer.MappedPort(cm.ctx, "3306")
	if err != nil {
		return "", err
	}

	return mappedPort.Port(), nil
}

// GetMySQLPort returns the mapped port for MySQL.
func (cm *ContainerManager) GetMySQLPort() (string, error) {
	if cm.mysqlContainer == nil {
		return "", errors.New("mysql container not started")
	}

	mappedPort, err := cm.mysqlContainer.MappedPort(cm.ctx, "3306")
	if err != nil {
		return "", err
	}

	return mappedPort.Port(), nil
}

// GetStroltContainer returns the Strolt container.
func (cm *ContainerManager) GetStroltContainer() testcontainers.Container {
	return cm.stroltContainer
}

// Cleanup terminates all containers and removes the network.
func (cm *ContainerManager) Cleanup() error {
	var errs []error

	if cm.stroltContainer != nil {
		if err := cm.stroltContainer.Terminate(cm.ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if cm.postgresContainer != nil {
		if err := cm.postgresContainer.Terminate(cm.ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if cm.mongoContainer != nil {
		if err := cm.mongoContainer.Terminate(cm.ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if cm.mariadbContainer != nil {
		if err := cm.mariadbContainer.Terminate(cm.ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if cm.mysqlContainer != nil {
		if err := cm.mysqlContainer.Terminate(cm.ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if cm.minioContainer != nil {
		if err := cm.minioContainer.Terminate(cm.ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if cm.network != nil {
		if err := cm.network.Remove(cm.ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	return nil
}

// StartAllContainers starts all containers in parallel.
func (cm *ContainerManager) StartAllContainers() error {
	type containerResult struct {
		name      string
		container testcontainers.Container
		err       error
	}

	results := make(chan containerResult, 5)

	// Start containers in parallel
	go func() {
		container, err := cm.startPostgres()
		results <- containerResult{"postgres", container, err}
	}()
	go func() {
		container, err := cm.startMongo()
		results <- containerResult{"mongo", container, err}
	}()
	go func() {
		container, err := cm.startMariaDB()
		results <- containerResult{"mariadb", container, err}
	}()
	go func() {
		container, err := cm.startMySQL()
		results <- containerResult{"mysql", container, err}
	}()
	go func() {
		container, err := cm.startMinio()
		results <- containerResult{"minio", container, err}
	}()

	// Collect results
	for i := 0; i < 5; i++ {
		result := <-results
		if result.err != nil {
			return fmt.Errorf("failed to start %s: %w", result.name, result.err)
		}

		switch result.name {
		case "postgres":
			cm.postgresContainer = result.container
		case "mongo":
			cm.mongoContainer = result.container
		case "mariadb":
			cm.mariadbContainer = result.container
		case "mysql":
			cm.mysqlContainer = result.container
		case "minio":
			cm.minioContainer = result.container
		}
	}

	return nil
}

func (cm *ContainerManager) startPostgres() (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13.2-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"TZ":                "UTC",
			"POSTGRES_DB":       "strolt",
			"POSTGRES_PASSWORD": "strolt",
			"POSTGRES_USER":     "strolt",
		},
		Networks:       []string{"strolt"},
		NetworkAliases: map[string][]string{"strolt": {"postgres"}},
		WaitingFor:     wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	return testcontainers.GenericContainer(cm.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

func (cm *ContainerManager) startMongo() (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "mongo:4.4.15",
		ExposedPorts: []string{"27017/tcp"},
		Env: map[string]string{
			"PUID": "1000",
			"PGID": "1000",
		},
		Networks:       []string{"strolt"},
		NetworkAliases: map[string][]string{"strolt": {"mongo"}},
		WaitingFor:     wait.ForListeningPort("27017/tcp").WithStartupTimeout(30 * time.Second),
	}

	return testcontainers.GenericContainer(cm.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

func (cm *ContainerManager) startMariaDB() (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "mariadb:10.8.3",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"TZ":                  "UTC",
			"MYSQL_DATABASE":      "strolt",
			"MYSQL_USER":          "strolt",
			"MYSQL_PASSWORD":      "strolt",
			"MYSQL_ROOT_PASSWORD": "strolt",
		},
		Networks:       []string{"strolt"},
		NetworkAliases: map[string][]string{"strolt": {"mariadb"}},
		WaitingFor:     wait.ForListeningPort("3306/tcp").WithStartupTimeout(30 * time.Second),
	}

	return testcontainers.GenericContainer(cm.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

func (cm *ContainerManager) startMySQL() (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "mysql:8.0.30",
		ExposedPorts: []string{"3306/tcp"},
		Cmd:          []string{"mysqld", "--default-authentication-plugin=mysql_native_password"},
		Env: map[string]string{
			"TZ":                  "UTC",
			"MYSQL_DATABASE":      "strolt",
			"MYSQL_USER":          "strolt",
			"MYSQL_PASSWORD":      "strolt",
			"MYSQL_ROOT_PASSWORD": "strolt",
		},
		Networks:       []string{"strolt"},
		NetworkAliases: map[string][]string{"strolt": {"mysql"}},
		WaitingFor:     wait.ForListeningPort("3306/tcp").WithStartupTimeout(30 * time.Second),
	}

	return testcontainers.GenericContainer(cm.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

func (cm *ContainerManager) startMinio() (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "minio/minio:RELEASE.2022-08-13T21-54-44Z",
		ExposedPorts: []string{"9000/tcp", "9001/tcp"},
		Cmd:          []string{"server", "--address", "0.0.0.0:9000", "--console-address", "0.0.0.0:9001", "/data"},
		Env: map[string]string{
			"MINIO_ROOT_USER":     "minioadmin",
			"MINIO_ROOT_PASSWORD": "minioadmin",
		},
		Networks:       []string{"strolt"},
		NetworkAliases: map[string][]string{"strolt": {"minio"}},
		WaitingFor:     wait.ForHTTP("/minio/health/live").WithPort("9000/tcp").WithStartupTimeout(30 * time.Second),
	}

	return testcontainers.GenericContainer(cm.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

// GetContainerName returns the container name for a given container.
func GetContainerName(ctx context.Context, container testcontainers.Container) (string, error) {
	if container == nil {
		return "", errors.New("container is nil")
	}

	name, err := container.Name(ctx)
	if err != nil {
		return "", err
	}
	// Remove the leading "/" from the container name
	if len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}

	return name, nil
}
