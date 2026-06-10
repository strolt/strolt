package e2e_test

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"
	"github.com/testcontainers/testcontainers-go/modules/minio"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/sync/errgroup"
)

// Service images used by the e2e environment. Keep the database server
// versions compatible with the client tools baked into
// docker/strolt/Dockerfile when bumping these.
const (
	postgresImage = "postgres:18.4-alpine3.23"
	mongoImage    = "mongo:8.0.23"
	mariadbImage  = "mariadb:11.4.12"
	mysqlImage    = "mysql:8.4.9"
	minioImage    = "minio/minio:RELEASE.2025-09-07T16-13-09Z"
	stroltImage   = "strolt/strolt:development"

	dbName     = "strolt"
	dbUser     = "strolt"
	dbPassword = "strolt" // pragma: allowlist secret
)

// ContainerManager manages all testcontainers for e2e tests.
type ContainerManager struct {
	ctx                   context.Context //nolint:containedctx // test helper carries the suite context
	network               *testcontainers.DockerNetwork
	postgresContainer     testcontainers.Container
	mongoContainer        testcontainers.Container
	mariadbContainer      testcontainers.Container
	mysqlContainer        testcontainers.Container
	minioContainer        testcontainers.Container
	stroltContainer       testcontainers.Container
	stroltDaemonContainer testcontainers.Container
}

// NewContainerManager creates a new container manager.
func NewContainerManager(ctx context.Context) (*ContainerManager, error) {
	return &ContainerManager{
		ctx: ctx,
	}, nil
}

// SetupNetwork creates a Docker network for containers.
func (cm *ContainerManager) SetupNetwork() error {
	nw, err := network.New(cm.ctx)
	if err != nil {
		return err
	}

	cm.network = nw

	return nil
}

// NetworkName returns the name of the Docker network shared by the containers.
func (cm *ContainerManager) NetworkName() string {
	return cm.network.Name
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
		Image: stroltImage,
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      absConfigPath,
				ContainerFilePath: "/strolt/config.yml",
				FileMode:          0o644,
			},
		},
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			hostConfig.Binds = append(hostConfig.Binds,
				absStroltPath+":/strolt/.strolt",
				absTempPath+":/e2e/input",
			)
		},
		Networks:       []string{cm.NetworkName()},
		NetworkAliases: map[string][]string{cm.NetworkName(): {"strolt"}},
		Entrypoint:     []string{"/bin/sh"},
		Cmd:            []string{"-c", "sleep infinity"},
		WaitingFor: wait.ForExec([]string{"sh", "-c", "test -d /e2e/input"}).
			WithExitCodeMatcher(func(exitCode int) bool { return exitCode == 0 }).
			WithStartupTimeout(60 * time.Second),
	}

	stroltContainer, err := testcontainers.GenericContainer(cm.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		if stroltContainer != nil {
			_ = stroltContainer.Terminate(cm.ctx)
		}

		return err
	}

	cm.stroltContainer = stroltContainer

	return nil
}

// StartStroltDaemon starts strolt in daemon mode (`start --json`) with the
// HTTP API exposed, mirroring how the image runs in production.
func (cm *ContainerManager) StartStroltDaemon() error {
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
		Image: stroltImage,
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      absConfigPath,
				ContainerFilePath: "/strolt/config.yml",
				FileMode:          0o644,
			},
		},
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			hostConfig.Binds = append(hostConfig.Binds,
				absStroltPath+":/strolt/.strolt",
				absTempPath+":/e2e/input",
			)
		},
		Networks:       []string{cm.NetworkName()},
		NetworkAliases: map[string][]string{cm.NetworkName(): {"strolt-daemon"}},
		ExposedPorts:   []string{"8080/tcp"},
		Cmd:            []string{"start", "--json"},
		WaitingFor: wait.ForHTTP("/api/v1/ping").
			WithPort("8080/tcp").
			WithStartupTimeout(60 * time.Second),
	}

	daemonContainer, err := testcontainers.GenericContainer(cm.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		if daemonContainer != nil {
			_ = daemonContainer.Terminate(cm.ctx)
		}

		return err
	}

	cm.stroltDaemonContainer = daemonContainer

	return nil
}

// GetDaemonAPIPort returns the mapped port of the strolt daemon HTTP API.
func (cm *ContainerManager) GetDaemonAPIPort() (string, error) {
	if cm.stroltDaemonContainer == nil {
		return "", errors.New("strolt daemon container not started")
	}

	mappedPort, err := cm.stroltDaemonContainer.MappedPort(cm.ctx, "8080")
	if err != nil {
		return "", err
	}

	return mappedPort.Port(), nil
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

// Cleanup terminates all containers and removes the network. It uses its own
// context so cleanup still runs when the suite context is already done.
func (cm *ContainerManager) Cleanup() error {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var errs []error

	containers := []testcontainers.Container{
		cm.stroltContainer,
		cm.stroltDaemonContainer,
		cm.postgresContainer,
		cm.mongoContainer,
		cm.mariadbContainer,
		cm.mysqlContainer,
		cm.minioContainer,
	}

	for _, c := range containers {
		if c != nil {
			if err := c.Terminate(cleanupCtx); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if cm.network != nil {
		if err := cm.network.Remove(cleanupCtx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// StartAllContainers starts all containers in parallel.
func (cm *ContainerManager) StartAllContainers() error {
	g := new(errgroup.Group)

	g.Go(func() error {
		c, err := cm.startPostgres()
		cm.postgresContainer = c

		return err
	})
	g.Go(func() error {
		c, err := cm.startMongo()
		cm.mongoContainer = c

		return err
	})
	g.Go(func() error {
		c, err := cm.startMariaDB()
		cm.mariadbContainer = c

		return err
	})
	g.Go(func() error {
		c, err := cm.startMySQL()
		cm.mysqlContainer = c

		return err
	})
	g.Go(func() error {
		c, err := cm.startMinio()
		cm.minioContainer = c

		return err
	})

	return g.Wait()
}

// startPostgres starts PostgreSQL via the official testcontainers module.
// BasicWaitStrategies handles the temporary-server restart performed by the
// image entrypoint during initialization, which a plain SQL probe races with.
func (cm *ContainerManager) startPostgres() (testcontainers.Container, error) {
	c, err := postgres.Run(cm.ctx, postgresImage,
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.BasicWaitStrategies(),
		testcontainers.WithEnv(map[string]string{"TZ": "UTC"}),
		network.WithNetwork([]string{"postgres"}, cm.network),
	)
	if err != nil {
		if c != nil {
			_ = c.Terminate(cm.ctx)
		}

		return nil, err
	}

	return c, nil
}

func (cm *ContainerManager) startMongo() (testcontainers.Container, error) {
	c, err := mongodb.Run(cm.ctx, mongoImage,
		network.WithNetwork([]string{"mongo"}, cm.network),
	)
	if err != nil {
		if c != nil {
			_ = c.Terminate(cm.ctx)
		}

		return nil, err
	}

	return c, nil
}

func (cm *ContainerManager) startMariaDB() (testcontainers.Container, error) {
	c, err := mariadb.Run(cm.ctx, mariadbImage,
		mariadb.WithDatabase(dbName),
		mariadb.WithUsername(dbUser),
		mariadb.WithPassword(dbPassword),
		testcontainers.WithEnv(map[string]string{"TZ": "UTC"}),
		network.WithNetwork([]string{"mariadb"}, cm.network),
	)
	if err != nil {
		if c != nil {
			_ = c.Terminate(cm.ctx)
		}

		return nil, err
	}

	return c, nil
}

func (cm *ContainerManager) startMySQL() (testcontainers.Container, error) {
	c, err := mysql.Run(cm.ctx, mysqlImage,
		mysql.WithDatabase(dbName),
		mysql.WithUsername(dbUser),
		mysql.WithPassword(dbPassword),
		testcontainers.WithEnv(map[string]string{"TZ": "UTC"}),
		network.WithNetwork([]string{"mysql"}, cm.network),
	)
	if err != nil {
		if c != nil {
			_ = c.Terminate(cm.ctx)
		}

		return nil, err
	}

	return c, nil
}

// startMinio starts MinIO via the official testcontainers module. The module
// defaults to minioadmin/minioadmin credentials, matching .strolt/secrets.yml.
func (cm *ContainerManager) startMinio() (testcontainers.Container, error) {
	c, err := minio.Run(cm.ctx, minioImage,
		network.WithNetwork([]string{"minio"}, cm.network),
	)
	if err != nil {
		if c != nil {
			_ = c.Terminate(cm.ctx)
		}

		return nil, err
	}

	return c, nil
}
