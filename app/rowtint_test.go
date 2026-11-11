package app

import (
	"testing"

	"github.com/godwit-email/godwit/internal/i18n"
)

// With Settings.TintRows on, a list row takes a light tint of its label's colour.
func TestRowTint(t *testing.T) {
	h, _, m := labeledMail(t)
	h.autoLabel(m.ID, "travel") // blue
	ls, err := h.env.Svc.ListLabels(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range ls {
		if l.Slug == "travel" {
			h.waitMail("the row shows its chip", func() bool { return h.tt.HasText(l.Name) })
		}
	}
	at := func() (r, g, b uint8) {
		rect, ok := h.tt.Find("Check-in open")
		if !ok {
			return 0, 0, 0
		}
		px := h.tt.Image().RGBAAt(int(rect.X)-4, int(rect.Y+rect.H/2))
		return px.R, px.G, px.B
	}
	r0, g0, b0 := at()
	st := h.app.settings
	st.TintRows = true
	h.post(func() { h.app.changeSettings(st) })
	h.drain()
	h.until("the row's background changes", func() bool {
		h.tt.Frame()
		r, g, b := at()
		return r != r0 || g != g0 || b != b0
	})
	r1, _, b1 := at()
	if int(b1)-int(r1) <= int(b0)-int(r0) {
		t.Errorf("the tint is not toward blue: (%d,%d,%d) -> (%d,_,%d)", r0, g0, b0, r1, b1)
	}
}

// The switch is in Settings › Labels.
func TestRowTintSetting(t *testing.T) {
	h, _ := withAccount(t)
	h.openSettings("labels")
	h.see(i18n.T("settings.labels.tint"))
}

// A row can carry a solid bar of its label's colour on its left edge (Settings.RowBar), with or
// without the tint: it shows where the tint alone is too light to tell rows apart.
func TestRowTintBar(t *testing.T) {
	h, _, m := labeledMail(t)
	h.autoLabel(m.ID, "travel") // blue
	ls, err := h.env.Svc.ListLabels(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range ls {
		if l.Slug == "travel" {
			h.waitMail("the row shows its chip", func() bool { return h.tt.HasText(l.Name) })
		}
	}
	st := h.app.settings
	st.RowBar = true // the bar alone: the tint stays off
	h.post(func() { h.app.changeSettings(st) })
	h.drain()
	bar := func() bool { // a strongly blue pixel in the row's left margin
		rect, ok := h.tt.Find("Check-in open")
		if !ok {
			return false
		}
		y := int(rect.Y + rect.H/2)
		for x := int(rect.X) - 30; x < int(rect.X); x++ {
			if x < 0 {
				continue
			}
			if px := h.tt.Image().RGBAAt(x, y); int(px.B)-int(px.R) > 90 {
				return true
			}
		}
		return false
	}
	h.until("a bar of the label's colour on the row's edge", func() bool { h.tt.Frame(); return bar() })
}

// A row can carry a small dot of its label's colour before its first line (Settings.RowDot): the quiet
// way, where the bar or the tint is too much.
func TestRowDot(t *testing.T) {
	h, _, m := labeledMail(t)
	h.autoLabel(m.ID, "travel") // blue
	ls, err := h.env.Svc.ListLabels(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range ls {
		if l.Slug == "travel" {
			h.waitMail("the row shows its chip", func() bool { return h.tt.HasText(l.Name) })
		}
	}
	// With the dot on the first line starts a dot's width and gap further in, the dot in that room; the
	// unread mark is further left than any of it.
	dotShown := func() bool {
		rect, ok := h.tt.Find("bob@example.com")
		if !ok {
			return false
		}
		y := int(rect.Y + rect.H/2)
		for x := int(rect.X) - 13; x < int(rect.X)-6; x++ {
			if x < 0 {
				continue
			}
			if px := h.tt.Image().RGBAAt(x, y); int(px.B)-int(px.R) > 90 {
				return true
			}
		}
		return false
	}
	st := h.app.settings
	st.RowDot = true // the dot alone: no tint, no bar
	h.post(func() { h.app.changeSettings(st) })
	h.drain()
	h.until("a dot of the label's colour before the first line", func() bool { h.tt.Frame(); return dotShown() })
}
