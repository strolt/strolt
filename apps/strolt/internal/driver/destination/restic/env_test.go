package restic

import (
	"slices"
	"testing"
)

func TestGetEnvPairsWithoutExtra(t *testing.T) {
	t.Parallel()

	i := &Restic{env: Env{RESTIC_REPOSITORY: "s3:http://minio:9000/repo", RESTIC_PASSWORD: "secret"}}

	want := []string{"RESTIC_REPOSITORY=s3:http://minio:9000/repo", "RESTIC_PASSWORD=secret"}

	if got := i.getEnvPairs(); !slices.Equal(got, want) {
		t.Fatalf("getEnvPairs() = %v, want %v", got, want)
	}
}

// env_extra is appended after the typed fields, in a stable order: os/exec
// keeps the last occurrence of a duplicated variable, so a value set there
// overrides the typed one.
func TestGetEnvPairsExtraOverridesTypedFields(t *testing.T) {
	t.Parallel()

	i := &Restic{
		env: Env{RESTIC_REPOSITORY: "s3:http://minio:9000/repo"},
		config: Config{ExtraEnv: map[string]string{
			"RESTIC_REPOSITORY":  "rest:http://rest:8000/repo",
			"AWS_DEFAULT_REGION": "eu-central-1",
		}},
	}

	want := []string{
		"RESTIC_REPOSITORY=s3:http://minio:9000/repo",
		"AWS_DEFAULT_REGION=eu-central-1",
		"RESTIC_REPOSITORY=rest:http://rest:8000/repo",
	}

	if got := i.getEnvPairs(); !slices.Equal(got, want) {
		t.Fatalf("getEnvPairs() = %v, want %v", got, want)
	}
}

func TestGetEnvPairsSkipsEmptyValues(t *testing.T) {
	t.Parallel()

	i := &Restic{env: Env{RESTIC_REPOSITORY: "s3:http://minio:9000/repo"}}

	for _, pair := range i.getEnvPairs() {
		if pair == "RESTIC_PASSWORD=" {
			t.Fatalf("getEnvPairs() = %v, want no empty RESTIC_PASSWORD", i.getEnvPairs())
		}
	}
}
