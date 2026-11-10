package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/godwit-email/godwit/core"
	"github.com/godwit-email/godwit/internal/icons"
	"github.com/godwit-email/godwit/internal/kit"
	"github.com/godwit-email/godwit/internal/theme"
)

// A label may show an uploaded image (a client's logo) instead of a glyph.

// labelImage is the decoded image of a label whose Icon is image:<n>, nil while
// it loads or when it does not decode. One per label: a new upload has a new key.
func (a *App) labelImage(l core.Label) ui.ImageSource {
	if !strings.HasPrefix(l.Icon, "image:") {
		return nil
	}
	if e, ok := a.labelImages[l.ID]; ok && e.key == l.Icon {
		return e.img
	}
	if a.labelImages == nil {
		a.labelImages = map[int64]accountImage{}
	}
	a.labelImages[l.ID] = accountImage{key: l.Icon}
	var img ui.ImageSource
	a.do(func(ctx context.Context) error {
		icon, err := a.svc.LabelIcon(ctx, l.ID)
		if err == nil {
			img = decodeIcon(icon.ContentType, icon.Data)
		}
		return err
	}, func(error) {
		if e := a.labelImages[l.ID]; e.key == l.Icon {
			a.labelImages[l.ID] = accountImage{key: l.Icon, img: img}
		}
	})
	return nil
}

// labelGlyph draws a label's image, or its glyph in its color.
func (a *App) labelGlyph(c *ui.Context, l core.Label, size float32) {
	if l.Icon == core.LabelIconMonogram || strings.HasPrefix(l.Icon, "image:") {
		if img := a.labelImage(l); img != nil || l.Icon == core.LabelIconMonogram {
			kit.LabelBadge(c, kit.LabelBadgeOpts{Name: l.Name, Size: size, Color: theme.Of(c).Label(l.Color), Image: img}).Margin(0, c.Theme().Space(0.5), 0, 0)
		} else {
			ui.Row(c).Size(size, size) // keeps its place while it loads
		}
		return
	}
	ui.Icon(c, icons.Get("label-"+l.Icon)).Size(size, size).TextColor(theme.Of(c).Label(l.Color).Text)
}

// pickLabelImage asks for an image off the UI thread and selects it in the label
// dialog, if it still shows.
func (v *labelsView) pickLabelImage() {
	a := v.a
	var data []byte
	var img ui.ImageSource
	a.do(func(context.Context) error {
		path, err := a.plat.OpenImage()
		if err != nil || path == "" {
			return err
		}
		if fi, err := os.Stat(path); err != nil {
			return err
		} else if fi.Size() > core.AccountIconMax {
			return core.ErrAccountIcon
		}
		if data, err = os.ReadFile(path); err != nil {
			return err
		}
		ct := "image/png" // any bitmap: DecodeBitmap sniffs
		if strings.EqualFold(filepath.Ext(path), ".svg") {
			ct = "image/svg+xml"
		}
		if img = decodeIcon(ct, data); img == nil {
			return core.ErrAccountIcon
		}
		return nil
	}, func(err error) {
		if err != nil || img == nil || !v.dlg {
			a.toastErr(err)
			return
		}
		v.image, v.imagePreview = data, img
	})
}

// dialogImage is the image the label dialog shows: the one just picked, else the
// label's own, else nil (a glyph).
func (v *labelsView) dialogImage() ui.ImageSource {
	if v.imagePreview != nil {
		return v.imagePreview
	}
	if strings.HasPrefix(v.icon, "image:") {
		return v.a.labelImage(v.editLabel)
	}
	return nil
}
