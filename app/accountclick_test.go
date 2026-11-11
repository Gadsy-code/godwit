package app

import (
	"testing"

	"github.com/godwit-email/godwit/core"
)

// With Settings.AccountOpensInbox on, a click on an account in the sidebar opens its inbox at once and
// unfolds it, instead of only unfolding it.
func TestAccountClickOpensInbox(t *testing.T) {
	h, accs := mailHarness(t, nil, "alice@example.com")
	st := h.app.settings
	st.AccountOpensInbox = true
	h.post(func() { h.app.changeSettings(st) })
	h.drain()
	h.app.navigate(Route{Page: PageMail, View: core.View{Unified: "starred"}})
	h.drain()
	h.tt.Frame()
	inbox := func() int64 {
		for _, f := range h.app.mail.folders[accs[0].ID] {
			if f.Role == "inbox" {
				return f.ID
			}
		}
		return 0
	}
	h.until("the account's folders", func() bool { return inbox() != 0 })
	h.click("alice@example.com")
	h.until("the account's inbox shown", func() bool { return h.app.route.View.FolderID == inbox() })
	if !*h.app.mail.accountOpen(accs[0].ID) {
		t.Error("the account stays folded")
	}
	// Off, a click folds or unfolds, as it always did.
	st.AccountOpensInbox = false
	h.post(func() { h.app.changeSettings(st) })
	h.drain()
	h.app.navigate(Route{Page: PageMail, View: core.View{Unified: "starred"}})
	h.drain()
	h.tt.Frame()
	was := *h.app.mail.accountOpen(accs[0].ID)
	h.click("alice@example.com")
	if h.app.route.View.FolderID == inbox() {
		t.Error("a click went to the inbox with the option off")
	}
	if *h.app.mail.accountOpen(accs[0].ID) == was {
		t.Error("a click did not fold or unfold it with the option off")
	}
}
