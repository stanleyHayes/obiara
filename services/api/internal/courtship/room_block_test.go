package courtship

import (
	"context"
	"errors"
	"testing"
)

type blockStub struct {
	blocked bool
	err     error
	asked   [][2]string
}

func (stub *blockStub) Blocked(_ context.Context, memberID, otherID string) (bool, error) {
	stub.asked = append(stub.asked, [2]string{memberID, otherID})
	return stub.blocked, stub.err
}

func TestARoomIsNotOpenedAcrossABlock(t *testing.T) {
	// A room is a private conversation between two people. Opening one
	// across a block is the plainest possible violation of what a block is
	// for, and nothing stopped it.
	//
	// The module is deliberately zero-valued: if the check does not refuse
	// first, Start reaches a nil Pace service and panics rather than
	// quietly passing, so this test cannot succeed for the wrong reason.
	blocks := &blockStub{blocked: true}
	room := NewRoom(RoomModule{}).WithBlocks(blocks)

	if _, err := room.Start(context.Background(), "room-1", []string{"alice", "bob"}, "cmd-1", "alice"); !errors.Is(err, ErrBlocked) {
		t.Fatalf("err = %v, want ErrBlocked", err)
	}
	if len(blocks.asked) == 0 {
		t.Fatal("the block list was never consulted")
	}
	// The actor is not asked about themselves.
	for _, pair := range blocks.asked {
		if pair[0] == pair[1] {
			t.Fatalf("the block list was asked about %q and themselves", pair[0])
		}
	}
}

func TestARoomWithNoBlockCheckIsNotOpenedAtAll(t *testing.T) {
	// Five aggregates are opened below the check. A composition that forgot
	// the block check must not open a room blind to blocks, so the absence
	// of the check closes the surface rather than opening it.
	room := NewRoom(RoomModule{})
	if _, err := room.Start(context.Background(), "room-1", []string{"alice", "bob"}, "cmd-1", "alice"); !errors.Is(err, ErrBlocked) {
		t.Fatalf("err = %v, want ErrBlocked", err)
	}
}

func TestAnUnreadableBlockListDoesNotOpenARoom(t *testing.T) {
	blocks := &blockStub{err: errors.New("safety unavailable")}
	room := NewRoom(RoomModule{}).WithBlocks(blocks)
	if _, err := room.Start(context.Background(), "room-1", []string{"alice", "bob"}, "cmd-1", "alice"); !errors.Is(err, ErrBlocked) {
		t.Fatalf("err = %v, want ErrBlocked", err)
	}
}
