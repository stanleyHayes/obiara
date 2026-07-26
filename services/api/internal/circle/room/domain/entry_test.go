package domain

import (
	"strings"
	"testing"
	"time"
)

func TestVoiceOnlyStoresRefsAndRetentionIsBounded(t *testing.T) {
	n := time.Now().UTC()
	m, _ := NewMediaRef("asset:1", "transcript:1", "audio/ogg", time.Minute)
	e, err := New(Params{ID: "entry:1", CircleID: "circle:1", AuthorKey: strings.Repeat("a", 64), CommandID: "cmd:1", Kind: KindVoice, Media: m, CreatedAt: n, ExpiresAt: n.Add(MaxRetention)})
	if err != nil || !e.Visible(n) || e.Media().AssetID() != "asset:1" {
		t.Fatal(e, err)
	}
	_, err = New(Params{ID: "entry:2", CircleID: "circle:1", AuthorKey: strings.Repeat("a", 64), CommandID: "cmd:2", Kind: KindVoice, Media: m, CreatedAt: n, ExpiresAt: n.Add(MaxRetention + time.Second)})
	if err == nil {
		t.Fatal("unbounded retention accepted")
	}
}
func TestEventsAndNoticesUseOpaqueContentRefs(t *testing.T) {
	n := time.Now().UTC()
	for _, k := range []Kind{KindEvent, KindNotice} {
		p := Params{ID: "entry:1", CircleID: "circle:1", AuthorKey: strings.Repeat("a", 64), ContentRef: "content:1", CommandID: "cmd:1", Kind: k, CreatedAt: n, ExpiresAt: n.Add(time.Hour)}
		if k == KindEvent {
			p.StartsAt = n.Add(time.Hour)
			p.EndsAt = n.Add(2 * time.Hour)
		}
		if _, e := New(p); e != nil {
			t.Fatal(k, e)
		}
	}
}
func FuzzRetentionNeverExceedsBound(f *testing.F) {
	f.Add(int64(1))
	f.Fuzz(func(t *testing.T, s int64) {
		if s <= 0 || s > int64((MaxRetention+time.Hour)/time.Second) {
			t.Skip()
		}
		n := time.Unix(1, 0)
		m, _ := NewMediaRef("asset:1", "", "audio/ogg", time.Second)
		e, err := New(Params{ID: "entry:1", CircleID: "circle:1", AuthorKey: strings.Repeat("a", 64), CommandID: "cmd:1", Kind: KindVoice, Media: m, CreatedAt: n, ExpiresAt: n.Add(time.Duration(s) * time.Second)})
		if err == nil && e.ExpiresAt().Sub(e.CreatedAt()) > MaxRetention {
			t.Fatal("retention exceeded")
		}
	})
}
