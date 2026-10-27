package e2e_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	redisStringKey   = "e2e:string"
	redisStringValue = "redis-backup-payload-🚀-данные"
	redisHashKey     = "e2e:hash"
	redisHashField   = "f1"
	redisHashValue   = "v1"
)

// RedisSuite proves the redis source driver end to end against Redis and
// Valkey. The driver cannot restore (an RDB only loads through the server's
// data directory), so the round trip dumps the artifact out of the restic
// repository, loads it into a fresh Redis instance and checks the data; the
// valkey subtest doubles as a cross-implementation compatibility proof.
type RedisSuite struct {
	suite.Suite
}

func (s *RedisSuite) BeforeTest(_, _ string) {
	// Valkey runs with requirepass, so its data ops go through the authed
	// helper; the plain Redis container needs no password.
	ops := [][]string{
		{"SET", redisStringKey, redisStringValue},
		{"HSET", redisHashKey, redisHashField, redisHashValue},
	}

	for _, op := range ops {
		_, err := cacheExec(containerManager.redisContainer, append([]string{"redis-cli"}, op...)...)
		s.Require().NoError(err)

		_, err = cacheExecAuthed(containerManager.valkeyContainer, op...)
		s.Require().NoError(err)
	}
}

func (s *RedisSuite) TestRedis() {
	s.roundTrip("e2e", containerManager.redisContainer, "s3:http://minio:9000/restic-redis")
}

func (s *RedisSuite) TestRedis_pipe() {
	s.roundTrip("e2e-pipe", containerManager.redisContainer, "s3:http://minio:9000/pipe/restic-redis")
}

func (s *RedisSuite) TestRedis_valkey() {
	s.roundTrip("e2e-valkey", containerManager.valkeyContainer, "s3:http://minio:9000/restic-redis-valkey")
}

func (s *RedisSuite) roundTrip(serviceName string, source testcontainers.Container, repo string) {
	s.T().Helper()

	s.Require().NotNil(source)

	s.Require().NoError(strolt("backup", "--service", serviceName, "--task", serviceName, "--y"))

	// Stderr is dropped inside the container: restic logs there, and the
	// exec stream multiplexes it with stdout, which would corrupt the
	// binary payload.
	dump, err := resticExec(repo, "dump latest strolt_driver_redis.rdb 2>/dev/null")
	s.Require().NoError(err)

	s.Require().Greater(len(dump), 100)
	s.Require().Equal("REDIS", string(dump[:5]), "artifact must be an RDB file")

	s.Equal(redisStringValue, s.loadIntoFreshRedis(serviceName, dump))
}

// loadIntoFreshRedis writes the dumped RDB into a throwaway Redis container
// as /data/dump.rdb (loaded on boot) and returns the restored string value.
// The file is copied in rather than bind-mounted: the image entrypoint chowns
// /data contents at startup, so a read-only mount aborts the container.
func (s *RedisSuite) loadIntoFreshRedis(name string, rdb []byte) string {
	s.T().Helper()

	hostPath, err := filepath.Abs(filepath.Join(".temp", "dump-"+name+".rdb"))
	s.Require().NoError(err)
	s.Require().NoError(os.WriteFile(hostPath, rdb, 0o644))

	verifyC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Image: redisImage,
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      hostPath,
				ContainerFilePath: "/data/dump.rdb",
				FileMode:          0o644,
			},
		},
		WaitingFor: wait.ForLog("Ready to accept connections").WithOccurrence(1),
		Started:    true,
	})
	s.Require().NoError(err)

	defer func() { _ = verifyC.Terminate(context.Background()) }()

	hashValue, err := cacheExec(verifyC, "redis-cli", "--raw", "HGET", redisHashKey, redisHashField)
	s.Require().NoError(err)
	s.Equal(redisHashValue, hashValue, "the RDB must preserve non-string types")

	value, err := cacheExec(verifyC, "redis-cli", "--raw", "GET", redisStringKey)
	s.Require().NoError(err)

	return value
}

//nolint:thelper
func RedisSuiteTest(t *testing.T) {
	tt := timeTook("RedisSuiteTest")

	suite.Run(t, new(RedisSuite))
	tt.stop()
}

// cacheExec runs a command inside a cache container and returns its trimmed
// output. All data checks go through it: the host has no redis client and
// the containers ship one.
func cacheExec(c testcontainers.Container, args ...string) (string, error) {
	code, reader, err := c.Exec(ctx, args, tcexec.Multiplexed())
	if err != nil {
		return "", fmt.Errorf("%s: %w", strings.Join(args, " "), err)
	}

	out, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("%s: read output: %w", strings.Join(args, " "), err)
	}

	if code != 0 {
		return "", fmt.Errorf("%s exited with code %d: %s", strings.Join(args, " "), code, out)
	}

	return strings.TrimSpace(string(out)), nil
}

// cacheExecAuthed runs a redis-cli command with REDISCLI_AUTH set, for the
// requirepass Valkey instance (the password reaches the process through the
// environment, mirroring the driver's own delivery).
func cacheExecAuthed(c testcontainers.Container, args ...string) (string, error) {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = "'" + arg + "'"
	}

	return cacheExec(c, "/bin/sh", "-c", "REDISCLI_AUTH="+redisPassword+" exec redis-cli "+strings.Join(quoted, " "))
}
