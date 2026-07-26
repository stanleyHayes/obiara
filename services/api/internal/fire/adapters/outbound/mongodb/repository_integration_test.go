//go:build integration

package mongodb_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/application"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/domain"
)

const fireIntegrationTimeout = 3 * time.Minute

func TestFireAttendanceEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), fireIntegrationTimeout)
	t.Cleanup(cancel)

	container, err := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	if err != nil {
		t.Fatalf("start MongoDB Testcontainer (Docker/container runtime required): %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Errorf("terminate MongoDB Testcontainer: %v", err)
		}
	})

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("read Testcontainer connection string: %v", err)
	}
	separator := "?"
	if strings.Contains(uri, "?") {
		separator = "&"
	}
	uri += separator + "directConnection=true"
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatalf("connect via platform helper: %v", err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	database := client.Database("obiara_fire_test")
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatalf("ensure indexes: %v", err)
	}
	service := application.NewFireService(repository, time.Now,
		func() string { return "fire_test" })

	fire, err := service.Schedule(ctx, "host-1", "circle-1", "Friday Fire", time.Now().Add(6*time.Hour), 5)
	if err != nil {
		t.Fatalf("schedule: %v", err)
	}

	// Tier gate (FR-401).
	if _, err := service.RSVP(ctx, fire.ID(), "m-tier0", 0); err != domain.ErrTierTooLow {
		t.Fatalf("tier 0 = %v, want ErrTierTooLow", err)
	}

	// Capacity race: 20 members race for 5 seats.
	const racers = 20
	var wg sync.WaitGroup
	statuses := make([]domain.RSVPStatus, racers)
	for i := 0; i < racers; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			rsvp, err := service.RSVP(ctx, fire.ID(), fmt.Sprintf("m-%d", index), 1)
			if err != nil {
				t.Errorf("rsvp m-%d: %v", index, err)
				return
			}
			statuses[index] = rsvp.Status()
		}(i)
	}
	wg.Wait()

	going, waitlisted := 0, 0
	for _, status := range statuses {
		switch status {
		case domain.RSVPGoing:
			going++
		case domain.RSVPWaitlisted:
			waitlisted++
		}
	}
	if going != 5 || waitlisted != 15 {
		t.Fatalf("going=%d waitlisted=%d, want 5/15 after capacity race", going, waitlisted)
	}
	loaded, err := repository.FindByID(ctx, fire.ID())
	if err != nil || loaded.GoingCount() != 5 {
		t.Fatalf("fire counter = %d, %v", loaded.GoingCount(), err)
	}

	// Duplicate RSVP is rejected.
	if _, err := service.RSVP(ctx, fire.ID(), "m-0", 1); !errors.Is(err, application.ErrAlreadyRSVPed) {
		t.Fatalf("duplicate = %v, want ErrAlreadyRSVPed", err)
	}

	// Cancelling a going seat promotes the member holding waitlist position 1.
	var waitlistHead struct {
		MemberID string `bson:"memberId"`
	}
	if err := database.Collection("fire_attendance").FindOne(ctx,
		map[string]any{"fireId": fire.ID(), "status": "waitlisted", "position": 1}).Decode(&waitlistHead); err != nil {
		t.Fatalf("read waitlist head: %v", err)
	}
	var goingID string
	for i, status := range statuses {
		if status == domain.RSVPGoing {
			goingID = fmt.Sprintf("m-%d", i)
			break
		}
	}
	promoted, err := service.Cancel(ctx, fire.ID(), goingID)
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if promoted == nil || promoted.MemberID() != waitlistHead.MemberID || promoted.Status() != domain.RSVPGoing {
		t.Fatalf("promoted = %#v, want %s going", promoted, waitlistHead.MemberID)
	}
	loaded, _ = repository.FindByID(ctx, fire.ID())
	if loaded.GoingCount() != 5 {
		t.Fatalf("after promotion going = %d, want 5", loaded.GoingCount())
	}

	// Upcoming listing includes the fire.
	upcoming, err := service.Upcoming(ctx, 10)
	if err != nil || len(upcoming) == 0 {
		t.Fatalf("upcoming = %#v, %v", upcoming, err)
	}
}
