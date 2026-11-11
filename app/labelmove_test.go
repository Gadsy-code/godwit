package app

import (
	"testing"

	"github.com/godwit-email/godwit/internal/i18n"
)

// A label's row in Settings › Labels has arrows that move it up or down a place, for where a drag is awkward
// (a long list scrolls under the pointer).
func TestLabelMoveButtons(t *testing.T) {
	h, _ := withAccount(t)
	h.openSettings("labels")
	order := func() []string {
		ls, err := h.env.Svc.ListLabels(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		out := make([]string, len(ls))
		for i, l := range ls {
			out[i] = l.Slug
		}
		return out
	}
	before := order()
	ls, _ := h.env.Svc.ListLabels(t.Context())
	first, second := ls[0], ls[1]
	h.click(i18n.Tf("settings.labels.move_down", first.Name))
	h.until("the first label one place down", func() bool {
		o := order()
		return o[0] == before[1] && o[1] == before[0]
	})
	h.click(i18n.Tf("settings.labels.move_up", first.Name))
	h.until("it back up", func() bool {
		o := order()
		return o[0] == before[0] && o[1] == before[1]
	})
	_ = second
}
