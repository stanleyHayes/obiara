// Package domain is the producer-enforced analytics schema registry
// (E15-S01; Doc 08 §3). Every event name and property is declared here;
// anything undeclared fails at the producer boundary. The registry is the
// mechanism behind the product law: no analytics event contains raw
// conversation content, voice, or free text.
package domain

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	ErrUnregisteredEvent = errors.New("analytics event is not registered")
	ErrUnknownProp       = errors.New("analytics event prop is not in the schema")
	ErrInvalidPropValue  = errors.New("analytics prop value violates its schema")
)

// PropKind constrains what a property may carry.
type PropKind int

const (
	KindNumber PropKind = iota
	KindBoolean
	// KindEnum: one of a bounded, declared value set.
	KindEnum
	// KindOpaqueID: short machine identifier, never free text.
	KindOpaqueID
)

// PropSchema declares one event property.
type PropSchema struct {
	Kind     PropKind
	Values   []string // for KindEnum
	Required bool
}

var (
	eventNamePattern = regexp.MustCompile(`^[a-z]+\.[a-z_]+$`)
	opaqueIDPattern  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,63}$`)
)

// EventSchema declares one registered event (zone.object_action).
type EventSchema struct {
	Name  string
	Props map[string]PropSchema
}

func number(required bool) PropSchema   { return PropSchema{Kind: KindNumber, Required: required} }
func boolean(required bool) PropSchema  { return PropSchema{Kind: KindBoolean, Required: required} }
func opaqueID(required bool) PropSchema { return PropSchema{Kind: KindOpaqueID, Required: required} }
func enum(required bool, values ...string) PropSchema {
	return PropSchema{Kind: KindEnum, Values: values, Required: required}
}

// Registry is the authoritative event catalog (Doc 08 §3 taxonomy).
var Registry = []EventSchema{
	{Name: "epono.pod_heard", Props: map[string]PropSchema{
		"durationPct":   number(true),
		"trustPathType": enum(true, "intro", "circle", "ember", "nnoboa", "matchmaker"),
	}},
	{Name: "epono.seed_sown", Props: map[string]PropSchema{
		"targetSource": enum(true, "intro", "circle", "ember", "nnoboa", "matchmaker"),
	}},
	{Name: "epono.sprout_opened", Props: map[string]PropSchema{
		"latencyBucket": enum(true, "minutes", "hours", "day", "days"),
		"source":        enum(true, "intro", "circle", "ember", "nnoboa", "matchmaker"),
	}},
	{Name: "epono.room_opened", Props: map[string]PropSchema{
		"latencyBucket": enum(true, "minutes", "hours", "day", "days"),
		"source":        enum(true, "intro", "circle", "ember", "nnoboa", "matchmaker"),
	}},
	{Name: "danmu.drum_passed", Props: map[string]PropSchema{
		"responseLatencyBucket": enum(true, "minutes", "hours", "day", "days"),
	}},
	{Name: "danmu.theme_completed", Props: map[string]PropSchema{
		"themeId": opaqueID(true),
	}},
	{Name: "danmu.room_closed", Props: map[string]PropSchema{
		"mode": enum(true, "kind", "expiry", "relight"),
	}},
	{Name: "gyaase.fire_attended", Props: map[string]PropSchema{
		"type":        enum(true, "weekly", "durbar", "special"),
		"minutes":     number(true),
		"gamesPlayed": number(false),
	}},
	{Name: "gyaase.ember_converted", Props: map[string]PropSchema{}},
	{Name: "ceremony.gate_crossed", Props: map[string]PropSchema{}},
	{Name: "ceremony.aseda_declared", Props: map[string]PropSchema{}},
	{Name: "wellbeing.regret_reported", Props: map[string]PropSchema{
		"surface": enum(true, "room", "doorway", "pod", "circle", "fire", "profile"),
	}},
	{Name: "commerce.order_completed", Props: map[string]PropSchema{
		"sku":    opaqueID(true),
		"market": enum(true, "gh", "diaspora"),
	}},
}

func find(name string) (EventSchema, bool) {
	for _, schema := range Registry {
		if schema.Name == name {
			return schema, true
		}
	}
	return EventSchema{}, false
}

// ValidateProps enforces the registry: the event must be registered, every
// prop must be declared with a conforming value, and every required prop
// must be present. Free text cannot validate (KindEnum and KindOpaqueID
// accept only bounded machine values).
func ValidateProps(name string, props map[string]any) error {
	schema, ok := find(name)
	if !ok || !eventNamePattern.MatchString(name) {
		return fmt.Errorf("%w: %q", ErrUnregisteredEvent, name)
	}
	for key, value := range props {
		declaration, declared := schema.Props[key]
		if !declared {
			return fmt.Errorf("%w: %q on %q", ErrUnknownProp, key, name)
		}
		if err := validateValue(declaration, value); err != nil {
			return fmt.Errorf("%w: prop %q on %q", err, key, name)
		}
	}
	for key, declaration := range schema.Props {
		if declaration.Required {
			if _, present := props[key]; !present {
				return fmt.Errorf("%w: required prop %q on %q", ErrInvalidPropValue, key, name)
			}
		}
	}
	return nil
}

func validateValue(declaration PropSchema, value any) error {
	switch declaration.Kind {
	case KindNumber:
		switch value.(type) {
		case float64, float32, int, int64, int32:
			return nil
		}
		return ErrInvalidPropValue
	case KindBoolean:
		if _, ok := value.(bool); ok {
			return nil
		}
		return ErrInvalidPropValue
	case KindEnum:
		text, ok := value.(string)
		if !ok {
			return ErrInvalidPropValue
		}
		for _, allowed := range declaration.Values {
			if text == allowed {
				return nil
			}
		}
		return ErrInvalidPropValue
	case KindOpaqueID:
		text, ok := value.(string)
		if !ok || !opaqueIDPattern.MatchString(text) {
			return ErrInvalidPropValue
		}
		return nil
	default:
		return ErrInvalidPropValue
	}
}
