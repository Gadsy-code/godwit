package core

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

// A label may show an uploaded image (a client's favicon) instead of a
// glyph from LabelIcons, the way an account does (accounts.icon = image:<n>).

var tinyPNG, _ = base64.StdEncoding.DecodeString(
	"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==")

func TestLabelImageIcon(t *testing.T) {
	s := newLabelService(t, "", "en")
	ctx := t.Context()

	l, err := s.CreateLabel(ctx, LabelInput{Name: "Acme", Color: "blue", Icon: "briefcase-4", IconImage: tinyPNG})
	require.NoError(t, err)
	require.Contains(t, l.Icon, "image:", "an uploaded image replaces the glyph name")

	img, err := s.LabelIcon(ctx, l.ID)
	require.NoError(t, err)
	require.Equal(t, "image/png", img.ContentType)
	require.Equal(t, tinyPNG, img.Data)

	// Saving the label again without a new image keeps the one it has.
	l2, err := s.UpdateLabel(ctx, l.ID, LabelInput{Name: "Acme srl", Color: "blue", Icon: l.Icon})
	require.NoError(t, err)
	require.Equal(t, l.Icon, l2.Icon)

	// Picking a glyph again drops the image.
	l3, err := s.UpdateLabel(ctx, l.ID, LabelInput{Name: "Acme srl", Color: "blue", Icon: "home-4"})
	require.NoError(t, err)
	require.Equal(t, "home-4", l3.Icon)
	_, err = s.LabelIcon(ctx, l.ID)
	require.Error(t, err)

	// Not an image: refused.
	_, err = s.UpdateLabel(ctx, l.ID, LabelInput{Name: "Acme srl", Color: "blue", Icon: "home-4", IconImage: []byte("hello")})
	require.ErrorIs(t, err, ErrAccountIcon)
}

// A label can also show its initial in its color, with nothing stored.
func TestLabelMonogramIcon(t *testing.T) {
	s := newLabelService(t, "", "en")
	ctx := t.Context()

	l, err := s.CreateLabel(ctx, LabelInput{Name: "Acme", Color: "blue", Icon: LabelIconMonogram})
	require.NoError(t, err)
	require.Equal(t, LabelIconMonogram, l.Icon)

	l, err = s.UpdateLabel(ctx, l.ID, LabelInput{Name: "Acme", Color: "blue", Icon: "home-4"})
	require.NoError(t, err)
	require.Equal(t, "home-4", l.Icon)

	_, err = s.CreateLabel(ctx, LabelInput{Name: "Other", Color: "blue", Icon: "not-an-icon"})
	require.Error(t, err)
}
