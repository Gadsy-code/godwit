package app

import (
	"github.com/egoist/mygo/ui"
	"github.com/godwit-email/godwit/core"
	"github.com/godwit-email/godwit/internal/theme"
)

// rowTint is the light tint, and the bar of its stronger colour, a list row takes from its labels: the first one set by
// hand or a rule, else the first the model set. ok is false for a row with no label.
func (a *App) rowTint(c *ui.Context, labels []core.AppliedLabel) (tint, bar ui.Color, ok bool) {
	pick := -1
	for i, al := range labels {
		if _, ok := a.label(al.LabelID); !ok {
			continue
		}
		if !al.Auto {
			pick = i
			break
		}
		if pick < 0 {
			pick = i
		}
	}
	if pick < 0 {
		return
	}
	l, _ := a.label(labels[pick].LabelID)
	lc := theme.Of(c).Label(l.Color)
	return lc.Fill.Alpha(theme.RowTintAlpha), lc.Text, true
}
