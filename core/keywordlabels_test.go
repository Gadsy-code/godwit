package core

import (
	"testing"

	"github.com/godwit-email/godwit/core/store"
	"github.com/stretchr/testify/require"
)

// Tags written on the mail server as IMAP keywords (by an agent, a script or
// another client) show up in Godwit as labels. The seam is the Service: flags as
// a sync stored them in, labels as the reader asks for them.

func addMailWithFlags(t *testing.T, s *Service, mid, flags string) int64 {
	t.Helper()
	res, err := s.db.Execute(t.Context(), `INSERT INTO messages(account_id, message_id, subject, from_addr, date, label_fp, flags)
		VALUES (1, ?, 's', 'a@x', 1, ?, ?)`, mid, "fp:"+mid, flags)
	require.NoError(t, err)
	return res.LastInsertID
}

func appliedSlugs(t *testing.T, s *Service, msg int64) []string {
	t.Helper()
	applied, err := s.MessageLabels(t.Context(), []int64{msg})
	require.NoError(t, err)
	ls, err := s.ListLabels(t.Context())
	require.NoError(t, err)
	bySlugID := map[int64]string{}
	for _, l := range ls {
		bySlugID[l.ID] = l.Slug
	}
	var out []string
	for _, a := range applied[msg] {
		out = append(out, bySlugID[a.LabelID])
	}
	return out
}

func TestKeywordOnTheServerBecomesALabel(t *testing.T) {
	s := newLabelService(t, "", "en")
	msg := addMailWithFlags(t, s, "a@x", `\Seen cliente-acme`)

	require.NoError(t, s.importKeywordLabels(t.Context(), 1))

	require.Equal(t, []string{"cliente-acme"}, appliedSlugs(t, s, msg))
}

func TestKeywordMatchingAnExistingLabelUsesIt(t *testing.T) {
	s := newLabelService(t, "", "en")
	before, err := s.ListLabels(t.Context())
	require.NoError(t, err)
	msg := addMailWithFlags(t, s, "b@x", `newsletter`)

	require.NoError(t, s.importKeywordLabels(t.Context(), 1))

	require.Equal(t, []string{"newsletter"}, appliedSlugs(t, s, msg))
	after, err := s.ListLabels(t.Context())
	require.NoError(t, err)
	require.Len(t, after, len(before), "the preset is reused, not duplicated")
}

func TestSystemFlagsAreNotLabels(t *testing.T) {
	s := newLabelService(t, "", "en")
	msg := addMailWithFlags(t, s, "c@x", `\Seen \Flagged $Forwarded $MDNSent NonJunk`)

	require.NoError(t, s.importKeywordLabels(t.Context(), 1))

	require.Empty(t, appliedSlugs(t, s, msg))
}

func TestKeywordRemovedOnTheServerTakesTheLabelOff(t *testing.T) {
	s := newLabelService(t, "", "en")
	msg := addMailWithFlags(t, s, "d@x", `cliente-acme`)
	require.NoError(t, s.importKeywordLabels(t.Context(), 1))
	require.Equal(t, []string{"cliente-acme"}, appliedSlugs(t, s, msg))

	_, err := s.db.Execute(t.Context(), "UPDATE messages SET flags = '\\Seen' WHERE id = ?", msg)
	require.NoError(t, err)
	require.NoError(t, s.importKeywordLabels(t.Context(), 1))

	require.Empty(t, appliedSlugs(t, s, msg))
}

func TestHandMadeChoiceBeatsTheServerKeyword(t *testing.T) {
	s := newLabelService(t, "", "en")
	msg := addMailWithFlags(t, s, "e@x", `newsletter`)
	news := labelIDs(t, s)["newsletter"]
	require.NoError(t, s.SetLabel(t.Context(), []int64{msg}, news, false)) // taken off by hand

	require.NoError(t, s.importKeywordLabels(t.Context(), 1))

	require.Empty(t, appliedSlugs(t, s, msg), "a label removed by hand is not put back by the import")
}

// The other way round: a label put on by hand is written to the server as a keyword.

func pendingFlagOps(t *testing.T, s *Service, msg int64) []string {
	t.Helper()
	rows, err := s.db.Query(t.Context(), "SELECT payload FROM ops WHERE message_id = ? AND kind = 'flags' ORDER BY id", msg)
	require.NoError(t, err)
	var out []string
	for _, v := range rows.Values {
		out = append(out, store.Str(v[0]))
	}
	return out
}

func TestLabelPutOnByHandIsQueuedAsAServerKeyword(t *testing.T) {
	s := newLabelService(t, "", "en")
	msg := addMailWithFlags(t, s, "f@x", `\Seen`)
	news := labelIDs(t, s)["newsletter"]

	require.NoError(t, s.SetLabel(t.Context(), []int64{msg}, news, true))

	require.Equal(t, []string{`{"add":["newsletter"]}`}, pendingFlagOps(t, s, msg))
	rows, err := s.db.Query(t.Context(), "SELECT flags FROM messages WHERE id = ?", msg)
	require.NoError(t, err)
	require.Contains(t, store.Str(rows.Values[0][0]), "newsletter")
}

func TestLabelTakenOffByHandRemovesTheServerKeyword(t *testing.T) {
	s := newLabelService(t, "", "en")
	msg := addMailWithFlags(t, s, "g@x", `newsletter`)
	news := labelIDs(t, s)["newsletter"]

	require.NoError(t, s.SetLabel(t.Context(), []int64{msg}, news, false))

	require.Equal(t, []string{`{"remove":["newsletter"]}`}, pendingFlagOps(t, s, msg))
}
