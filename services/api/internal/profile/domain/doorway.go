package domain

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

// DoorwayQuestionLimit bounds the one question sowers must answer
// (Doc 06 S-07: 60 characters, 8 suggestions or write-own).
const DoorwayQuestionLimit = 60

var ErrDoorwayQuestionInvalid = errors.New("doorway question must be 1-60 safe characters")

// DoorwayQuestion is the single prompt a member's sowers answer. Exactly
// one per member; updates replace it wholesale.
type DoorwayQuestion struct {
	memberID  string
	text      string
	custom    bool
	updatedAt time.Time
}

func NewDoorwayQuestion(memberID, text string, custom bool, now time.Time) (DoorwayQuestion, error) {
	text = trimAndCollapse(text)
	if !utf8.ValidString(text) || text == "" || utf8.RuneCountInString(text) > DoorwayQuestionLimit {
		return DoorwayQuestion{}, ErrDoorwayQuestionInvalid
	}
	if containsDisallowedPersonalData(text) {
		return DoorwayQuestion{}, ErrUnsafeProfile
	}
	return DoorwayQuestion{memberID: memberID, text: text, custom: custom, updatedAt: now.UTC()}, nil
}

func (question DoorwayQuestion) MemberID() string     { return question.memberID }
func (question DoorwayQuestion) Text() string         { return question.text }
func (question DoorwayQuestion) Custom() bool         { return question.custom }
func (question DoorwayQuestion) UpdatedAt() time.Time { return question.updatedAt }

func trimAndCollapse(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
