package sponsorship

import (
	"context"
	"testing"
	"time"

	ledgerapplication "github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/application"
	ledgerdomain "github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/domain"
)

type posted struct {
	commands []ledgerapplication.PostCommand
}

func (p *posted) Post(
	_ context.Context, command ledgerapplication.PostCommand,
) (ledgerdomain.Posting, error) {
	p.commands = append(p.commands, command)
	return ledgerdomain.Posting{}, nil
}

func balanced(t *testing.T, command ledgerapplication.PostCommand, want int64) {
	t.Helper()
	var debits, credits int64
	for _, line := range command.Lines {
		switch line.Side {
		case ledgerdomain.SideDebit:
			debits += line.Minor
		case ledgerdomain.SideCredit:
			credits += line.Minor
		}
	}
	if debits != credits || debits != want {
		t.Fatalf("debits %d credits %d, want %d each", debits, credits, want)
	}
}

func TestADepositIsHeldAsALiabilityAndNotAsRevenue(t *testing.T) {
	// The platform is holding somebody else's money against seats nobody has
	// taken. Booking it as revenue would recognise income for something not
	// yet delivered, and leave nothing to recognise when the seat was taken.
	poster := &posted{}
	if err := NewBook(poster, "system").RecordDeposit(
		context.Background(), "fund_1:cmd_2", 100_000, time.Now(),
	); err != nil {
		t.Fatal(err)
	}
	command := poster.commands[0]
	balanced(t, command, 100_000)
	var sawLiability, sawRevenue bool
	for _, line := range command.Lines {
		if line.Class == ledgerdomain.ClassLiability && line.Side == ledgerdomain.SideCredit {
			sawLiability = true
		}
		if line.Class == ledgerdomain.ClassRevenue {
			sawRevenue = true
		}
	}
	if !sawLiability {
		t.Fatal("a deposit did not create a liability")
	}
	if sawRevenue {
		t.Fatal("a deposit was booked as revenue before a seat was taken")
	}
}

func TestADrawTurnsTheLiabilityIntoRevenue(t *testing.T) {
	// This is the moment a seat is actually given, which is when it is earned.
	poster := &posted{}
	if err := NewBook(poster, "system").RecordDraw(
		context.Background(), "seat:cmd_1", 5_000, time.Now(),
	); err != nil {
		t.Fatal(err)
	}
	command := poster.commands[0]
	balanced(t, command, 5_000)
	var liabilityDown, revenueUp bool
	for _, line := range command.Lines {
		if line.Class == ledgerdomain.ClassLiability && line.Side == ledgerdomain.SideDebit {
			liabilityDown = true
		}
		if line.Class == ledgerdomain.ClassRevenue && line.Side == ledgerdomain.SideCredit {
			revenueUp = true
		}
	}
	if !liabilityDown || !revenueUp {
		t.Fatalf("a draw did not move the liability into revenue: %#v", command.Lines)
	}
	// No cash moves here: the money arrived at deposit.
	for _, line := range command.Lines {
		if line.Class == ledgerdomain.ClassAsset {
			t.Fatal("a draw moved cash that had already arrived")
		}
	}
}

func TestADepositAndItsDrawArePostedUnderDifferentCommands(t *testing.T) {
	// Sharing a command id would make the second posting look like a replay
	// of the first, and one of the two would silently not land.
	poster := &posted{}
	book := NewBook(poster, "system")
	_ = book.RecordDeposit(context.Background(), "same", 5_000, time.Now())
	_ = book.RecordDraw(context.Background(), "same", 5_000, time.Now())
	if poster.commands[0].CommandID == poster.commands[1].CommandID {
		t.Fatalf("both posted under %q", poster.commands[0].CommandID)
	}
}

func TestNothingIsPostedForNothing(t *testing.T) {
	poster := &posted{}
	book := NewBook(poster, "system")
	_ = book.RecordDeposit(context.Background(), "fund_1", 0, time.Now())
	_ = book.RecordDraw(context.Background(), "seat_1", -1, time.Now())
	if len(poster.commands) != 0 {
		t.Fatalf("%d postings for nothing", len(poster.commands))
	}
}
