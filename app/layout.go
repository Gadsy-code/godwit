package app

import (
	"github.com/godwit-email/godwit/core"
	"github.com/godwit-email/godwit/internal/i18n"
	"github.com/godwit-email/godwit/internal/kit"
	"github.com/godwit-email/godwit/internal/theme"
)

// readerMaxWidth is the widest the reading column may grow, as Settings.ReaderWidth says.
func (a *App) readerMaxWidth() float32 {
	switch a.settings.ReaderWidth {
	case core.ReaderWidthWide:
		return theme.ReaderWideMaxWidth
	case core.ReaderWidthFull:
		return theme.ReaderFullMaxWidth
	}
	return theme.ReaderMaxWidth
}

// sidebarOrderOptions are the six orders of the sidebar's sections, the default first.
func sidebarOrderOptions() []kit.Option[string] {
	name := map[string]string{"views": i18n.T("view.title"), "labels": i18n.T("mail.labels"), "accounts": i18n.T("mail.accounts")}
	var opts []kit.Option[string]
	for _, a := range core.SidebarSections {
		for _, b := range core.SidebarSections {
			for _, c := range core.SidebarSections {
				if a != b && b != c && a != c {
					opts = append(opts, kit.Option[string]{Value: a + "," + b + "," + c, Label: name[a] + " › " + name[b] + " › " + name[c]})
				}
			}
		}
	}
	return opts
}
