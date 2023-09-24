package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	tt := timeTook("docker compose down")

	assert.NoError(t, dockerComposeDown())
	tt.stop()

	tt = timeTook("docker compose up")

	assert.NoError(t, dockerComposeUp("minio", "postgres", "mongo", "mariadb", "mysql"))
	tt.stop()

	tt = timeTook("strolt up")

	assert.NoError(t, dockerComposeUpStrolt())
	tt.stop()

	tt = timeTook("strolt init")

	assert.NoError(t, strolt("init"))
	tt.stop()

	LocalSuiteTest(t)
	PruneSuiteTest(t)

	PostgresqlSuiteTest(t)
	MongoSuiteTest(t)
	MySQLSuiteTest(t)
	MariaDBSuiteTest(t)

	tt = timeTook("docker compose down")

	assert.NoError(t, dockerComposeDown())
	tt.stop()
}
