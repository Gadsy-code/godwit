package app

import (
	"github.com/godwit-email/godwit/core"
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
