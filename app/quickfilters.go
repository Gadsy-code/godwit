package app

import (
	"github.com/egoist/mygo/ui"
	"github.com/godwit-email/godwit/internal/i18n"
	"github.com/godwit-email/godwit/internal/icons"
	"github.com/godwit-email/godwit/internal/kit"
	"github.com/godwit-email/godwit/internal/theme"
)

// The list's filters, as p.filter holds them: all, unread, starred, with an attachment.
const (
	filterAll = iota
	filterUnread
	filterStarred
	filterAttach
)

// active is the list's filter and label filter. Without the quick filters
// only All and Unread exist, so a leftover choice of the others counts for nothing.
func (p *mailPage) active() (filter int, label int64) {
	if p.a.settings.QuickFilters {
		return p.filter, p.quickLabel
	}
	return min(p.filter, filterUnread), 0
}

// filtered reports whether any filter narrows the list.
func (p *mailPage) filtered() bool {
	f, l := p.active()
	return f != filterAll || l != 0
}

// keeps reports whether a row passes the list's filters.
func (p *mailPage) keeps(it mailItem) bool {
	f, label := p.active()
	switch {
	case f == filterUnread && it.Unread == 0,
		f == filterStarred && it.Flagged == 0,
		f == filterAttach && !it.Attach:
		return false
	}
	if label == 0 {
		return true
	}
	for _, al := range it.Labels {
		if al.LabelID == label {
			return true
		}
	}
	return false
}

// quickFilters is the row under the list's heading with the filters at hand:
// all, unread, starred, with an attachment, and one label.
func (p *mailPage) quickFilters(c *ui.Context) {
	t, th := theme.Of(c), c.Theme()
	ui.Row(c).AlignItems(ui.Center).Wrap().Padding(0, th.Space(4), th.Space(2), th.Space(4)).Gap(th.Space(2)).Children(func() {
		kit.Tabs(c, &p.filter, i18n.T("mail.filter.all"), i18n.T("mail.filter.unread"), i18n.T("mail.filter.starred"), i18n.T("mail.filter.attach"))
		cur, set := p.a.label(p.quickLabel)
		set = set && p.quickLabel != 0
		name := i18n.T("mail.filter.label")
		if set {
			name = cur.Name
		}
		chip := filterChip(c, name, set, &p.labelOpen, nil).Label(i18n.T("mail.filter.label.name"))
		kit.Popover(c, chip, &p.labelOpen, func() {
			filterItem(c, i18n.T("mail.filter.label.any"), !set, func() { p.quickLabel, p.labelOpen = 0, false })
			for _, l := range p.a.labels {
				id, on := l.ID, set && l.ID == cur.ID
				kit.Pressable(c.Key(id), kit.PressableOpts{Selected: on, OnClick: func() { p.quickLabel, p.labelOpen = id, false }}).
					Label(i18n.Tf("mail.filter.label.item", l.Name)).Checked(on).MinWidth(theme.FilterMenuMin).Padding(th.Space(1.5), th.Space(2)).
					Gap(th.Space(2)).AlignItems(ui.Center).Radius(t.Radius.Md).Children(func() {
					ui.Icon(c, icons.Get(labelIcon(l.Icon))).Size(theme.IconXs, theme.IconXs).TextColor(t.Label(l.Color).Text)
					kit.Text(c, l.Name, t.Text.Sm).SingleLine()
				})
			}
		})
	})
}
