// Copyright (c) 2026 Reiner Pröls
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// SPDX-License-Identifier: MIT
//
// Author: Reiner Pröls

package entrylayout

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type LabelEntryIconLayout struct{}

func (l *LabelEntryIconLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) != 3 {
		return
	}

	label := objects[0]
	entry := objects[1]
	icon := objects[2]

	// Berechne Label + Entry Höhe
	labelHeight := label.MinSize().Height
	entryHeight := entry.MinSize().Height
	formWidth := size.Width - icon.MinSize().Width - theme.Padding()

	// Position Label + Entry links
	label.Move(fyne.NewPos(0, 0))
	label.Resize(fyne.NewSize(formWidth, labelHeight))

	entry.Move(fyne.NewPos(0, labelHeight))
	entry.Resize(fyne.NewSize(formWidth, entryHeight))

	// Icon unten rechts
	iconWidth := icon.MinSize().Width
	iconHeight := icon.MinSize().Height
	if iconHeight > labelHeight+entryHeight {
		iconHeight = labelHeight + entryHeight
		iconWidth = labelHeight + entryHeight
	}
	y := iconHeight/2 - entryHeight/2
	icon.Move(fyne.NewPos(size.Width-iconWidth, labelHeight-y))
	icon.Resize(fyne.NewSize(iconWidth, iconHeight))
}

func (l *LabelEntryIconLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) != 3 {
		return fyne.NewSize(0, 0)
	}

	label := objects[0]
	entry := objects[1]
	icon := objects[2]

	se := entry.MinSize()
	w := label.MinSize().Width
	if se.Width > w {
		w = se.Width
	}
	h := label.MinSize().Height + se.Height + icon.MinSize().Height/2 - se.Height/2
	wIcon := icon.MinSize().Width

	return fyne.NewSize(w+wIcon, h)
}

type LabelValueLayout struct{}

func (l *LabelValueLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) != 2 {
		return
	}

	label := objects[0]
	value := objects[1]

	// Berechne Label + Entry Höhe
	labelHeight := label.MinSize().Height
	valueHeight := value.MinSize().Height

	// Position Label + Entry links
	label.Move(fyne.NewPos(0, 0))
	label.Resize(fyne.NewSize(size.Width, labelHeight))

	value.Move(fyne.NewPos(0, labelHeight*.7))
	value.Resize(fyne.NewSize(size.Width, valueHeight))
}

func (l *LabelValueLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) != 2 {
		return fyne.NewSize(0, 0)
	}

	label := objects[0]
	value := objects[1]

	w := label.MinSize().Width
	if value.MinSize().Width > w {
		w = value.MinSize().Width
	}
	h := label.MinSize().Height*.7 + value.MinSize().Height

	return fyne.NewSize(w, h)
}

type EntryTitleLayout struct{}

func (l *EntryTitleLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) != 3 {
		return
	}

	icon := objects[0]
	title := objects[1]
	date := objects[2]
	pad := theme.Padding()

	icon.Move(fyne.NewPos(0, pad))
	icon.Resize(fyne.NewSize(icon.MinSize().Width, icon.MinSize().Height))

	title.Move(fyne.NewPos(icon.MinSize().Width+pad, 0))
	title.Resize(fyne.NewSize(size.Width-icon.MinSize().Width-pad, title.MinSize().Height))

	date.Move(fyne.NewPos(size.Width-date.MinSize().Width, title.MinSize().Height*.8))
	date.Resize(date.MinSize())
}

func (l *EntryTitleLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) != 3 {
		return fyne.NewSize(0, 0)
	}

	icon := objects[0]
	title := objects[1]
	date := objects[2]
	pad := theme.Padding()

	w := title.MinSize().Width + icon.MinSize().Width + pad
	h := title.MinSize().Height*.8 + date.MinSize().Height
	if h < icon.MinSize().Height+pad {
		h = icon.MinSize().Height + pad
	}

	return fyne.NewSize(w, h)
}
