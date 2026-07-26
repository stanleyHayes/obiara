package flags

import "time"

// AuditKind is a safe, bounded event category.
type AuditKind string

const (
	AuditEvaluation     AuditKind = "evaluation"
	AuditConfiguration  AuditKind = "configuration_change"
	AuditRejectedChange AuditKind = "configuration_rejected"
)

// ChangedField lists configuration fields without carrying raw input.
type ChangedField string

const (
	ChangedEnabled    ChangedField = "enabled"
	ChangedKillSwitch ChangedField = "kill_switch"
)

// RejectionCode is deliberately finite so audit output cannot echo malformed
// configuration, user identifiers, secrets, or free-form reasons.
type RejectionCode string

const (
	RejectUnknownFlag RejectionCode = "unknown_flag"
	RejectEmptyChange RejectionCode = "empty_change"
)

// AuditRecord contains only canonical, low-cardinality configuration metadata.
// It intentionally has no actor, subject, request payload, reason, or raw
// environment value fields.
type AuditRecord struct {
	At            time.Time
	Kind          AuditKind
	Flag          Flag
	Known         bool
	Enabled       bool
	Killed        bool
	Source        Source
	Version       uint64
	ChangedFields []ChangedField
	Rejection     RejectionCode
}

// AuditSink receives serialized audit records. Registry serializes calls, so a
// sink need not provide its own concurrency control.
type AuditSink interface {
	Record(AuditRecord)
}

// DiscardAudit is the explicit no-op sink.
type DiscardAudit struct{}

func (DiscardAudit) Record(AuditRecord) {}

func evaluationRecord(at time.Time, decision Decision) AuditRecord {
	flag := decision.Flag
	if !decision.Known {
		// Unknown input is intentionally not copied into audit storage because
		// an untrusted value could itself contain personal data or a secret.
		flag = ""
	}
	return AuditRecord{
		At:      at,
		Kind:    AuditEvaluation,
		Flag:    flag,
		Known:   decision.Known,
		Enabled: decision.Enabled,
		Killed:  decision.Killed,
		Source:  decision.Source,
		Version: decision.Version,
	}
}

func changeRecord(at time.Time, next Decision, change Change) AuditRecord {
	record := AuditRecord{
		At:      at,
		Kind:    AuditConfiguration,
		Flag:    change.Flag,
		Known:   true,
		Enabled: next.Enabled,
		Killed:  next.Killed,
		Source:  next.Source,
		Version: next.Version,
	}
	if change.Enabled != nil {
		record.ChangedFields = append(record.ChangedFields, ChangedEnabled)
	}
	if change.Killed != nil {
		record.ChangedFields = append(record.ChangedFields, ChangedKillSwitch)
	}
	return record
}

func rejectedRecord(at time.Time, flag Flag, rejection RejectionCode) AuditRecord {
	known := IsCanonical(flag)
	if !known {
		// Retain only the bounded rejection code for untrusted flag names.
		flag = ""
	}
	return AuditRecord{
		At:        at,
		Kind:      AuditRejectedChange,
		Flag:      flag,
		Known:     known,
		Source:    SourceUnknown,
		Rejection: rejection,
	}
}
