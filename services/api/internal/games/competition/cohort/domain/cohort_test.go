package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

func member(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func command(id string, revision uint64) Command {
	return Command{ID: id, ExpectedRevision: revision, At: time.Unix(int64(revision+1), 0)}
}

func TestCohortRequiresExplicitPowerOfTwoOptInBeforeStart(t *testing.T) {
	t.Parallel()
	cohort, err := Create("cohort-one", 4, command("create", 0))
	if err != nil {
		t.Fatal(err)
	}
	for index, id := range []string{"a", "b", "c", "d"} {
		cohort, err = cohort.Join(member(id), command("join-"+id, uint64(index+1)))
		if err != nil {
			t.Fatal(err)
		}
	}
	if cohort.Status() != StatusLocked {
		t.Fatalf("status = %s", cohort.Status())
	}
	cohort, err = cohort.Start("competition-one", command("start", 5))
	if err != nil || cohort.Status() != StatusStarted {
		t.Fatalf("start = %#v, %v", cohort, err)
	}
	if _, err = cohort.Leave(member("a"), command("leave", 6)); !errorsIs(err, ErrTransition) {
		t.Fatalf("started cohort allowed leave: %v", err)
	}
}

func errorsIs(actual, target error) bool { return actual == target }
