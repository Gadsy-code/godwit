package app

import (
	"image/color"
	"testing"

	"github.com/godwit-email/godwit/core"
	"github.com/godwit-email/godwit/core/coretest"
)

// With Settings.UnreadBadges on, the unread counts in the sidebar sit in coloured circles (amber for a few).
func TestUnreadBadgesInSidebar(t *testing.T) {
	h, accs := mailHarness(t, nil, "alice@example.com")
	h.deliverOne(accs[0], coretest.Msg("bob@example.com", accs[0].Email, "One unread", "text"))
	h.waitMail("the mail", func() bool { return h.tt.HasText("One unread") })
	amber := func() int { // pixels of the badge's amber
		n, img := 0, h.tt.Image()
		b := img.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Min.X+320; x++ { // the sidebar's width
				if px := img.RGBAAt(x, y); isAmber(px) {
					n++
				}
			}
		}
		return n
	}
	h.tt.Frame()
	if n := amber(); n > 0 {
		t.Fatalf("%d amber pixels in the sidebar with the badges off", n)
	}
	st := h.app.settings
	st.UnreadBadges = true
	h.post(func() { h.app.changeSettings(st) })
	h.drain()
	h.until("an amber badge in the sidebar", func() bool { h.tt.Frame(); return amber() > 20 })
	_ = core.View{}
}

func isAmber(px color.RGBA) bool { return px.R > 225 && px.G > 130 && px.G < 190 && px.B < 60 }
