package kit

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
	"github.com/godwit-email/godwit/internal/theme"
)

// LabelInitial is the letter a label shows when it has no glyph: the first
// letter of its name, or of the part after a " · " when the name has one (a
// "Client · Acme" label shows A). "?" when there is nothing to show.
func LabelInitial(name string) string {
	if i := strings.LastIndex(name, "·"); i >= 0 {
		name = name[i+len("·"):]
	}
	name = strings.TrimSpace(name)
	r, _ := utf8.DecodeRuneInString(name)
	if name == "" {
		return "?"
	}
	return string(unicode.ToUpper(r))
}

// LabelBadgeOpts is a label's mark: its Image, else its initial, in a rounded
// square like an Avatar's.
type LabelBadgeOpts struct {
	Name  string
	Size  float32
	Color theme.LabelColor
	Image ui.ImageSource
}

// LabelBadge draws a label's image or its initial in a rounded square. An image
// is clipped to the square and edged, so any logo sits the way an app icon does.
func LabelBadge(c *ui.Context, o LabelBadgeOpts) ui.Element {
	t := theme.Of(c)
	radius := t.Radius.Md
	font := t.Text.Xs // a letter in a small badge, a step up in a large one
	if o.Size >= theme.IconLg {
		font = t.Text.Sm
	}
	if o.Image != nil {
		return ui.Box(c).Size(o.Size, o.Size).Shrink(0).Radius(radius).Clip().Border(theme.Stroke, t.Border).Label(o.Name).Children(func() {
			ui.Image(c, o.Image).Size(o.Size, o.Size).Fit(ui.Cover)
		})
	}
	return ui.Box(c).Size(o.Size, o.Size).Shrink(0).Radius(radius).Background(o.Color.Fill).Center().Label(o.Name).Children(func() {
		ui.Text(c, LabelInitial(o.Name)).FontWeight(600).FontSize(font).LineHeight(theme.LineTight).TextColor(o.Color.Text)
	})
}
