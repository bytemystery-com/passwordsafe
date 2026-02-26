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

package main

import (
	"fmt"
	"strings"

	"bytemystery-com/passwordsafe/clicklabel"
	"bytemystery-com/passwordsafe/database"
	"bytemystery-com/passwordsafe/entrylayout"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/bytemystery-com/colorlabel"
	"github.com/bytemystery-com/picbutton"
)

type EntryViewWidget struct {
	value *clicklabel.ClickLabel
	label *colorlabel.ColorLabel
}

type EntryView struct {
	data         *database.EntryFullDataType
	content      *fyne.Container
	vBox         *fyne.Container
	name         *colorlabel.ColorLabel
	dateTime     *colorlabel.ColorLabel
	icon         *canvas.Image
	view         *picbutton.PicButton
	entryWidgets []*EntryViewWidget
	scroll       *container.Scroll
}

var _ UpdateToolbarInterface = (*EntryView)(nil)

func NewEntryView() *EntryView {
	e := EntryView{
		data: &database.EntryFullDataType{},
	}

	e.name = colorlabel.NewColorLabel("", Gui.Theme.GetSpecialColor("entry_title"), nil, Gui.Theme.GetSpecialSize("entry_view_title_scale"))
	e.name.SetTextStyle(&fyne.TextStyle{
		Bold: true,
	})
	e.name.SetTruncateMode(colorlabel.End)
	ico, ok := Gui.IconCollection.GetByKey(e.data.Icon)
	if !ok {
		ico = GetDefaultIcon()
	}
	e.icon = canvas.NewImageFromResource(ico)
	e.icon.FillMode = canvas.ImageFillCover
	si := Gui.Theme.GetSpecialSize("entry_view_icon_size")
	e.icon.SetMinSize(fyne.NewSize(si, si))
	e.icon.Refresh()

	e.dateTime = colorlabel.NewColorLabel("", nil, nil, Gui.Theme.GetSpecialSize("entry_view_datetime_scale"))
	e.dateTime.SetAlinment(fyne.TextAlignTrailing)
	baseItem := container.New(&entrylayout.EntryTitleLayout{}, e.icon, e.name, e.dateTime)

	e.vBox = container.NewVBox()
	e.scroll = container.NewScroll(e.vBox)
	for _, entry := range e.data.Fields {
		e.newFieldContent(entry)
	}
	back := picbutton.NewPicButton(Gui.IconBackUp.StaticContent, Gui.IconBackDown.StaticContent, nil, nil, false, func() {
		SetCatgeoryView(-1)
	}, nil)
	edit := picbutton.NewPicButton(Gui.IconEditUp.StaticContent, Gui.IconEditDown.StaticContent, nil, nil, false, e.doEdit, nil)
	e.view = picbutton.NewPicButton(Gui.IconViewUp.StaticContent, Gui.IconViewDown.StaticContent, nil, nil, true, e.doShowPass, nil)
	lock := picbutton.NewPicButton(Gui.IconLockUp.StaticContent, Gui.IconLockDown.StaticContent, nil, nil, false, DoLock, nil)
	buttonLine := container.NewHBox(container.NewCenter(back), layout.NewSpacer(), container.NewCenter(edit), layout.NewSpacer(),
		container.NewCenter(e.view), layout.NewSpacer(), container.NewCenter(lock))
	e.content = container.NewBorder(baseItem, buttonLine, nil, nil, e.scroll)
	return &e
}

func (e *EntryView) doEdit() {
	SetEditView(e.data.Id)
}

func (e *EntryView) MakePassText(str string) string {
	return strings.Repeat("*", len(str))
}

func (e *EntryView) doShowPass() {
	showPass := e.view.IsDown()
	for index, item := range e.data.Fields {
		if item.IsPassword {
			if showPass {
				e.entryWidgets[index].value.SetText(item.Value)
			} else {
				e.entryWidgets[index].value.SetText(e.MakePassText(item.Value))
			}
		}
	}
}

func (e *EntryView) newFieldContent(entry *database.EntryFieldData) {
	label := colorlabel.NewColorLabel(entry.Name, Gui.Theme.GetSpecialColor("entry_label"), nil, Gui.Theme.GetSpecialSize("entry_view_label_scale"))
	var value *clicklabel.ClickLabel

	if !entry.IsPassword || e.view.IsDown() {
		value = clicklabel.NewClickLabel(entry.Value)
	} else {
		value = clicklabel.NewClickLabel(e.MakePassText(entry.Value))
	}
	value.Wrapping = fyne.TextWrapWord
	value.TextStyle = fyne.TextStyle{
		Bold: true,
	}
	value.OnTappedSecondary = func() {
		if Gui.Settings.CopyToClipboard == COPYTOCLIPBOARD_SECONDARYTAPPED || Gui.Settings.CopyToClipboard == COPYTOCLIPBOARD_BOTH {
			Gui.App.Clipboard().SetContent(value.Text)
		}
	}
	value.OnTapped = func() {
		if Gui.Settings.CopyToClipboard == COPYTOCLIPBOARD_TAPPED || Gui.Settings.CopyToClipboard == COPYTOCLIPBOARD_BOTH {
			Gui.App.Clipboard().SetContent(value.Text)
		}
	}
	item := container.New(&entrylayout.LabelValueLayout{}, label, value)
	e.entryWidgets = append(e.entryWidgets, &EntryViewWidget{
		label: label,
		value: value,
	})
	e.vBox.Add(item)
}

func (e *EntryView) Update(id int64) {
	e.icon.Resource = Gui.IconEmpty
	SetBusy(true)
	go func() {
		defer SetBusy(false)
		d, err := Database.GetEntry(id, MasterPasswort, Crypt)
		if err != nil {
			UIErrorHandler(err)
			return
		}
		n, err := Database.GetCategoryName(d.CategoryId)
		if err != nil {
			UIErrorHandler(err)
		}
		fyne.Do(func() {
			e.data = d
			e.name.SetText(d.Name)
			e.dateTime.SetText(fmt.Sprintf("%s  -  %s", n, e.data.Timestamp.Local().Format("02.01.06 - 15:04")))
			icon, ok := Gui.IconCollection.GetByKey(e.data.Icon)
			if !ok {
				icon = GetDefaultIcon()
			}
			e.icon.Resource = icon
			e.icon.FillMode = canvas.ImageFillCover
			si := Gui.Theme.GetSpecialSize("entry_view_icon_size")
			e.icon.SetMinSize(fyne.NewSize(si, si))
			e.icon.Refresh()
			e.entryWidgets = make([]*EntryViewWidget, 0, len(e.data.Fields))
			e.view.SetDown(false)
			e.vBox.RemoveAll()
			for _, item := range e.data.Fields {
				e.newFieldContent(item)
			}
			e.scroll.ScrollToTop()
		})
	}()
}

func (e *EntryView) DelEntry() *fyne.Container {
	dialog.ShowConfirm(lang.X("entry.view.delete.title", "Delete entry"), fmt.Sprintf(lang.X("entry.view.delete.msg", "Really delete entry\n%s ?"), e.name.GetText()),
		func(ok bool) {
			if !ok {
				return
			}
			err := Database.DeleteEntry(e.data.Id)
			if err != nil {
				UIErrorHandler(err)
			}
			SetCatgeoryView(-1)
		}, Gui.MainWindow)
	return e.content
}

func (e *EntryView) GetContent() *fyne.Container {
	return e.content
}

func (e *EntryView) UpdateToolBar() {
	Gui.Toolbar.Items = []widget.ToolbarItem{Gui.toolToggleTheme, widget.NewToolbarSeparator(), Gui.toolDelEntry, widget.NewToolbarSeparator(), Gui.toolChangePasswd, widget.NewToolbarSeparator(), Gui.toolSettings, widget.NewToolbarSeparator(), Gui.toolExport, Gui.toolImport, widget.NewToolbarSpacer(), Gui.toolInfo}
	Gui.Toolbar.Refresh()
}

func (e *EntryView) ThemeChanged() {
	e.name.SetTextColor(Gui.Theme.GetSpecialColor("entry_title"))
	for _, item := range e.entryWidgets {
		item.label.SetTextColor(Gui.Theme.GetSpecialColor("entry_label"))
	}
	e.content.Refresh()
}
