package e2e_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

const initRetries = 3

var (
	containerManager *ContainerManager
	ctx              context.Context
)

// It sets up containers once and reuses them across all tests.
func TestMain(m *testing.M) {
	if err := setupContainers(); err != nil {
		if cleanupErr := cleanupContainers(); cleanupErr != nil {
			log.Printf("Failed to cleanup containers after setup failure: %v", cleanupErr)
		}

		log.Fatalf("Failed to setup containers: %v", err)
	}

	exitCode := m.Run()

	if err := cleanupContainers(); err != nil {
		log.Printf("Failed to cleanup containers: %v", err)
	}

	os.Exit(exitCode)
}

func setupContainers() error {
	ctx = context.Background()

	// Start from a clean slate: leftovers from a previous run are a source
	// of false test results.
	if err := os.RemoveAll(".temp"); err != nil {
		return fmt.Errorf("failed to clean .temp: %w", err)
	}

	if err := os.MkdirAll(".temp/input", 0o755); err != nil {
		return fmt.Errorf("failed to create .temp/input: %w", err)
	}

	tt := timeTook("setup containers")

	cm, err := NewContainerManager(ctx)
	if err != nil {
		return err
	}

	containerManager = cm

	if err := cm.SetupNetwork(); err != nil {
		return err
	}

	if err := cm.StartAllContainers(); err != nil {
		return err
	}

	// Start Strolt container after databases are ready
	if err := cm.StartStrolt(); err != nil {
		return err
	}

	tt.stop()

	tt = timeTook("strolt init")
	defer tt.stop()

	log.Println("Initializing strolt repositories...")

	if err := stroltInit(); err != nil {
		return err
	}

	log.Println("Strolt initialization successful")

	return nil
}

// stroltInit initializes the restic repositories, retrying to absorb
// transient object-storage hiccups right after MinIO startup.
func stroltInit() error {
	var err error

	for attempt := 1; attempt <= initRetries; attempt++ {
		if err = strolt("init"); err == nil {
			return nil
		}

		log.Printf("strolt init attempt %d/%d failed: %v", attempt, initRetries, err)
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("strolt init failed after %d attempts: %w", initRetries, err)
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

	t.Run("MariaDB", func(t *testing.T) {
		MariaDBSuiteTest(t)
	})
}
