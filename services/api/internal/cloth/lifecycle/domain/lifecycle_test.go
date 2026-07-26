package domain

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func fresh(t *testing.T) Lifecycle {
	v, e := New("cloth-life-01", []string{key(1), key(2)}, Provenance{1, "ref_recipeabcdefghijklmnop"})
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func TestOwnershipArchiveExportDelete(t *testing.T) {
	for _, a := range []string{key(1), key(2)} {
		v, e := fresh(t).Archive(Command{ID: "archive-command-01", ActorKey: a, ArchiveRef: "ref_archiveabcdefghijklmnop", At: time.Unix(100, 0)})
		if e != nil {
			t.Fatal(e)
		}
		if _, e = v.Export(a, "ref_archiveabcdefghijklmnop"); e != nil {
			t.Fatal(e)
		}
		if _, e = v.Export(key(9), "ref_archiveabcdefghijklmnop"); e != ErrDenied {
			t.Fatalf("%v", e)
		}
		v, e = v.Delete(Command{ID: "delete-command-01", ActorKey: a, ReceiptKey: key(8), ExpectedRevision: 1, At: time.Unix(101, 0)})
		if e != nil || v.Status() != StatusDeleted || v.ArchiveRef() != "" {
			t.Fatalf("%v %v", v.Status(), e)
		}
		if v.Tombstone().Provenance.BandVersion != 1 {
			t.Fatal("lost provenance")
		}
	}
}
func TestReplayCASAndPrivacy(t *testing.T) {
	c := Command{ID: "archive-command-01", ActorKey: key(1), ArchiveRef: "ref_archiveabcdefghijklmnop", At: time.Unix(100, 0)}
	v, _ := fresh(t).Archive(c)
	r, e := v.Archive(c)
	if e != nil || r.Revision() != 1 {
		t.Fatalf("%v", e)
	}
	c.ArchiveRef = "ref_changedabcdefghijklmnop"
	if _, e = v.Archive(c); e != ErrCommandMismatch {
		t.Fatalf("%v", e)
	}
	c = Command{ID: "archive-command-02", ActorKey: key(1), ArchiveRef: "ref_archiveabcdefghijklmnop", ExpectedRevision: 9, At: time.Now()}
	if _, e = fresh(t).Archive(c); e != ErrStaleRevision {
		t.Fatalf("%v", e)
	}
	wire := strings.ToLower(fmt.Sprintf("%+v", v))
	for _, bad := range []string{"raw content", "public", "reverse", "list"} {
		if strings.Contains(wire, bad) {
			t.Fatalf("%q", bad)
		}
	}
}
func TestExactlyTwoOpaqueMembers(t *testing.T) {
	for _, m := range [][]string{{key(1)}, {key(1), key(2), key(3)}, {key(1), "raw"}} {
		if _, e := New("cloth-life-01", m, Provenance{1, "ref_recipeabcdefghijklmnop"}); e == nil {
			t.Fatal(m)
		}
	}
}
