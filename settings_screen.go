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
	"path"
	"strconv"

	"bytemystery-com/passwordsafe/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

type SettingsView struct {
	ok              *widget.Button
	cancel          *widget.Button
	content         *fyne.Container
	autoLogOutTime  *widget.Entry
	copytoclipboard *widget.Select
	autoExportPath  *widget.Entry
	exportMode      *widget.Select
	viewAfterEdit   *widget.Select
	viewAfterAdd    *widget.Select
	checkForUpdates *widget.Check
}

var _ UpdateToolbarInterface = (*SettingsView)(nil)

func NewSettingsView() *SettingsView {
	s := SettingsView{}
	s.autoLogOutTime = widget.NewEntry()
	s.autoLogOutTime.OnChanged = util.GetNumberFilter(s.autoLogOutTime, nil)
	s.copytoclipboard = widget.NewSelect([]string{
		lang.X("copytoclipboard.none", "None"),
		lang.X("copytoclipboard.tapped", "Tapped"),
		lang.X("copytoclipboard.secondarytapped", "Secondary tapped"),
		lang.X("copytoclipboard.both", "Both"),
	}, nil)
	s.autoExportPath = widget.NewEntry()
	s.autoExportPath.Wrapping = fyne.TextWrapWord

	ok := widget.NewButton(lang.X("ok", "Ok"), func() { s.doSave() })
	ok.Importance = widget.HighImportance
	cancel := widget.NewButton(lang.X("cancel", "Cancel"), func() { s.doCancel() })

	wOk := util.GetDefaultTextWidth(ok.Text + "XXX")
	wCancel := util.GetDefaultTextWidth(cancel.Text + "XXX")
	w := wCancel
	if wOk > wCancel {
		w = wCancel
	}
	btnSize := fyne.NewSize(w, ok.MinSize().Height)
	okC := container.NewGridWrap(btnSize, ok)
	cancelC := container.NewGridWrap(btnSize, cancel)

	labelAutoLogOut := widget.NewLabel(lang.X("settings.autologout", "Logout after [sec]\n(0=disabled)"))
	labelCopyToClipboard := widget.NewLabel(lang.X("settings.copytoclipboard", "Copy to clipboard"))
	labelAutoExportPath := widget.NewLabel(lang.X("settings.autoexporpath", "Auto export path"))
	labelExportMode := widget.NewLabel(lang.X("settings.exportmode", "Export file mode"))

	s.exportMode = widget.NewSelect([]string{
		lang.X("settings.exportmode.new", "New file"),
		lang.X("settings.exportmode.overwrite", "Overwrite existing file"),
		lang.X("settings.exportmode.ask", "Ask"),
	}, nil)

	s.viewAfterEdit = widget.NewSelect([]string{
		lang.X("settings.viewmode.categoty", "Catergory"),
		lang.X("settings.viewmode.view", "View"),
	}, nil)

	s.viewAfterAdd = widget.NewSelect([]string{
		lang.X("settings.viewmode.categoty", "Catergory"),
		lang.X("settings.viewmode.view", "View"),
	}, nil)
	s.checkForUpdates = widget.NewCheck(lang.X("settings.autocheckupdate", "Automatically check for updates"), nil)

	form := container.New(layout.NewFormLayout(),
		labelAutoLogOut, s.autoLogOutTime,
		labelCopyToClipboard, s.copytoclipboard,
		widget.NewLabel(lang.X("settings.viewmode.after.add", "View after Add")), s.viewAfterAdd,
		widget.NewLabel(lang.X("settings.viewmode.after.edit", "View after Edit")), s.viewAfterEdit,
	)
	if !Gui.IsDesktop {
		form.Add(labelExportMode)
		form.Add(s.exportMode)
	}
	form = container.NewVBox(form, s.checkForUpdates)
	browseBtn := widget.NewButton(lang.X("settings.browse", "Browse"),
		func() {
			dia := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
				if err != nil {
					return
				}
				if writer == nil {
					return
				}
				defer writer.Close()
				s.autoExportPath.SetText(writer.URI().String())
			}, Gui.MainWindow)
			dia.SetView(dialog.ListView)
			fName := path.Base(s.autoExportPath.Text)
			dia.SetFileName(fName)
			filter := storage.NewExtensionFileFilter([]string{".db"})
			dia.SetFilter(filter)
			dia.Show()
			if Gui.IsDesktop {
				si := Gui.MainWindow.Canvas().Size()
				var windowScale float32 = 1.0
				dia.Resize(fyne.NewSize(si.Width*windowScale, si.Height*windowScale))
			}
		})
	exportJson := widget.NewButton(lang.X("settings.export.json", "Export to JSON"), func() {
		c := widget.NewLabel(lang.X("settings.export.msg", "Do you really want to export the whole database to unencrypted JSON text file ?"))
		c.Wrapping = fyne.TextWrapWord
		c.Importance = widget.DangerImportance
		c.TextStyle = fyne.TextStyle{
			Bold: true,
		}
		dia := dialog.NewCustomConfirm(lang.X("settings.export.json", "Export as unencrypted JSON"),
			lang.X("settings.export", "Export"), lang.X("cancel", "Cancel"),
			c, func(ok bool) {
				if !ok {
					return
				}
				diaf := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
					if err != nil {
						return
					}
					if writer == nil {
						return
					}
					defer writer.Close()
					SetBusy(true)
					go ExportToJson(writer, func(string, error) {
						if err != nil {
							UIErrorHandler(err)
						}
						SetBusy(false)
					})
				}, Gui.MainWindow)
				diaf.SetView(dialog.ListView)
				filter := storage.NewExtensionFileFilter([]string{".json"})
				diaf.SetFilter(filter)
				diaf.Show()
				si := Gui.MainWindow.Canvas().Size()
				var windowScale float32 = 1.0
				diaf.Resize(fyne.NewSize(si.Width*windowScale, si.Height*windowScale))
			}, Gui.MainWindow)
		dia.SetConfirmImportance(widget.DangerImportance)
		dia.Show()
	})
	exportJson.Importance = widget.DangerImportance

	exportJsonC := container.NewHBox(exportJson)

	updateCheck := widget.NewButton(lang.X("settings.checkupdate", "Check for update"), func() {
		CheckForUpdate(false)
	})

	if Gui.IsDesktop {
		p := container.NewBorder(labelAutoExportPath, nil, nil, browseBtn, s.autoExportPath)
		s.content = container.NewVBox(form, p, util.NewVFiller(1), exportJsonC, util.NewVFiller(1), updateCheck, util.NewVFiller(1), layout.NewSpacer(), container.NewHBox(layout.NewSpacer(), cancelC, okC, layout.NewSpacer()))
	} else {
		s.content = container.NewVBox(form, util.NewVFiller(1), exportJsonC, util.NewVFiller(1), updateCheck, layout.NewSpacer(), container.NewHBox(layout.NewSpacer(), cancelC, okC, layout.NewSpacer()))
	}
	return &s
}

func (s *SettingsView) SelToViewMode(index int) ViewModeType {
	var viewMode ViewModeType
	switch index {
	case 0:
		viewMode = VIEWMODE_CATEGORY
	case 1:
		viewMode = VIEWMODE_VIEW
	}
	return viewMode
}

func (s *SettingsView) ViewModeToSel(viewMode ViewModeType) int {
	var index int
	switch viewMode {
	case VIEWMODE_CATEGORY:
		index = 0
	case VIEWMODE_VIEW:
		index = 1
	}
	return index
}

func (s *SettingsView) doSave() {
	v, err := strconv.Atoi(s.autoLogOutTime.Text)
	if err != nil {
		return
	}
	Gui.Settings.AutoLogOut = v
	index := s.copytoclipboard.SelectedIndex()
	mode := COPYTOCLIPBOARD_NONE
	switch index {
	case 0:
		mode = COPYTOCLIPBOARD_NONE
	case 1:
		mode = COPYTOCLIPBOARD_TAPPED
	case 2:
		mode = COPYTOCLIPBOARD_SECONDARYTAPPED
	case 3:
		mode = COPYTOCLIPBOARD_BOTH
	}
	Gui.Settings.CopyToClipboard = mode
	Gui.Settings.AutoExportPath = s.autoExportPath.Text

	exportMode := EXPORTMODE_ASK
	index = s.exportMode.SelectedIndex()
	switch index {
	case 0:
		exportMode = EXPORTMODE_NEW
	case 1:
		exportMode = EXPORTMODE_OVERWRITE
	case 2:
		exportMode = EXPORTMODE_ASK
	}
	Gui.Settings.ExportMode = exportMode
	Gui.Settings.ViewModeAfterAdd = s.SelToViewMode(s.viewAfterAdd.SelectedIndex())
	Gui.Settings.ViewModeAfterEdit = s.SelToViewMode(s.viewAfterEdit.SelectedIndex())

	Gui.Settings.AutoUpdateCheck = s.checkForUpdates.Checked

	Gui.Settings.Store()
	RestoreBeforeSettings()
}

func (s *SettingsView) doCancel() {
	RestoreBeforeSettings()
}

func (s *SettingsView) Init() {
	s.autoLogOutTime.SetText(strconv.Itoa(Gui.Settings.AutoLogOut))
	selIndex := 0
	switch Gui.Settings.CopyToClipboard {
	case COPYTOCLIPBOARD_NONE:
		selIndex = 0
	case COPYTOCLIPBOARD_TAPPED:
		selIndex = 1
	case COPYTOCLIPBOARD_SECONDARYTAPPED:
		selIndex = 2
	case COPYTOCLIPBOARD_BOTH:
		selIndex = 3
	}
	s.copytoclipboard.SetSelectedIndex(selIndex)
	s.autoExportPath.SetText(Gui.Settings.AutoExportPath)
	selIndex = 0
	switch Gui.Settings.ExportMode {
	case EXPORTMODE_NEW:
		selIndex = 0
	case EXPORTMODE_OVERWRITE:
		selIndex = 1
	case EXPORTMODE_ASK:
		selIndex = 2
	}
	s.exportMode.SetSelectedIndex(selIndex)

	s.viewAfterAdd.SetSelectedIndex(s.ViewModeToSel(Gui.Settings.ViewModeAfterAdd))
	s.viewAfterEdit.SetSelectedIndex(s.ViewModeToSel(Gui.Settings.ViewModeAfterEdit))

	s.checkForUpdates.SetChecked(Gui.Settings.AutoUpdateCheck)
}

func (s *SettingsView) GetContent() *fyne.Container {
	Gui.MainWindow.Canvas().Focus(s.autoLogOutTime)
	return s.content
}

func (s *SettingsView) UpdateToolBar() {
	Gui.Toolbar.Items = []widget.ToolbarItem{Gui.toolToggleTheme, widget.NewToolbarSpacer(), Gui.toolInfo}
	Gui.Toolbar.Refresh()
}

func (s *SettingsView) ThemeChanged() {
	s.content.Refresh()
}
