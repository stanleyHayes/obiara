// Package fireacceptance provides the deterministic, synthetic Fire release
// gate for E09-S10. It does not make network calls or use member data.
package fireacceptance

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"
)

const (
	AcceptanceSeats     = 150
	ControlLatencyLimit = 400 * time.Millisecond
	listenOnlyFloorKbps = 32
)

type MediaMode string

const (
	ModeAudio      MediaMode = "audio"
	ModeListenOnly MediaMode = "listen_only"
)

type Control string

const (
	ControlSafety        Control = "safety"
	ControlLeave         Control = "leave"
	ControlConsentOptIn  Control = "consent_opt_in"
	ControlConsentRevoke Control = "consent_revoke"
)

var requiredControls = []Control{
	ControlSafety,
	ControlLeave,
	ControlConsentOptIn,
	ControlConsentRevoke,
}

type Device struct {
	Name           string
	Platform       string
	MemoryMiB      int
	ControlPenalty time.Duration
}

type Network struct {
	Name           string
	ThroughputKbps int
	RTT            time.Duration
	Jitter         time.Duration
	LossPercent    int
}

type Probe struct {
	Seat    int
	Device  Device
	Network Network
}

type Observation struct {
	Mode           MediaMode
	ControlLatency map[Control]time.Duration
}

// ProbeTransport is the one seam a real device-lab adapter can implement.
// The default acceptance suite uses DeterministicTransport and never touches
// production traffic.
//
//go:generate mockgen -source=harness.go -destination=mock_transport_test.go -package=fireacceptance
type ProbeTransport interface {
	Probe(context.Context, Probe) (Observation, error)
}

type MatrixResult struct {
	Device              string
	Seats               int
	ListenOnlySeats     int
	UnavailableControls int
	P90ControlLatency   time.Duration
}

type Report struct {
	Seats               int
	ListenOnlySeats     int
	UnavailableControls int
	P90ControlLatency   time.Duration
	Matrix              []MatrixResult
}

type Harness struct {
	transport ProbeTransport
}

func New(transport ProbeTransport) Harness {
	return Harness{transport: transport}
}

func (h Harness) Run(ctx context.Context, seats int, devices []Device, network Network) (Report, error) {
	if h.transport == nil || seats != AcceptanceSeats {
		return Report{}, errors.New("invalid fire acceptance configuration")
	}
	if err := validateMatrix(devices, network); err != nil {
		return Report{}, errors.New("invalid fire acceptance configuration")
	}

	report := Report{Seats: seats}
	matrix := make(map[string]*MatrixResult, len(devices))
	matrixLatencies := make(map[string][]time.Duration, len(devices))
	allLatencies := make([]time.Duration, 0, seats*len(requiredControls))

	for seat := range seats {
		device := devices[seat%len(devices)]
		result := matrix[device.Name]
		if result == nil {
			result = &MatrixResult{Device: device.Name}
			matrix[device.Name] = result
		}
		result.Seats++

		observation, err := h.transport.Probe(ctx, Probe{Seat: seat + 1, Device: device, Network: network})
		if err != nil {
			return Report{}, fmt.Errorf("probe seat %d: %w", seat+1, err)
		}
		if observation.Mode == ModeListenOnly {
			report.ListenOnlySeats++
			result.ListenOnlySeats++
		}
		for _, control := range requiredControls {
			latency, available := observation.ControlLatency[control]
			if !available || latency <= 0 {
				report.UnavailableControls++
				result.UnavailableControls++
				continue
			}
			allLatencies = append(allLatencies, latency)
			matrixLatencies[device.Name] = append(matrixLatencies[device.Name], latency)
		}
	}

	for _, device := range devices {
		result := matrix[device.Name]
		result.P90ControlLatency = percentile(matrixLatencies[device.Name], 90)
		report.Matrix = append(report.Matrix, *result)
	}
	report.P90ControlLatency = percentile(allLatencies, 90)

	if report.UnavailableControls != 0 {
		return report, errors.New("a safety, leave, or consent control became unavailable")
	}
	if report.P90ControlLatency > ControlLatencyLimit {
		return report, fmt.Errorf("control latency p90 %s exceeds %s", report.P90ControlLatency, ControlLatencyLimit)
	}
	if network.ThroughputKbps < listenOnlyFloorKbps && report.ListenOnlySeats != seats {
		return report, errors.New("constrained seats did not all enter listen-only mode")
	}
	return report, nil
}

func validateMatrix(devices []Device, network Network) error {
	if len(devices) < 5 || network.Name == "" || network.ThroughputKbps <= 0 || network.RTT <= 0 || network.Jitter < 0 || network.LossPercent < 0 || network.LossPercent > 100 {
		return errors.New("invalid matrix")
	}
	seen := make(map[string]struct{}, len(devices))
	for _, device := range devices {
		if device.Name == "" || device.Platform == "" || device.MemoryMiB <= 0 || device.ControlPenalty < 0 {
			return errors.New("invalid device")
		}
		if _, exists := seen[device.Name]; exists {
			return errors.New("duplicate device")
		}
		seen[device.Name] = struct{}{}
	}
	return nil
}

func percentile(values []time.Duration, p int) time.Duration {
	if len(values) == 0 {
		return 0
	}
	sorted := slices.Clone(values)
	slices.Sort(sorted)
	index := (p*len(sorted) + 99) / 100
	return sorted[index-1]
}

func deterministicLatency(seat, controlIndex int, device Device, network Network) time.Duration {
	jitterWindow := int(network.Jitter / time.Millisecond)
	jitter := 0
	if jitterWindow > 0 {
		jitter = (seat*17 + controlIndex*11) % (jitterWindow + 1)
	}
	contention := (seat*7 + controlIndex*3) % 41
	return network.RTT + device.ControlPenalty + time.Duration(jitter+contention)*time.Millisecond
}
