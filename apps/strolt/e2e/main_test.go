package e2e_test

import (
	"context"
	"log"
	"os"
	"testing"
)

var (
	containerManager *ContainerManager
	ctx              context.Context
)

// It sets up containers once and reuses them across all tests.
func TestMain(m *testing.M) {
	var exitCode int

	// Setup
	if err := setupContainers(); err != nil {
		log.Fatalf("Failed to setup containers: %v", err)
	}

	// Run tests
	exitCode = m.Run()

	// Cleanup
	if err := cleanupContainers(); err != nil {
		log.Printf("Failed to cleanup containers: %v", err)
	}

	os.Exit(exitCode)
}

func setupContainers() error {
	ctx = context.Background()

	tt := timeTook("setup containers")

	// Initialize container manager
	cm, err := NewContainerManager(ctx)
	if err != nil {
		return err
	}

	containerManager = cm

	// Setup network
	if err := cm.SetupNetwork(); err != nil {
		return err
	}

	// Start all containers in parallel
	if err := cm.StartAllContainers(); err != nil {
		return err
	}

	// Start Strolt container after databases are ready
	if err := cm.StartStrolt(); err != nil {
		return err
	}

	tt.stop()

	tt = timeTook("strolt init")

	if err := strolt("init"); err != nil {
		return err
	}

	tt.stop()

	return nil
}

func cleanupContainers() error {
	if containerManager == nil {
		return nil
	}

	tt := timeTook("cleanup containers")
	defer tt.stop()

	return containerManager.Cleanup()
}

// TestE2E runs all e2e test suites.
func TestE2E(t *testing.T) {
	t.Run("Local", func(t *testing.T) {
		LocalSuiteTest(t)
	})

	t.Run("Prune", func(t *testing.T) {
		PruneSuiteTest(t)
	})

	t.Run("PostgreSQL", func(t *testing.T) {
		PostgresqlSuiteTest(t)
	})

	t.Run("MongoDB", func(t *testing.T) {
		MongoSuiteTest(t)
	})

	t.Run("MySQL", func(t *testing.T) {
		MySQLSuiteTest(t)
	})

	t.Run("MariaDB", func(t *testing.T) {
		MariaDBSuiteTest(t)
	})
}
