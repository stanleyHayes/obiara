package privacy

import "testing"

func TestKeyerIsScopedAndDeterministic(t *testing.T) {
	keyer, err := NewKeyer([]byte("01234567890123456789012345678901"))
	if err != nil {
		t.Fatal(err)
	}
	first, _ := keyer.Key("pace_member", "member-1")
	again, _ := keyer.Key("pace_member", "member-1")
	other, _ := keyer.Key("pace_room", "member-1")
	if first != again || first == other || first == "member-1" {
		t.Fatal("privacy key contract failed")
	}
}
