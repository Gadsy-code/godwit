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
