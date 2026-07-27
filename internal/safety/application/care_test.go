package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/internal/safety/domain"
)

var careSvcNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func TestFlagClosureSetsQuietening(t *testing.T) {
	ctrl := gomock.NewController(t)
	cases := NewMockCareRepository(ctrl)
	quietening := NewMockQuieteningStore(ctrl)
	service := NewCareService(cases, quietening, func() time.Time { return careSvcNow }, func() string { return "care_test" })

	cases.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	quietening.EXPECT().Set(gomock.Any(), "m-1", careSvcNow.Add(domain.QuieteningWindow)).Return(nil)

	careCase, err := service.Flag(context.Background(), "m-1", domain.SignalClosure)
	if err != nil || careCase.ID() != "care_test" {
		t.Fatalf("flag = %#v, %v", careCase, err)
	}
}

func TestFlagDistressSetsNoQuietening(t *testing.T) {
	ctrl := gomock.NewController(t)
	cases := NewMockCareRepository(ctrl)
	quietening := NewMockQuieteningStore(ctrl)
	service := NewCareService(cases, quietening, func() time.Time { return careSvcNow }, func() string { return "care_test" })
	// No Set expectation: distress flags route without quietening.
	cases.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	if _, err := service.Flag(context.Background(), "m-1", domain.SignalDistressReport); err != nil {
		t.Fatal(err)
	}
}

func TestEngageResolveFlow(t *testing.T) {
	ctrl := gomock.NewController(t)
	cases := NewMockCareRepository(ctrl)
	service := NewCareService(cases, nil, func() time.Time { return careSvcNow }, func() string { return "care_test" })

	open := domain.ReconstituteCareCase("c-1", "m-1", domain.SignalVictimReport, domain.CareOpen, nil, 1, careSvcNow, nil)
	cases.EXPECT().FindByID(gomock.Any(), "c-1").Return(open, nil)
	cases.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	if _, err := service.Engage(context.Background(), "c-1"); err != nil {
		t.Fatal(err)
	}

	engaged := domain.ReconstituteCareCase("c-1", "m-1", domain.SignalVictimReport, domain.CareEngaged, nil, 2, careSvcNow, nil)
	cases.EXPECT().FindByID(gomock.Any(), "c-1").Return(engaged, nil)
	cases.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, careCase domain.CareCase) error {
			if careCase.Status() != domain.CareResolved || len(careCase.Scripts()) != 1 {
				t.Fatalf("case = %#v", careCase)
			}
			return nil
		})
	if _, err := service.Resolve(context.Background(), "c-1", []domain.ScriptKey{domain.ScriptHelplineDirectory}); err != nil {
		t.Fatal(err)
	}
}
