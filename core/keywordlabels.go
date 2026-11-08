package core

// IMAP keywords on the server become labels in Godwit.
//
// A keyword is any flag that is not a system flag (\Seen), not a $-flag
// ($Forwarded) and not a junk marker. Its lowercase form is the label slug, the
// identity docs/adr/0014 reserved for this. Rows get source 'imap': a label put on
// or taken off by hand ('manual') and one a rule set ('rule') are never overruled.

import (
	"context"
	"regexp"
	"strings"

	"github.com/egoist/mygo/plugins/sqlite"
	"github.com/godwit-email/godwit/core/events"
	"github.com/godwit-email/godwit/core/store"
)

var keywordSlug = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)

// notLabelKeywords are plain flags servers and clients use for their own ends.
var notLabelKeywords = map[string]bool{"nonjunk": true, "junk": true, "notjunk": true, "forwarded": true, "mdnsent": true}

// keywordsOf returns the label slugs a message's flag string carries.
func keywordsOf(flags string) []string {
	var out []string
	for _, f := range strings.Fields(flags) {
		if strings.HasPrefix(f, `\`) || strings.HasPrefix(f, "$") {
			continue
		}
		slug := strings.ToLower(f)
		if notLabelKeywords[slug] || !keywordSlug.MatchString(slug) {
			continue
		}
		out = append(out, slug)
	}
	return out
}

// importKeywordLabels makes the account's cached messages carry the labels the
// server's keywords name, creating a label for a keyword it has none for.
func (s *Service) importKeywordLabels(ctx context.Context, accountID int64) error {
	s.labelsMu.Lock()
	defer s.labelsMu.Unlock()
	rows, err := s.db.Query(ctx, "SELECT id, flags FROM messages WHERE account_id = ? AND flags != ''", accountID)
	if err != nil {
		return err
	}
	byMsg := map[int64][]string{}
	wanted := map[string]bool{}
	for _, v := range rows.Values {
		kws := keywordsOf(store.Str(v[1]))
		if len(kws) == 0 {
			continue
		}
		byMsg[store.Int(v[0])] = kws
		for _, k := range kws {
			wanted[k] = true
		}
	}
	var stmts []sqlite.Statement
	for slug := range wanted {
		stmts = append(stmts, sqlite.Statement{SQL: `INSERT INTO labels(slug, name, color, icon, builtin, position)
			SELECT ?1, CASE WHEN EXISTS (SELECT 1 FROM labels WHERE lower(name) = ?1) THEN ?1 || ' (imap)' ELSE ?1 END,
				'gray', 'folder-3', 0, (SELECT COALESCE(MAX(position) + 1, 0) FROM labels)
			WHERE NOT EXISTS (SELECT 1 FROM labels WHERE slug = ?1)`, Args: []any{slug}})
	}
	for id, kws := range byMsg {
		for _, k := range kws {
			stmts = append(stmts, sqlite.Statement{SQL: `INSERT INTO message_labels(message_id, label_id, source, active)
				SELECT ?, id, 'imap', 1 FROM labels WHERE slug = ?
				ON CONFLICT(message_id, label_id) DO UPDATE SET source = 'imap', active = 1 WHERE source = 'auto' OR source = 'imap'`,
				Args: []any{id, k}})
		}
	}
	// A keyword gone from the server takes its imap label off; hand-made rows stay.
	stmts = append(stmts, sqlite.Statement{SQL: `UPDATE message_labels SET active = 0
		WHERE source = 'imap' AND active = 1
		AND message_id IN (SELECT id FROM messages WHERE account_id = ?)
		AND NOT EXISTS (SELECT 1 FROM messages m JOIN labels l ON l.id = message_labels.label_id
			WHERE m.id = message_labels.message_id AND instr(' ' || lower(m.flags) || ' ', ' ' || l.slug || ' ') > 0)`,
		Args: []any{accountID}})
	if _, err := s.db.Transaction(ctx, stmts); err != nil {
		return err
	}
	s.bus.Publish(events.Event{Type: events.LabelsChanged})
	return nil
}

// pushLabelKeyword queues the label's slug as an IMAP keyword on (or off) the
// messages, so other clients and agents see a tag put on by hand. A slug a
// server cannot hold as a keyword is left local.
func (s *Service) pushLabelKeyword(ctx context.Context, ids []int64, labelID int64, on bool) error {
	rows, err := s.db.Query(ctx, "SELECT slug FROM labels WHERE id = ?", labelID)
	if err != nil || len(rows.Values) == 0 {
		return err
	}
	slug := store.Str(rows.Values[0][0])
	if !keywordSlug.MatchString(slug) {
		return nil
	}
	if on {
		return s.SetFlags(ctx, ids, []string{slug}, nil)
	}
	return s.SetFlags(ctx, ids, nil, []string{slug})
}
