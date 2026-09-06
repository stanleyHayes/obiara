package config

import "testing"

func TestASchemeWithNoRateIsNotConfigured(t *testing.T) {
	// A commission with no withholding rate cannot legally pay out, and a
	// scheme that accrues but can never pay is a liability that only grows.
	for name, affiliates := range map[string]Affiliates{
		"no commission":          {WithholdingBasisPoints: 750},
		"no withholding":         {CommissionPesewas: 2000},
		"nothing at all":         {},
		"a full hundred percent": {CommissionPesewas: 2000, WithholdingBasisPoints: 10_000},
	} {
		t.Run(name, func(t *testing.T) {
			if affiliates.Configured() {
				t.Fatal("an unset scheme reported itself ready to pay people")
			}
		})
	}
	if !(Affiliates{CommissionPesewas: 2000, WithholdingBasisPoints: 750}).Configured() {
		t.Fatal("a configured scheme was refused")
	}
}

func TestATypoInARateReadsAsAbsentRatherThanAsANumber(t *testing.T) {
	// Zero is the "not set" value everywhere these are used, so a typo has to
	// land there rather than on some other number.
	for _, value := range []string{"7.5", "750%", "-750", "seven fifty", "", " "} {
		if got := wholeNumber(value); got != 0 {
			t.Fatalf("%q read as %d, want 0", value, got)
		}
	}
	if got := wholeNumber("750"); got != 750 {
		t.Fatalf("750 read as %d", got)
	}
}
