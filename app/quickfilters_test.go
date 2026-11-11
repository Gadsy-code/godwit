package app

import (
	"bytes"
	"testing"

	"github.com/emersion/go-imap/v2"

	"github.com/godwit-email/godwit/core"
	"github.com/godwit-email/godwit/core/coretest"
	"github.com/godwit-email/godwit/internal/i18n"
)

// With Settings.QuickFilters on, the list has Starred and Attachments tabs and a label
// menu beside All and Unread, and each narrows the list; off, they are not there.
func TestQuickFilters(t *testing.T) {
	h, accs := mailHarness(t, nil, "alice@example.com")
	acc := accs[0]
	for _, m := range []struct {
		subj string
		opts []coretest.MsgOption
	}{
		{"Plain note", nil},
		{"Starred note", nil},
		{"File note", []coretest.MsgOption{coretest.Attachment("a.pdf", "application/pdf", []byte("x"))}},
		{"Tagged note", nil},
	} {
		raw := coretest.Msg("bob@example.com", acc.Email, m.subj, "body", m.opts...)
		var flags []imap.Flag
		if m.subj == "Starred note" {
			flags = append(flags, imap.FlagFlagged)
		}
		if _, err := h.env.IMAP(acc.ID).User.Append("INBOX", bytes.NewReader(raw), &imap.AppendOptions{Flags: flags}); err != nil {
			t.Fatal(err)
		}
	}
	h.syncNow(acc)
	h.waitMail("the four mails", func() bool { return h.tt.HasText("Plain note") && h.tt.HasText("Tagged note") })
	page, err := h.env.Svc.ListMessages(t.Context(), core.View{Unified: "inbox"}, "", 10)
	if err != nil {
		t.Fatal(err)
	}
	id := map[string]int64{}
	for _, m := range page.Items {
		id[m.Subject] = m.ID
	}
	h.autoLabel(id["Tagged note"], "travel")
	h.settle()

	starred := i18n.T("mail.filter.starred")
	if _, ok := h.tt.Find(starred); ok {
		t.Fatalf("the Starred tab shows with QuickFilters off")
	}
	st := h.app.settings
	st.QuickFilters = true
	h.post(func() { h.app.changeSettings(st) })
	h.drain()
	h.until("the Starred tab", func() bool { _, ok := h.tt.Find(starred); return ok })

	only := func(want string, others ...string) {
		t.Helper()
		h.until("only "+want, func() bool {
			h.tt.Frame()
			if !h.tt.HasText(want) {
				return false
			}
			for _, o := range others {
				if h.tt.HasText(o) {
					return false
				}
			}
			return true
		})
	}
	h.click(starred)
	only("Starred note", "Plain note", "File note", "Tagged note")
	h.click(i18n.T("mail.filter.attach"))
	only("File note", "Plain note", "Starred note", "Tagged note")
	h.click(i18n.T("mail.filter.all"))
	only("Plain note", "Zzz")
	h.see("Starred note")

	ls, err := h.env.Svc.ListLabels(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range ls {
		if l.Slug == "travel" {
			h.waitMail("the row shows its chip", func() bool { return h.tt.HasText(l.Name) })
			h.click(i18n.T("mail.filter.label.name"))
			h.click(i18n.Tf("mail.filter.label.item", l.Name))
		}
	}
	only("Tagged note", "Plain note", "File note", "Starred note")
}

// The switch is in Settings › General.
func TestQuickFiltersSetting(t *testing.T) {
	h, _ := withAccount(t)
	h.openSettings("general")
	h.see(i18n.T("settings.quick_filters"))
}
