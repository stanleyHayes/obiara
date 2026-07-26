package fireacceptance

import (
	"context"
	"time"
)

// DeterministicTransport is a reproducible load model. It contains no
// credentials, identifiers, clock reads, randomness, sockets, or production
// endpoints.
type DeterministicTransport struct{}

func (DeterministicTransport) Probe(_ context.Context, probe Probe) (Observation, error) {
	mode := ModeAudio
	if probe.Network.ThroughputKbps < listenOnlyFloorKbps {
		mode = ModeListenOnly
	}
	latencies := make(map[Control]time.Duration, len(requiredControls))
	for index, control := range requiredControls {
		latencies[control] = deterministicLatency(probe.Seat, index, probe.Device, probe.Network)
	}
	return Observation{Mode: mode, ControlLatency: latencies}, nil
}

// ReferenceDevices is the representative automated matrix. The physical
// release run must still validate these classes on real hardware.
func ReferenceDevices() []Device {
	return []Device{
		{Name: "android-api26-2gb", Platform: "android", MemoryMiB: 2048, ControlPenalty: 62 * time.Millisecond},
		{Name: "android-current-budget", Platform: "android", MemoryMiB: 3072, ControlPenalty: 48 * time.Millisecond},
		{Name: "android-current-mid", Platform: "android", MemoryMiB: 6144, ControlPenalty: 28 * time.Millisecond},
		{Name: "ios-se2", Platform: "ios", MemoryMiB: 3072, ControlPenalty: 34 * time.Millisecond},
		{Name: "web-entry-laptop", Platform: "web", MemoryMiB: 4096, ControlPenalty: 42 * time.Millisecond},
	}
}

func Constrained3G() Network {
	return Network{
		Name:           "constrained-3g",
		ThroughputKbps: 24,
		RTT:            180 * time.Millisecond,
		Jitter:         70 * time.Millisecond,
		LossPercent:    5,
	}
}
