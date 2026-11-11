package app

import (
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/godwit-email/godwit/core"
	"github.com/godwit-email/godwit/internal/icons"
	"github.com/godwit-email/godwit/internal/kit"
	"github.com/godwit-email/godwit/internal/theme"
)

// labelSep sets a label's group apart from its own name: "Client · Acme" is Acme in the group
// Client, nested in the sidebar the way Gmail nests "Client/Acme".
const labelSep = " · "

// sideLabel is a label as the sidebar shows it: by its short name, under its group.
type sideLabel struct {
	core.Label
	Short string
}

// sideEntry is one row of the sidebar's label list: a label alone (Group ""), or a group and its
// labels, in the order their labels come.
type sideEntry struct {
	Group  string
	Labels []sideLabel
}

// groupLabels nests the labels whose names have a group before the first labelSep: a group stands
// where its first label does. A name with nothing on either side of the separator stays whole.
func groupLabels(ls []core.Label) []sideEntry {
	var out []sideEntry
	at := map[string]int{}
	for _, l := range ls {
		g, short, ok := strings.Cut(l.Name, labelSep)
		g, short = strings.TrimSpace(g), strings.TrimSpace(short)
		if !ok || g == "" || short == "" {
			out = append(out, sideEntry{Labels: []sideLabel{{l, l.Name}}})
			continue
		}
		if i, seen := at[g]; seen {
			out[i].Labels = append(out[i].Labels, sideLabel{l, short})
			continue
		}
		at[g] = len(out)
		out = append(out, sideEntry{Group: g, Labels: []sideLabel{{l, short}}})
	}
	return out
}

// labelGroupKey names a group's fold: the account's (0: every account's), and the group.
type labelGroupKey struct {
	acc   int64
	group string
}

// labelGroupOpen is whether a group of labels is unfolded in the sidebar; default yes.
func (p *mailPage) labelGroupOpen(acc int64, group string) *bool {
	if p.labelGroups == nil {
		p.labelGroups = map[labelGroupKey]*bool{}
	}
	k := labelGroupKey{acc, group}
	if p.labelGroups[k] == nil {
		open := true
		p.labelGroups[k] = &open
	}
	return p.labelGroups[k]
}

// labelRows draws labels in the sidebar, nested by their groups: rowOf draws one label's row, by its
// name (a group's own is the short one), indented as the rows of that list are. A label with no
// mail (Count 0) is left out, and a group with none of its own.
func (p *mailPage) labelRows(c *ui.Context, labels []core.Label, acc int64, indent float32, rowOf func(l core.Label, name string, indent float32)) {
	t := theme.Of(c)
	for _, e := range groupLabels(labels) {
		var shown []sideLabel
		unread := 0
		for _, l := range e.Labels {
			if l.Count > 0 {
				shown = append(shown, l)
				unread += l.Unread
			}
		}
		if len(shown) == 0 {
			continue
		}
		if e.Group == "" {
			rowOf(shown[0].Label, shown[0].Short, indent)
			continue
		}
		open := p.labelGroupOpen(acc, e.Group)
		kit.SidebarCollapsible(c.Key(e.Group), open, true, func() {
			ui.Icon(c, icons.Get("label-folder-3")).Size(theme.IconXs, theme.IconXs).Shrink(0)
			ui.Text(c, e.Group).SingleLine().Grow(1).Shrink(1)
			if !*open && unread > 0 {
				kit.Text(c, strconv.Itoa(unread), t.Text.Sm).TextColor(t.Muted)
			}
		}, func() {
			for _, l := range shown {
				rowOf(l.Label, l.Short, indent+theme.LabelGroupIndent)
			}
		})
	}
}
