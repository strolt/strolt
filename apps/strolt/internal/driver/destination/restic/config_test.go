package restic

import (
	"slices"
	"testing"
)

func TestGetKeepFlagsEmptyPolicy(t *testing.T) {
	t.Parallel()

	i := &Restic{}

	if flags := i.getKeepFlags(); len(flags) != 0 {
		t.Fatalf("getKeepFlags() with empty policy = %v, want none", flags)
	}
}

func TestGetKeepFlagsSingleField(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		keep ResticConfigKeep
		want string
	}{
		{name: "last", keep: ResticConfigKeep{Last: 3}, want: "--keep-last=3"},
		{name: "hourly", keep: ResticConfigKeep{Hourly: 2}, want: "--keep-hourly=2"},
		{name: "daily", keep: ResticConfigKeep{Daily: 7}, want: "--keep-daily=7"},
		{name: "weekly", keep: ResticConfigKeep{Weekly: 4}, want: "--keep-weekly=4"},
		{name: "monthly", keep: ResticConfigKeep{Monthly: 12}, want: "--keep-monthly=12"},
		{name: "yearly", keep: ResticConfigKeep{Yearly: 5}, want: "--keep-yearly=5"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			i := &Restic{config: Config{Keep: tc.keep}}

			flags := i.getKeepFlags()
			if len(flags) != 1 || flags[0] != tc.want {
				t.Fatalf("getKeepFlags() = %v, want exactly [%s]", flags, tc.want)
			}
		})
	}
}

func TestGetKeepFlagsCombined(t *testing.T) {
	t.Parallel()

	i := &Restic{config: Config{Keep: ResticConfigKeep{
		Last:    1,
		Hourly:  2,
		Daily:   3,
		Weekly:  4,
		Monthly: 5,
		Yearly:  6,
	}}}

	flags := i.getKeepFlags()

	want := []string{
		"--keep-last=1",
		"--keep-hourly=2",
		"--keep-daily=3",
		"--keep-weekly=4",
		"--keep-monthly=5",
		"--keep-yearly=6",
	}

	if len(flags) != len(want) {
		t.Fatalf("getKeepFlags() = %v, want all of %v", flags, want)
	}

	for _, flag := range want {
		if !slices.Contains(flags, flag) {
			t.Fatalf("getKeepFlags() = %v, missing %s", flags, flag)
		}
	}
}
