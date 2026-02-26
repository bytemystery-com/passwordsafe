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
	"fmt"

	"bytemystery-com/passwordsafe/database"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/bytemystery-com/picbutton"
)

type CategoryView struct {
	data             []*database.CategoryDataType
	entryData        []*database.EntryDataType
	list             *widget.Select
	content          *fyne.Container
	entryList        *widget.List
	isSearchView     bool
	lastSearch       string
	selIndexFromProg bool
}

var _ UpdateToolbarInterface = (*CategoryView)(nil)

func NewCategoryView() *CategoryView {
	v := CategoryView{}
	v.list = widget.NewSelect(nil, func(string) {
		index := v.list.SelectedIndex()
		if index < 0 {
			return
		}
		if v.data == nil {
			return
		}
		Gui.Settings.LastCategoryId = int(v.data[index].Id)
		Gui.Settings.Store()
		v.isSearchView = false
		v.UdateEntryList()
		if !v.selIndexFromProg {
			v.entryList.ScrollToTop()
		}
	})
	v.entryList = widget.NewList(v.listGetLength, v.listCreateObject, v.listUpdateObject)
	v.entryList.OnSelected = v.listSelected

	addEntry := picbutton.NewPicButton(Gui.IconPlusUp.StaticContent, Gui.IconPlusDown.StaticContent, nil, nil, false, AddEntry, nil)
	search := picbutton.NewPicButton(Gui.IconSearchUp.StaticContent, Gui.IconSearchDown.StaticContent, nil, nil, false, v.DoSearch, nil)
	lock := picbutton.NewPicButton(Gui.IconLockUp.StaticContent, Gui.IconLockDown.StaticContent, nil, nil, false, DoLock, nil)
	buttonLine := container.NewHBox(addEntry, layout.NewSpacer(), search, layout.NewSpacer(), lock)
	v.content = container.NewBorder(nil, nil, nil, nil, container.NewBorder(v.list, buttonLine, nil, nil, v.entryList))
	return &v
}

func (v *CategoryView) doSearch(search string) {
	data, err := Database.Search("%" + search + "%")
	if err != nil {
		UIErrorHandler(err)
		return
	}
	v.isSearchView = true
	v.list.ClearSelected()
	v.entryData = data
	v.entryList.Refresh()
}

func (v *CategoryView) DoSearch() {
	var dia *dialog.ConfirmDialog
	text := widget.NewEntry()
	text.OnSubmitted = func(string) {
		dia.Confirm()
	}
	text.SetPlaceHolder(lang.X("search.name.placeholder", "Search text"))
	c := container.NewBorder(widget.NewLabel(lang.X("search.name", "Search text")), nil, nil, nil, text)
	dia = dialog.NewCustomConfirm(lang.X("search.title", "Search entries"), lang.X("search", "Search"), lang.X("cancel", "Cancel"),
		c, func(ok bool) {
			if !ok {
				return
			}
			v.lastSearch = text.Text
			v.doSearch(v.lastSearch)
		}, Gui.MainWindow,
	)
	dia.Show()
	si := Gui.MainWindow.Canvas().Size()
	var windowScale float32 = 1.0
	dia.Resize(fyne.NewSize(si.Width*windowScale, dia.MinSize().Height))

	Gui.MainWindow.Canvas().Focus(text)
}

func (v *CategoryView) GetCategoryId() (int64, error) {
	index := v.list.SelectedIndex()
	if index < 0 {
		return -1, errors.New("no category selected")
	}
	return v.data[index].Id, nil
}

func (v *CategoryView) ShowCategoryDialog(str string, f func(string)) {
	var dia *dialog.ConfirmDialog
	label := widget.NewLabel(lang.X("category.dialog.name", "Category name"))
	name := widget.NewEntry()
	name.SetText(str)
	name.OnSubmitted = func(string) {
		dia.Confirm()
	}
	name.SetPlaceHolder(lang.X("category.dialog.name.placeholder", "Name of the category"))
	c := container.NewVBox(label, name)
	dia = dialog.NewCustomConfirm(lang.X("category.dialog.title", "Category"),
		lang.X("ok", "Ok"), lang.X("cancel", "Cancel"), c, func(ok bool) {
			if ok {
				if name.Text == "" {
					UIErrorHandler(errors.New(lang.X("category.dialog.name.empty", "Category name may not be empty !")))
					return
				}
				f(name.Text)
			}
		}, Gui.MainWindow)
	dia.Show()
	si := Gui.MainWindow.Canvas().Size()
	dia.Resize(fyne.NewSize(si.Width, dia.MinSize().Height))
	Gui.MainWindow.Canvas().Focus(name)
}

func (v *CategoryView) Add() {
	v.ShowCategoryDialog("", func(name string) {
		data := database.CategoryDataType{
			Name: name,
			Id:   -1,
		}
		err := Database.AddCategory(&data)
		if err != nil {
			UIErrorHandler(err)
			return
		}
		v.UpdateCategoryList(data.Id)
		v.list.SetSelected(name)
	})
}

func (v *CategoryView) Edit() {
	index := v.list.SelectedIndex()
	if index < 0 {
		return
	}
	data := v.data[index]
	v.ShowCategoryDialog(data.Name, func(name string) {
		data.Name = name
		err := Database.UpdateCategory(data)
		if err != nil {
			UIErrorHandler(err)
			return
		}
		v.UpdateCategoryList(-1)
	})
}

func (v *CategoryView) Delete() {
	index := v.list.SelectedIndex()
	if index < 0 {
		return
	}
	data := v.data[index]
	count, err := Database.GetEntriesInCategory(data.Id)
	if err != nil {
		UIErrorHandler(err)
		return
	}
	dialog.ShowConfirm(lang.X("category.delete.title", "Delete category"),
		fmt.Sprintf(lang.X("category.delete.msg", "Do you really want to delete the category\n'%s'\nwith %d entries ?"), data.Name, count),
		func(ok bool) {
			if ok {
				err := Database.DeleteCategory(data.Id)
				if err != nil {
					UIErrorHandler(err)
					return
				}
				v.UpdateCategoryList(-1)
			}
		}, Gui.MainWindow)
}

func (v *CategoryView) UpdateCategoryList(categoryId int64) {
	list, err := Database.GetCategories()
	if err != nil {
		UIErrorHandler(err)
		return
	}
	selIndex := v.list.SelectedIndex()
	var id int64 = -1
	if v.data != nil && selIndex >= 0 {
		id = v.data[selIndex].Id
	}
	v.data = list
	l := make([]string, 0, len(v.data))
	selIndex = -1
	if id == -1 || categoryId >= 0 {
		id = categoryId
	}
	for index, item := range v.data {
		l = append(l, item.Name)
		if item.Id == id {
			selIndex = index
		}
	}
	v.list.ClearSelected()
	v.list.SetOptions(l)
	if selIndex < 0 {
		selIndex = 0
	}
	if selIndex < len(v.data) {
		v.selIndexFromProg = true
		v.list.SetSelectedIndex(selIndex)
		v.selIndexFromProg = false
	}
}

func (v *CategoryView) listSelected(id widget.ListItemID) {
	SetEntryView(v.entryData[id].Id)
}

func (v *CategoryView) listGetLength() int {
	return len(v.entryData)
}

func (v *CategoryView) listCreateObject() fyne.CanvasObject {
	text := canvas.NewText("", theme.Color(theme.ColorNameForeground))
	text.TextStyle = fyne.TextStyle{}
	text.TextSize = theme.Size(theme.SizeNameText) * Gui.Theme.GetSpecialSize("category_view_text_scale")
	text.Refresh()
	icon := canvas.NewImageFromResource(theme.QuestionIcon())
	is := Gui.Theme.GetSpecialSize("category_view_icon_size")
	icon.SetMinSize(fyne.NewSize(is, is))
	icon.FillMode = canvas.ImageFillContain
	icon.Refresh()
	content := container.NewBorder(nil, nil, icon, nil, text)

	return content
}

func (v *CategoryView) listUpdateObject(id widget.ListItemID, o fyne.CanvasObject) {
	c, ok := o.(*fyne.Container)
	if !ok {
		return
	}

	text, ok := c.Objects[0].(*canvas.Text)
	if !ok {
		return
	}

	icon, ok := c.Objects[1].(*canvas.Image)
	if !ok {
		return
	}

	data := v.entryData[id]
	if v.isSearchView {
		text.Text = fmt.Sprintf("%s  (%s)", data.Name, v.getCategoryName(data.CategoryId))
	} else {
		text.Text = data.Name
	}
	text.Color = theme.Color(theme.ColorNameForeground)
	text.Refresh()
	ico, ok := Gui.IconCollection.GetByKey(v.entryData[id].Icon)
	if ok {
		icon.Resource = ico
	} else {
		icon.Resource = GetDefaultIcon()
	}
	is := Gui.Theme.GetSpecialSize("category_view_icon_size")
	icon.SetMinSize(fyne.NewSize(is, is))
	icon.FillMode = canvas.ImageFillContain
	icon.Refresh()
}

func (v *CategoryView) getCategoryName(categoryID int64) string {
	for _, item := range v.data {
		if item.Id == categoryID {
			return item.Name
		}
	}
	return "???"
}

func (v *CategoryView) UdateEntryList() {
	categoryId, err := v.GetCategoryId()
	if err != nil {
		return
	}
	list, err := Database.GetEntries(categoryId)
	if err != nil {
		UIErrorHandler(err)
		return
	}
	v.entryData = list
	v.entryList.Refresh()
}

func (v *CategoryView) Update(categoryId int64) {
	if !v.isSearchView {
		v.UpdateCategoryList(categoryId)
		v.UdateEntryList()
	} else {
		v.doSearch(v.lastSearch)
	}
	v.entryList.UnselectAll()
}

func (v *CategoryView) GetContent() *fyne.Container {
	return v.content
}

func (v *CategoryView) UpdateToolBar() {
	Gui.Toolbar.Items = []widget.ToolbarItem{Gui.toolToggleTheme, widget.NewToolbarSeparator(), Gui.toolAdd, Gui.toolEdit, Gui.toolRemove, widget.NewToolbarSeparator(), Gui.toolChangePasswd, Gui.toolSettings, Gui.toolExport, Gui.toolImport, widget.NewToolbarSpacer(), Gui.toolInfo}
	Gui.Toolbar.Refresh()
}

func (v *CategoryView) ThemeChanged() {
	v.content.Refresh()
}
