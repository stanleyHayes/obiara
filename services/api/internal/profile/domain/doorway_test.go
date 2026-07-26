package domain

import (
	"strings"
	"testing"
	"time"
)

var doorwayNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func TestDoorwayQuestionValidation(t *testing.T) {
	if _, err := NewDoorwayQuestion("m-1", "  What does home mean to you?  ", true, doorwayNow); err != nil {
		t.Fatalf("valid question = %v", err)
	}
	if _, err := NewDoorwayQuestion("m-1", strings.Repeat("a", DoorwayQuestionLimit+1), true, doorwayNow); err != ErrDoorwayQuestionInvalid {
		t.Fatalf("too long = %v", err)
	}
	if _, err := NewDoorwayQuestion("m-1", "   ", true, doorwayNow); err != ErrDoorwayQuestionInvalid {
		t.Fatalf("blank = %v", err)
	}
	if _, err := NewDoorwayQuestion("m-1", "call me at +233 55 000 0101", true, doorwayNow); err != ErrUnsafeProfile {
		t.Fatalf("phone leak = %v, want ErrUnsafeProfile", err)
	}
	if _, err := NewDoorwayQuestion("m-1", "find me at scam@example.com", true, doorwayNow); err != ErrUnsafeProfile {
		t.Fatalf("email leak = %v, want ErrUnsafeProfile", err)
	}
	question, err := NewDoorwayQuestion("m-1", "What does family mean to you?", false, doorwayNow)
	if err != nil || question.Custom() {
		t.Fatalf("suggested question = %#v, %v", question, err)
	}
}

func TestVaultItemValidation(t *testing.T) {
	if _, err := NewVaultItem("vi_1", "m-1", "asset-1", 0, doorwayNow); err != nil {
		t.Fatal(err)
	}
	for _, position := range []int{-1, VaultCapacity, VaultCapacity + 5} {
		if _, err := NewVaultItem("vi_1", "m-1", "asset-1", position, doorwayNow); err != ErrVaultItemInvalid {
			t.Fatalf("position %d = %v", position, err)
		}
	}
	if _, err := NewVaultItem("vi_1", "m-1", "", 0, doorwayNow); err != ErrVaultItemInvalid {
		t.Fatalf("missing asset = %v", err)
	}
}

func TestVeilPolicy(t *testing.T) {
	item, _ := NewVaultItem("vi_1", "m-1", "asset-1", 0, doorwayNow)

	if view := Veil(item, "m-1"); view.Veiled {
		t.Fatal("owner must see their own photos clear")
	}
	for _, viewer := range []string{"m-2", "m-3", ""} {
		if view := Veil(item, viewer); !view.Veiled {
			t.Fatalf("viewer %q must see the veil until acceptance exists", viewer)
		}
	}
}
