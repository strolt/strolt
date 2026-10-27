package e2e_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"
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
	postgresImage = "postgres:18.6-alpine3.23"
	mongoImage    = "mongo:9.0.4"
	mariadbImage  = "mariadb:11.4.13"
	mysqlImage    = "mysql:8.4.13"
	// Valkey speaks the Redis protocol and CLI, so the redis source driver
	// backs it up too; only the host differs.
	redisImage  = "redis:7.4.12-alpine"
	valkeyImage = "valkey/valkey:8.1.10-alpine"
	// Garage replaces MinIO: the MinIO community images were pulled from
	// Docker Hub and quay, and restic's S3 backend works against Garage.
	// The container is aliased "minio" so task configs keep their
	// s3:http://minio:9000 endpoints.
	garageImage = "dxflrs/garage:v2.5.0"
	stroltImage = "strolt/strolt:development"

	dbName     = "strolt"
	dbUser     = "strolt"
	dbPassword = "strolt" // pragma: allowlist secret

	// redisPassword protects the Valkey instance; the valkey task embeds it
	// in its source URI, exercising credential normalization in e2e.
	redisPassword = "e2e-valkey-pass" // pragma: allowlist secret

	// Fixed S3 credentials imported into Garage; .strolt/secrets.yml and
	// resticExec use the same pair.
	s3AccessKeyID     = "GK6d6e2e2a0c1b4f9d8a7e3c5b1d9f0a2e"                               // pragma: allowlist secret; gitleaks:allow
	s3SecretAccessKey = "4c7d1f8a3b6e9c2d5f0a1b4e7c8d3f6a9b2e5d8c1f4a7b0e3d6c9f2a5b8e1d4c" // pragma: allowlist secret; gitleaks:allow
)

// ContainerManager manages all testcontainers for e2e tests.
type ContainerManager struct {
	ctx                   context.Context //nolint:containedctx // test helper carries the suite context
	network               *testcontainers.DockerNetwork
	postgresContainer     testcontainers.Container
	mongoContainer        testcontainers.Container
	mariadbContainer      testcontainers.Container
	mysqlContainer        testcontainers.Container
	redisContainer        testcontainers.Container
	valkeyContainer       testcontainers.Container
	garageContainer       testcontainers.Container
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
		cm.redisContainer,
		cm.valkeyContainer,
		cm.garageContainer,
		cm.mysqlContainer,
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
		c, err := cm.startRedis()
		cm.redisContainer = c

		return err
	})
	g.Go(func() error {
		c, err := cm.startValkey()
		cm.valkeyContainer = c

		return err
	})
	g.Go(func() error {
		c, err := cm.startGarage()
		cm.garageContainer = c

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

// startCache starts a Redis or Valkey instance for the redis source driver;
// both images log the same readiness line and speak the same protocol, so
// only the image, network alias and server args differ.
func (cm *ContainerManager) startCache(image, alias string, cmd ...string) (testcontainers.Container, error) {
	req := testcontainers.GenericContainerRequest{
		Image:          image,
		Networks:       []string{cm.network.Name},
		NetworkAliases: map[string][]string{cm.network.Name: {alias}},
		WaitingFor:     wait.ForLog("Ready to accept connections").WithOccurrence(1),
		Started:        true,
	}
	if len(cmd) > 0 {
		req.Cmd = cmd
	}

	c, err := testcontainers.GenericContainer(cm.ctx, req)
	if err != nil {
		if c != nil {
			_ = c.Terminate(cm.ctx)
		}

		return nil, err
	}

	return c, nil
}

func (cm *ContainerManager) startRedis() (testcontainers.Container, error) {
	return cm.startCache(redisImage, "redis")
}

// startValkey runs with a password so the valkey task exercises the driver's
// URI-credential normalization and REDISCLI_AUTH delivery end to end.
func (cm *ContainerManager) startValkey() (testcontainers.Container, error) {
	return cm.startCache(valkeyImage, "valkey", "--requirepass", redisPassword)
}

// startGarage starts Garage, the S3-compatible object store that replaced
// MinIO (whose community images no longer exist on any registry). The
// container keeps the "minio" network alias and port 9000 so existing task
// configs work unchanged; after startup the single-node layout is applied
// and the fixed key from .strolt/secrets.yml is imported with permission to
// create buckets (every task provisions its own repository).
func (cm *ContainerManager) startGarage() (testcontainers.Container, error) {
	req := testcontainers.GenericContainerRequest{
		Image:        garageImage,
		Cmd:          []string{"/garage", "server"},
		ExposedPorts: []string{"9000/tcp"},
		// Attach explicitly by network name and register the "minio"
		// alias in the same declaration: an aliases-only attach proved
		// unreliable on CI runners (intermittent "lookup minio: no such
		// host" mid-suite, which restic turns into minutes of retries).
		Networks:       []string{cm.network.Name},
		NetworkAliases: map[string][]string{cm.network.Name: {"minio"}},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      "garage.toml",
				ContainerFilePath: "/etc/garage.toml",
			},
		},
		WaitingFor: wait.ForHTTP("/").
			WithPort("9000/tcp").
			WithStatusCodeMatcher(func(status int) bool { return status == 200 || status == 403 }),
		Started: true,
	}

	c, err := testcontainers.GenericContainer(cm.ctx, req)
	if err != nil {
		return nil, err
	}

	if err := cm.provisionGarage(c); err != nil {
		_ = c.Terminate(cm.ctx)
		return nil, err
	}

	return c, nil
}

// provisionGarage applies the storage layout and imports the fixed S3 key.
func (cm *ContainerManager) provisionGarage(c testcontainers.Container) error {
	statusOut, err := execInGarage(cm, c, "status")
	if err != nil {
		return err
	}

	nodeID := garageNodeID(statusOut)
	if nodeID == "" {
		return fmt.Errorf("garage status: node id not found, output was %q", statusOut)
	}

	steps := [][]string{
		{"layout", "assign", "-z", "dc1", "-c", "1G", nodeID},
		{"layout", "apply", "--version", "1"},
		{"key", "import", s3AccessKeyID, s3SecretAccessKey, "--yes", "-n", "e2e"},
		{"key", "allow", "--create-bucket", s3AccessKeyID},
	}

	for _, args := range steps {
		if _, err := execInGarage(cm, c, args...); err != nil {
			return err
		}
	}

	return nil
}

func execInGarage(cm *ContainerManager, c testcontainers.Container, args ...string) (string, error) {
	cmd := append([]string{"/garage"}, args...)

	code, reader, err := c.Exec(cm.ctx, cmd, tcexec.Multiplexed())
	if err != nil {
		return "", fmt.Errorf("garage %s: %w", strings.Join(args, " "), err)
	}

	out, readErr := io.ReadAll(reader)
	if readErr != nil {
		return "", fmt.Errorf("garage %s: read output: %w", strings.Join(args, " "), readErr)
	}

	if code != 0 {
		return "", fmt.Errorf("garage %s exited with %d: %s", strings.Join(args, " "), code, strings.TrimSpace(string(out)))
	}

	return string(out), nil
}

// garageNodeID extracts the node id from `garage status` output: the first
// hexadecimal token in the table below the HEALTHY NODES header.
func garageNodeID(status string) string {
	inTable := false

	for line := range strings.SplitSeq(status, "\n") {
		if strings.HasPrefix(line, "==== HEALTHY NODES") {
			inTable = true
			continue
		}

		if !inTable {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) > 1 && isGarageNodeID(fields[0]) {
			return fields[0]
		}
	}

	return ""
}

// isGarageNodeID reports whether the token looks like a Garage node id
// (16 lowercase hex characters).
func isGarageNodeID(token string) bool {
	if len(token) != 16 {
		return false
	}

	for _, r := range token {
		if !strings.ContainsRune("0123456789abcdef", r) {
			return false
		}
	}

	return true
}
