package app

import (
	"strconv"

	"github.com/egoist/mygo/ui"
	"github.com/godwit-email/godwit/internal/kit"
	"github.com/godwit-email/godwit/internal/theme"
)

// unreadMark is an unread count in the sidebar: a coloured circle (Settings.UnreadBadges), or the small muted number.
func (a *App) unreadMark(c *ui.Context, n int) {
	if a.settings.UnreadBadges {
		kit.CountBadge(c, n)
		return
	}
	kit.Text(c, strconv.Itoa(n), theme.Of(c).Text.Sm).TextColor(theme.Of(c).Muted)
}
