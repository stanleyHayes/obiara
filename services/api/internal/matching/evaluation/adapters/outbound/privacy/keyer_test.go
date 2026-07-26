package privacy

import "testing"

func TestHMACReviewerKey(t *testing.T) {
	k, e := New([]byte("01234567890123456789012345678901"))
	if e != nil {
		t.Fatal(e)
	}
	v, _ := k.Key("reviewer", "human@example.invalid")
	if len(v) != 64 || v == "human@example.invalid" {
		t.Fatal("reviewer identity leaked")
	}
}
