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
	"errors"
	"strings"

	"bytemystery-com/passwordsafe/database"
	"bytemystery-com/passwordsafe/entrylayout"
	"bytemystery-com/passwordsafe/menubutton"
	"bytemystery-com/passwordsafe/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/bytemystery-com/picbutton"
)

type EntryWidget struct {
	value *widget.Entry
	label *widget.Label
}

type EntryEditView struct {
	data          *database.EntryFullDataType
	content       *fyne.Container
	vBox          *fyne.Container
	icon          *picbutton.PicButton
	name          *widget.Entry
	entryWidgets  []*EntryWidget
	diaSelectIcon *dialog.CustomDialog
	editMode      bool
	categoryList  *widget.Select
	categoryData  []*database.CategoryDataType
	scroll        *container.Scroll
}

var _ UpdateToolbarInterface = (*EntryEditView)(nil)

func NewEntryEditView() *EntryEditView {
	e := EntryEditView{
		data: &database.EntryFullDataType{},
	}

	label := widget.NewLabel(lang.X("entry.name", "Entry"))
	e.name = widget.NewEntry()
	e.name.SetText(e.data.Name)
	// name.SetPlaceHolder(lang.X("entry.dialog.name.placeholder", "Name of the entry"))
	icon := GetDefaultIcon()
	e.icon = picbutton.NewPicButton(icon.StaticContent, icon.StaticContent, nil, nil, false, func() {
		e.showIconSelectDialog()
	}, nil)
	si := Gui.Theme.GetSpecialSize("entry_edit_icon_size")
	e.icon.SetMinSize(fyne.NewSize(si, si))
	e.categoryList = widget.NewSelect([]string{}, func(string) {
		index := e.categoryList.SelectedIndex()
		if index >= 0 {
			e.data.CategoryId = e.categoryData[index].Id
		}
	})
	baseItem1 := container.New(&entrylayout.LabelEntryIconLayout{}, label, e.name, e.icon)
	baseItem := container.NewVBox(baseItem1, widget.NewLabel(lang.X("entry.category", "Category")), e.categoryList)

	e.vBox = container.NewVBox()
	e.scroll = container.NewScroll(e.vBox)
	for index, entry := range e.data.Fields {
		e.newFieldContent(index, entry)
	}
	addField := picbutton.NewPicButton(Gui.IconPlusUp.StaticContent, Gui.IconPlusDown.StaticContent, nil, nil, false, func() { e.addField(-1) }, nil)
	cancel := widget.NewButton(lang.X("cancel", "Cancel"), e.doCancel)
	ok := widget.NewButton(lang.X("save", "Save"), e.doOk)
	ok.Importance = widget.HighImportance
	wOk := util.GetDefaultTextWidth(ok.Text + "XXX")
	wCancel := util.GetDefaultTextWidth(cancel.Text + "XXX")
	w := wCancel
	if wOk > wCancel {
		w = wCancel
	}
	btnSize := fyne.NewSize(w, ok.MinSize().Height)
	okC := container.NewGridWrap(btnSize, ok)
	cancelC := container.NewGridWrap(btnSize, cancel)
	lock := picbutton.NewPicButton(Gui.IconLockUp.StaticContent, Gui.IconLockDown.StaticContent, nil, nil, false, DoLock, nil)
	buttonLine := container.NewHBox(container.NewCenter(addField), layout.NewSpacer(),
		container.NewCenter(cancelC), container.NewCenter(okC), layout.NewSpacer(), container.NewCenter(lock))
	e.content = container.NewBorder(baseItem, buttonLine, nil, nil, e.scroll)

	// Icon select
	si = Gui.Theme.GetSpecialSize("icon_select_icon_size")
	grid := container.NewGridWrap(fyne.NewSize(si+theme.Padding()*2, si+theme.Padding()*2))
	for item := range Gui.IconCollection.IterValue() {
		b := picbutton.NewPicButton(item.StaticContent, item.StaticContent, nil, nil, false, func() {
			e.data.Icon = item.StaticName
			e.icon.SetUImg(item.StaticContent)
			e.icon.SetDImg(item.StaticContent)
			e.diaSelectIcon.Hide()
		}, nil)
		b.SetHooverImg(Gui.IconHooverBottom.StaticContent, nil)
		grid.Add(b)
	}
	scroll := container.NewScroll(grid)
	cancelb := container.NewHBox(layout.NewSpacer(), widget.NewButton(lang.X("cancel", "Cancel"), func() {
		e.diaSelectIcon.Hide()
	}), layout.NewSpacer())

	e.diaSelectIcon = dialog.NewCustomWithoutButtons(lang.X("entry.new.icon", "Icon"),
		container.NewBorder(nil, cancelb, nil, nil, scroll), Gui.MainWindow)

	return &e
}

func (e *EntryEditView) newFieldContent(index int, entry *database.EntryFieldData) {
	label := widget.NewLabel(entry.Name)
	value := widget.NewEntry()
	value.SetText(strings.ReplaceAll(entry.Value, "\n", "\\n"))
	e.entryWidgets = append(e.entryWidgets, &EntryWidget{
		label: label,
		value: value,
	})
	var icon *menubutton.MenuButton
	icon = menubutton.NewMenuButton("", theme.MoreHorizontalIcon(), func(ev *fyne.PointEvent) {
		popupMenu := fyne.NewMenu("")
		if index > 0 {
			popupMenu.Items = append(popupMenu.Items, fyne.NewMenuItemWithIcon(lang.X("menu.up", "Move up"), theme.MoveUpIcon(), func() {
				e.saveAllInData()
				e.data.Fields[index-1], e.data.Fields[index] = e.data.Fields[index], e.data.Fields[index-1]
				e.doUpdate()
			}))
		}
		popupMenu.Items = append(popupMenu.Items, fyne.NewMenuItemWithIcon(lang.X("menu.delete", "Delete"), theme.ContentRemoveIcon(), func() {
			e.saveAllInData()
			e.data.Fields = append(e.data.Fields[0:index], e.data.Fields[index+1:]...)
			e.doUpdate()
		}))
		popupMenu.Items = append(popupMenu.Items, fyne.NewMenuItemWithIcon(lang.X("menu.edit", "Edit"), theme.DocumentCreateIcon(), func() {
			e.addField(index)
		}))
		if index < len(e.data.Fields)-1 {
			popupMenu.Items = append(popupMenu.Items, fyne.NewMenuItemWithIcon(lang.X("menu.down", "Move down"), theme.MoveDownIcon(), func() {
				e.saveAllInData()
				e.data.Fields[index], e.data.Fields[index+1] = e.data.Fields[index+1], e.data.Fields[index]
				e.doUpdate()
			}))
		}
		widget.ShowPopUpMenuAtPosition(popupMenu, Gui.App.Driver().CanvasForObject(icon), ev.AbsolutePosition)
	})
	icon.Importance = widget.LowImportance
	s := fyne.NewSize(e.icon.MinSize().Width, icon.MinSize().Height)
	item := container.New(&entrylayout.LabelEntryIconLayout{}, label, value, container.NewGridWrap(s, icon))
	e.vBox.Add(item)
	Gui.MainWindow.Canvas().Focus(value)
}

func (e *EntryEditView) addField(index int) {
	f := database.EntryFieldData{}
	var dia *dialog.ConfirmDialog
	label := widget.NewLabel(lang.X("entry.new", "Label name"))
	name := widget.NewEntry()
	name.SetPlaceHolder(lang.X("entry.new.label.placeholder", "Label name for the new entry"))
	name.OnSubmitted = func(string) {
		dia.Confirm()
	}
	pass := widget.NewCheck(lang.X("entry.ispass", "Is password field"), nil)
	if index >= 0 {
		fi := e.data.Fields[index]
		name.SetText(fi.Name)
		pass.Checked = fi.IsPassword
	}
	c := container.NewVBox(label, name, pass)
	var t string
	if index >= 0 {
		t = lang.X("entry.new.titel", "New field")
	} else {
		t = lang.X("entry.edit.titel", "Edit field")
	}
	dia = dialog.NewCustomConfirm(t, lang.X("ok", "Ok"), lang.X("cancel", "Cancel"),
		c, func(ok bool) {
			if index >= 0 {
				fi := e.data.Fields[index]
				fi.Name = name.Text
				fi.IsPassword = pass.Checked
				e.entryWidgets[index].label.SetText(fi.Name)
			} else {
				f.Name = name.Text
				f.IsPassword = pass.Checked
				e.data.Fields = append(e.data.Fields, &f)
				e.newFieldContent(len(e.data.Fields)-1, &f)
			}
		}, Gui.MainWindow)
	dia.Show()
	si := Gui.MainWindow.Canvas().Size()
	dia.Resize(fyne.NewSize(si.Width, dia.MinSize().Height))
	Gui.MainWindow.Canvas().Focus(name)
}

func (e *EntryEditView) ResetIcon() {
	icon := GetDefaultIcon()
	if icon != nil {
		e.icon.SetUImg(icon.StaticContent)
		e.icon.SetDImg(icon.StaticContent)
	}
}

func (e *EntryEditView) SetEntry(id int64) {
	d, err := Database.GetEntry(id, MasterPasswort, Crypt)
	if err != nil {
		UIErrorHandler(err)
		return
	}
	e.data = d
	e.initCategoryData()
	e.Update()
	e.scroll.ScrollToTop()
}

func (e *EntryEditView) doUpdate() {
	e.name.SetText(e.data.Name)
	ico, ok := Gui.IconCollection.GetByKey(e.data.Icon)
	if ok {
		e.icon.SetUImg(ico.StaticContent)
		e.icon.SetDImg(ico.StaticContent)
	}
	e.entryWidgets = make([]*EntryWidget, 0, len(e.data.Fields))
	e.vBox.RemoveAll()
	for index, item := range e.data.Fields {
		e.newFieldContent(index, item)
	}
}

func (e *EntryEditView) Update() {
	e.doUpdate()
	e.editMode = true
}

func (e *EntryEditView) initCategoryData() {
	e.categoryList.ClearSelected()
	d, err := Database.GetCategories()
	if err == nil {
		e.categoryData = d
		list := make([]string, 0, len(e.categoryData))
		selIndex := -1
		for index, item := range e.categoryData {
			if e.data.CategoryId == item.Id {
				selIndex = index
			}
			list = append(list, item.Name)
		}
		e.categoryList.SetOptions(list)
		if selIndex < 0 {
			selIndex = 0
		}
		if len(e.categoryData) > 0 {
			e.categoryList.SetSelectedIndex(selIndex)
		}
	} else {
		UIErrorHandler(err)
	}
}

func (e *EntryEditView) NewEntry(categoryId int64) {
	e.data = &database.EntryFullDataType{}
	e.name.SetText("")
	e.entryWidgets = make([]*EntryWidget, 0, 5)
	e.vBox.RemoveAll()
	e.ResetIcon()
	e.data.CategoryId = categoryId
	e.editMode = false
	e.initCategoryData()
	e.scroll.ScrollToTop()
}

func (e *EntryEditView) doCancel() {
	if e.editMode == true {
		switch Gui.Settings.ViewModeAfterEdit {
		case VIEWMODE_CATEGORY:
			SetCatgeoryView(-1)
		default:
			SetEntryView(e.data.Id)
		}
	} else {
		SetCatgeoryView(-1)
	}
}

func (e *EntryEditView) saveAllInData() {
	e.data.Name = e.name.Text

	for index, item := range e.data.Fields {
		item.Name = e.entryWidgets[index].label.Text
		item.Value = strings.ReplaceAll(e.entryWidgets[index].value.Text, "\\n", "\n")
	}
}

func (e *EntryEditView) doOk() {
	e.saveAllInData()
	if e.data.Name == "" {
		UIErrorHandler(errors.New(lang.X("edit.empty_name", "The entry name may not be empty !")))
		return
	}
	if e.data.CategoryId == -1 {
		UIErrorHandler(errors.New(lang.X("edit.no_ctegory", "Category must be set !")))
		return
	}
	SetBusy(true)
	go func() {
		defer SetBusy(false)
		if e.editMode {
			err := Database.UpdateEntry(e.data, MasterPasswort, Crypt)
			if err != nil {
				UIErrorHandler(err)
				return
			}
		} else {
			id, err := Database.AddEntry(e.data, MasterPasswort, Crypt)
			if err != nil {
				UIErrorHandler(err)
				return
			}
			e.data.Id = id
		}
		fyne.Do(func() {
			if e.editMode == true {
				switch Gui.Settings.ViewModeAfterEdit {
				case VIEWMODE_CATEGORY:
					SetCatgeoryView(e.data.CategoryId)
				default:
					SetEntryView(e.data.Id)
				}
			} else {
				switch Gui.Settings.ViewModeAfterAdd {
				case VIEWMODE_CATEGORY:
					SetCatgeoryView(e.data.CategoryId)
				default:
					SetEntryView(e.data.Id)
				}
			}
		})
	}()
}

func (e *EntryEditView) showIconSelectDialog() {
	e.diaSelectIcon.Show()
	si := Gui.MainWindow.Canvas().Size()
	e.diaSelectIcon.Resize(fyne.NewSize(si.Width, si.Height))
}

func (e *EntryEditView) GetContent() *fyne.Container {
	return e.content
}

func (e *EntryEditView) UpdateToolBar() {
	Gui.Toolbar.Items = []widget.ToolbarItem{Gui.toolToggleTheme, widget.NewToolbarSpacer(), Gui.toolInfo}
	Gui.Toolbar.Refresh()
}

func (e *EntryEditView) ThemeChanged() {
	e.content.Refresh()
}
