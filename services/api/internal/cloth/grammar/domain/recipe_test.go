package domain

import (
	"errors"
	"fmt"
	"slices"
	"testing"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func input() Input {
	return Input{VersionV1, [2]string{key(2), key(1)}, []string{key(4), key(3)}, []string{key(6), key(5)}}
}
func TestCanonicalOrderIsDeterministic(t *testing.T) {
	a, err := Compile(input(), "command")
	if err != nil {
		t.Fatal(err)
	}
	in := input()
	slices.Reverse(in.ThemeKeys)
	slices.Reverse(in.ProvenanceKeys)
	in.PairKeys[0], in.PairKeys[1] = in.PairKeys[1], in.PairKeys[0]
	b, err := Compile(in, "command")
	if err != nil {
		t.Fatal(err)
	}
	if a.ID() != b.ID() || a.RenderSeed() != b.RenderSeed() || !slices.Equal(a.Tokens(), b.Tokens()) {
		t.Fatalf("drift %#v %#v", a, b)
	}
}
func TestRejectsUnknownDuplicateOversizeAndExecutableInput(t *testing.T) {
	tests := []Input{func() Input { x := input(); x.Version = "cloth.v2"; return x }(), func() Input { x := input(); x.ThemeKeys = []string{key(3), key(3)}; return x }(), func() Input {
		x := input()
		x.ThemeKeys = make([]string, MaxThemes+1)
		for i := range x.ThemeKeys {
			x.ThemeKeys[i] = key(i + 10)
		}
		return x
	}(), func() Input { x := input(); x.ThemeKeys = []string{"{{ execute .Payload }}"}; return x }()}
	for _, in := range tests {
		if _, err := Compile(in, "command"); err == nil {
			t.Fatalf("unsafe input accepted %#v", in)
		}
	}
	if _, err := Compile(func() Input { x := input(); x.Version = "cloth.v2"; return x }(), "command"); !errors.Is(err, ErrUnknownVersion) {
		t.Fatalf("version=%v", err)
	}
}
func FuzzCanonicalPermutationNeverDrifts(f *testing.F) {
	f.Add(uint8(7))
	f.Fuzz(func(t *testing.T, bits uint8) {
		in := input()
		if bits&1 != 0 {
			in.PairKeys[0], in.PairKeys[1] = in.PairKeys[1], in.PairKeys[0]
		}
		if bits&2 != 0 {
			slices.Reverse(in.ThemeKeys)
		}
		if bits&4 != 0 {
			slices.Reverse(in.ProvenanceKeys)
		}
		got, err := Compile(in, "command")
		if err != nil {
			t.Fatal(err)
		}
		baseline, _ := Compile(input(), "command")
		if got.RenderSeed() != baseline.RenderSeed() || !slices.Equal(got.Tokens(), baseline.Tokens()) {
			t.Fatal("canonical drift")
		}
	})
}
