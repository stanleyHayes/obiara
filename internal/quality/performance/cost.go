package performance

import "errors"

var ErrInvalidCostInput = errors.New("invalid cost input")

type CostAssumptions struct {
	FixedComputeCents, FixedDatabaseCents                                                                        int64
	StorageGiBCents, EgressGiBCents, NotificationThousandCents, AIMinuteThousandCents, LiveSeatHourThousandCents int64
}
type CostScenario struct {
	Name                                                                string
	MAU, StorageGiB, EgressGiB, Notifications, AIMinutes, LiveSeatHours int64
}
type CostEvidence struct {
	Scenario                                                                                           string `json:"scenario"`
	MonthlyCents, PerMAUCents                                                                          int64
	ComputeCents, DatabaseCents, StorageCents, EgressCents, NotificationCents, AICents, LiveAudioCents int64
}

func Estimate(a CostAssumptions, s CostScenario) (CostEvidence, error) {
	if s.Name == "" || s.MAU <= 0 || s.StorageGiB < 0 || s.EgressGiB < 0 || s.Notifications < 0 || s.AIMinutes < 0 || s.LiveSeatHours < 0 ||
		a.FixedComputeCents < 0 || a.FixedDatabaseCents < 0 || a.StorageGiBCents < 0 || a.EgressGiBCents < 0 || a.NotificationThousandCents < 0 || a.AIMinuteThousandCents < 0 || a.LiveSeatHourThousandCents < 0 {
		return CostEvidence{}, ErrInvalidCostInput
	}
	e := CostEvidence{Scenario: s.Name, ComputeCents: a.FixedComputeCents, DatabaseCents: a.FixedDatabaseCents}
	e.StorageCents = s.StorageGiB * a.StorageGiBCents
	e.EgressCents = s.EgressGiB * a.EgressGiBCents
	e.NotificationCents = ceilThousand(s.Notifications) * a.NotificationThousandCents
	e.AICents = ceilThousand(s.AIMinutes) * a.AIMinuteThousandCents
	e.LiveAudioCents = ceilThousand(s.LiveSeatHours) * a.LiveSeatHourThousandCents
	e.MonthlyCents = e.ComputeCents + e.DatabaseCents + e.StorageCents + e.EgressCents + e.NotificationCents + e.AICents + e.LiveAudioCents
	e.PerMAUCents = (e.MonthlyCents + s.MAU - 1) / s.MAU
	return e, nil
}
func ceilThousand(v int64) int64 { return (v + 999) / 1000 }
