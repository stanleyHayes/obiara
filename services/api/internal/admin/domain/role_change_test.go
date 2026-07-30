package domain

import (
	"testing"
	"time"
)

func TestRoleChangeRequiresDistinctApprover(t *testing.T) {
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	change, err := NewRoleChange("chg_1", "adm_target", 3, []Role{RoleAdmin}, "Grant command-centre coverage", "adm_first", now)
	if err != nil {
		t.Fatal(err)
	}
	if err := change.Approve("adm_first", now.Add(time.Minute)); err != ErrSameApprover {
		t.Fatalf("Approve by proposer = %v, want ErrSameApprover", err)
	}
	if err := change.Approve("adm_second", now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if change.Status() != RoleChangeApproved || change.ApproverID() != "adm_second" {
		t.Fatalf("approved change = %#v", change)
	}
	if err := change.Approve("adm_third", now.Add(2*time.Minute)); err != ErrRoleChangeClosed {
		t.Fatalf("second approval = %v, want ErrRoleChangeClosed", err)
	}
}

func TestRoleChangeRejectsUnboundedReason(t *testing.T) {
	_, err := NewRoleChange("chg_1", "adm_target", 1, []Role{RoleAdmin}, "too short", "adm_first", time.Now())
	if err != ErrRoleChangeReason {
		t.Fatalf("NewRoleChange = %v, want ErrRoleChangeReason", err)
	}
}
