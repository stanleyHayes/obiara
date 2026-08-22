package domain

import "errors"

// ErrTemplateUnknown reports a push template with no copy.
var ErrTemplateUnknown = errors.New("unknown push template")

// Copy is the title and body shown on the device.
type Copy struct {
	Title string
	Body  string
}

// templates is the whole surface of what push may say. It is deliberately a
// closed set: a notification is readable on a locked screen, so nothing here
// names a counterpart, quotes a room, or reveals what a member is doing.
var templates = map[string]Copy{
	"morning_greeting": {
		Title: "Obiara",
		Body:  "A new day in your courtyard. Something is waiting for you.",
	},
	"evening_check": {
		Title: "Obiara",
		Body:  "The evening is a good time to catch up.",
	},
	"fire_herald": {
		Title: "A Fire is starting soon",
		Body:  "Open Obiara to join when you are ready.",
	},
	"pod_alert": {
		Title: "Obiara",
		Body:  "Something new has arrived for you.",
	},
	"room_activity": {
		Title: "Obiara",
		Body:  "There is activity in one of your rooms.",
	},
	"safety_notice": {
		Title: "Obiara safety",
		Body:  "Please open Obiara. There is something you need to see.",
	},
}

// CopyFor returns the approved copy for a template.
func CopyFor(template string) (Copy, error) {
	copy, ok := templates[template]
	if !ok {
		return Copy{}, ErrTemplateUnknown
	}
	return copy, nil
}

// FallbackCopy is used when a routed notification names a template that has
// no push copy yet. It says nothing specific, which is the safe default for a
// lock screen, and the caller logs the gap.
func FallbackCopy() Copy {
	return Copy{Title: "Obiara", Body: "You have a new notification."}
}
