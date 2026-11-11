package app

import (
	"testing"

	"github.com/godwit-email/godwit/core"
	"github.com/godwit-email/godwit/core/coretest"
	"github.com/godwit-email/godwit/internal/i18n"
	"github.com/godwit-email/godwit/internal/theme"
)

// setLayout changes the settings the way the settings page does.
func (h *harness) setLayout(f func(*core.Settings)) {
	h.t.Helper()
	st := h.app.settings
	f(&st)
	h.post(func() { h.app.changeSettings(st) })
	h.drain()
	h.tt.Frame()
}

func (h *harness) top(label string) float32 {
	h.t.Helper()
	h.until("an element "+label, func() bool { _, ok := h.tt.Find(label); return ok })
	r, _ := h.tt.Find(label)
	return r.Y
}

// The sidebar's sections follow Settings.SidebarOrder.
func TestSidebarSectionOrder(t *testing.T) {
	h, _ := mailHarness(t, nil, "alice@example.com")
	views, accounts := i18n.T("view.title"), i18n.T("mail.accounts")
	if h.top(views) > h.top(accounts) {
		t.Fatalf("by default the views (%v) come before the accounts (%v)", h.top(views), h.top(accounts))
	}
	h.setLayout(func(st *core.Settings) { st.SidebarOrder = "accounts,labels,views" })
	h.until("the accounts first", func() bool { return h.top(accounts) < h.top(views) })
}

// A list row leads with who it is with, or with the subject, as Settings.ListRowOrder says.
func TestListRowOrder(t *testing.T) {
	h, accs := mailHarness(t, nil, "alice@example.com")
	h.deliverOne(accs[0], coretest.Msg("bob@example.com", accs[0].Email, "Quarterly numbers", "text"))
	h.waitMail("one thread", func() bool { return len(h.items()) == 1 })
	if h.top("bob@example.com") > h.top("Quarterly numbers") {
		t.Fatal("by default the sender leads the row")
	}
	h.setLayout(func(st *core.Settings) { st.ListRowOrder = core.ListRowSubject })
	h.until("the subject first", func() bool { return h.top("Quarterly numbers") < h.top("bob@example.com") })
}

// The reading column's widest size follows Settings.ReaderWidth.
func TestReaderMaxWidth(t *testing.T) {
	h, _ := mailHarness(t, nil, "alice@example.com")
	normal := h.app.readerMaxWidth()
	if normal != theme.ReaderMaxWidth {
		t.Fatalf("default width %v", normal)
	}
	h.setLayout(func(st *core.Settings) { st.ReaderWidth = core.ReaderWidthWide })
	h.until("a wider column", func() bool { return h.app.readerMaxWidth() > normal })
	wide := h.app.readerMaxWidth()
	h.setLayout(func(st *core.Settings) { st.ReaderWidth = core.ReaderWidthFull })
	h.until("the widest column", func() bool { return h.app.readerMaxWidth() > wide })
}

// The three layout choices are in Settings › General, and choosing one changes the settings.
func TestLayoutSettingsPage(t *testing.T) {
	h, _ := withAccount(t)
	p := h.openSettings("general")
	for _, key := range []string{"settings.sidebar_order", "settings.list_order", "settings.reader_width"} {
		h.see(i18n.T(key))
	}
	p.st.ReaderWidth = core.ReaderWidthWide
	h.drain()
	h.tt.Frame()
	if got := h.app.settings.ReaderWidth; got != core.ReaderWidthWide {
		t.Fatalf("reader width %q after choosing it", got)
	}
}
