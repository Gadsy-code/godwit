package app

import (
	"slices"
	"testing"

	"github.com/godwit-email/godwit/core"
	"github.com/stretchr/testify/assert"
)

// "Group · Name" nests a label under Group in the sidebar, the way Gmail nests "Group/Name".
func TestGroupLabels(t *testing.T) {
	ls := []core.Label{
		{ID: 1, Name: "Codes"},
		{ID: 2, Name: "Client · Acme"},
		{ID: 3, Name: "Client · Beta"},
		{ID: 4, Name: "Type · Invoice"},
		{ID: 5, Name: "Notes"},
		{ID: 6, Name: "Client · Gamma"}, // joins its group where that began
		{ID: 7, Name: " · broken"},      // no group name: stays alone, whole
		{ID: 8, Name: "broken · "},      // no name of its own: the same
		{ID: 9, Name: "A · B · C"},      // the group ends at the first separator
	}
	type row struct {
		group string
		names []string
	}
	var got []row
	for _, e := range groupLabels(ls) {
		r := row{group: e.Group}
		for _, l := range e.Labels {
			r.names = append(r.names, l.Short)
		}
		got = append(got, r)
	}
	assert.Equal(t, []row{
		{"", []string{"Codes"}},
		{"Client", []string{"Acme", "Beta", "Gamma"}},
		{"Type", []string{"Invoice"}},
		{"", []string{"Notes"}},
		{"", []string{" · broken"}},
		{"", []string{"broken · "}},
		{"A", []string{"B · C"}},
	}, got)
}

// The sidebar shows the group as a title and its labels by their own names.
func TestSidebarNestsLabels(t *testing.T) {
	h, _, m := labeledMail(t)
	ctx := t.Context()
	l, err := h.env.Svc.CreateLabel(ctx, core.LabelInput{Name: "Client · Acme", Color: "blue", Icon: "folder-3"})
	if err != nil {
		t.Fatal(err)
	}
	if err := h.env.Svc.SetLabel(ctx, []int64{m.ID}, l.ID, true); err != nil {
		t.Fatal(err)
	}
	h.waitMail("the group and the label in the sidebar", func() bool {
		return slices.Contains(h.tt.Texts(), "Client") && slices.Contains(h.tt.Texts(), "Acme")
	})
}
