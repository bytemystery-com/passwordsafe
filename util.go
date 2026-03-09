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
	"embed"
	"errors"
	"fmt"
	"net/url"
	"path"
	"runtime"
	"runtime/debug"
	"strings"

	"bytemystery-com/passwordsafe/omap"
	"bytemystery-com/passwordsafe/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func showInfoDialog() {
	vgo := runtime.Version()[2:]
	vfyne := ""
	os := runtime.GOOS
	arch := runtime.GOARCH
	info, _ := debug.ReadBuildInfo()
	for _, dep := range info.Deps {
		if dep.Path == "fyne.io/fyne/v2" {
			vfyne = dep.Version[1:]
		}
	}
	s := fyne.CurrentApp().Settings()
	t := Gui.Theme.GetVariant()
	thema := ""
	b := s.BuildType()
	_ = b
	switch t {
	case theme.VariantDark:
		thema = lang.X("info.thema_dark", "Dark")
	case theme.VariantLight:
		thema = lang.X("info.thema_light", "Light")
	default:
		thema = lang.X("info.thema_unknown", "Unknown")
	}

	build := ""
	switch b {
	case fyne.BuildStandard:
		build = lang.X("info.build_standard", "Standard")
	case fyne.BuildDebug:
		build = lang.X("info.build_debug", "Debug")
	case fyne.BuildRelease:
		build = lang.X("info.build_release", "Release")
	default:
		build = lang.X("info.build_unknown", "Unknown")
	}

	m := Gui.App.Metadata()
	v := fmt.Sprintf("%s (%d)", m.Version, m.Build)
	n := m.Name
	if n == "" {
		n = "PasswordSafe"
	}
	tsStr := ""
	ts := m.Custom["buildts"]
	if ts != "" {
		tsStr = "Build: " + ts + "\n"
	}
	wSize := Gui.MainWindow.Canvas().Size()

	categories := 0
	entries := 0
	count, err := Database.GetNumberOfCategories()
	if err == nil {
		categories = count
	}
	count, err = Database.GetNumberOfEntries()
	if err == nil {
		entries = count
	}
	lastWriteStr := "---"
	lastWrite, err := Database.GetLastWrite()
	if err == nil {
		lastWriteStr = util.FormatDateTime(lastWrite, true)
	}

	msg := fmt.Sprintf(lang.X("info.msg", "%s\n\nVersion: %s  \n%sAuthor: Reiner Pröls\n\nGo version: %s\n\nFyne version: %s\nBuild: %s\nThema: %s\nWindow size: %.0fx%.0f\n\nPlatform: %s\nArchitecture: %s\n\n%d Entries in %d categories\n\nDatabase: %s\nModified: %s"),
		n, v, tsStr, vgo, vfyne, build, thema, wSize.Width, wSize.Height, os, arch, entries, categories, Gui.DatabaseFile, lastWriteStr)
	dialog.ShowInformation(lang.X("info.title", "Info"), msg, Gui.MainWindow)
}

func loadPreferences() {
	Gui.Settings = NewPreferences()
}

func loadIcon(path, name string) *fyne.StaticResource {
	data, err := assets.ReadFile(path)
	if err != nil {
		return nil
	}
	return fyne.NewStaticResource(name, data)
}

func loadTranslations(fs embed.FS, dir string) {
	lang.AddTranslationsFS(fs, dir)
}

func loadIcons() {
	Gui.Icon = loadIcon("assets/icons/icon.png", "icon")
	Gui.App.SetIcon(Gui.Icon)

	Gui.IconPlusUp = loadIcon("assets/icons/add_u.png", "add_u")
	Gui.IconPlusDown = loadIcon("assets/icons/add_d.png", "add_d")
	Gui.IconPlusX = loadIcon("assets/icons/add_x.png", "add_x")

	Gui.IconLockUp = loadIcon("assets/icons/lock_u.png", "lock_u")
	Gui.IconLockDown = loadIcon("assets/icons/lock_d.png", "lock_d")
	Gui.IconLockX = loadIcon("assets/icons/lock_x.png", "lock_x")

	Gui.IconEditUp = loadIcon("assets/icons/edit_u.png", "edit_u")
	Gui.IconEditDown = loadIcon("assets/icons/edit_d.png", "edit_d")
	Gui.IconEditX = loadIcon("assets/icons/edit_x.png", "edit_x")

	Gui.IconViewUp = loadIcon("assets/icons/eye_u.png", "eye_u")
	Gui.IconViewDown = loadIcon("assets/icons/eye_d.png", "eye_d")
	Gui.IconViewX = loadIcon("assets/icons/eye_x.png", "eye_x")

	Gui.IconBackUp = loadIcon("assets/icons/back_u.png", "back_u")
	Gui.IconBackDown = loadIcon("assets/icons/back_d.png", "back_d")
	Gui.IconBackX = loadIcon("assets/icons/back_x.png", "back_x")

	Gui.IconSearchUp = loadIcon("assets/icons/search_u.png", "search_u")
	Gui.IconSearchDown = loadIcon("assets/icons/search_d.png", "search_d")
	Gui.IconSearchX = loadIcon("assets/icons/search_x.png", "search_x")

	Gui.IconEmpty = loadIcon("assets/icons/empty.png", "empty")

	Gui.IconHooverBottom = loadIcon("assets/icons/hoover_b.png", "hoover_b")

	Gui.IconImport = loadIcon("assets/icons/import.svg", "import")
	Gui.IconExport = loadIcon("assets/icons/export.svg", "export")

	loadIconsForTheme()

	loadIconCollection()
}

func loadIconsForTheme() {
	dir := ""
	switch fyne.CurrentApp().Settings().ThemeVariant() {
	case theme.VariantDark:
		dir = "dark"
	case theme.VariantLight:
		dir = "light"
	default:
		dir = "light"
	}
	/*	Gui.IconImport = loadIcon("assets/icons/"+dir+"/import.png", "import")
		Gui.IconExport = loadIcon("assets/icons/"+dir+"/export.png", "export")
	*/
	_ = dir
}

func loadIconCollection() {
	Gui.IconCollection = omap.NewOMap[string, *fyne.StaticResource](300)
	rootDir := "assets/icons/collection"
	data, err := assets.ReadDir(rootDir)
	if err != nil {
		return
	}
	for _, item := range data {
		if !item.IsDir() {
			name := path.Base(item.Name())
			parts := strings.Split(name, ".")
			if len(parts) > 1 {
				name = parts[0]
			}
			path := path.Join(rootDir, item.Name())
			r := loadIcon(path, name)
			if r != nil {
				Gui.IconCollection.Add(name, r)
			}
		}
	}
}

func CloseApp() {
	if Gui.IsDesktop {
		Gui.MainWindow.Close()
		Gui.App.Quit()
	} else {
		// LogOut()
	}
}

func SendNotification(title, msg string) {
	fyne.Do(func() {
		n := fyne.NewNotification(title, msg)
		Gui.App.SendNotification(n)
	})
}

func doHelp() {
	u := url.URL{
		Scheme: "https",
		Host:   "github.com",
		Path:   "/bytemystery-com/passwordsafe",
	}
	Gui.App.OpenURL(&u)
}

func ShowExportTypeDialog(f func(overwrite bool)) {
	if Gui.IsDesktop {
		f(false)
	} else {
		switch Gui.Settings.ExportMode {
		case EXPORTMODE_NEW:
			f(false)
		case EXPORTMODE_OVERWRITE:
			f(true)
		default:
			var dia *dialog.CustomDialog
			cancel := widget.NewButton(lang.X("cancel", "Cancel"), func() {
				dia.Hide()
			})
			overwrite := widget.NewButton(lang.X("export.overwrite", "Overwrite"), func() {
				dia.Hide()
				f(true)
			})
			overwrite.Importance = widget.HighImportance
			newFile := widget.NewButton(lang.X("export.new", "New"), func() {
				dia.Hide()
				f(false)
			})
			newFile.Importance = widget.HighImportance
			c := container.NewHBox(layout.NewSpacer(), cancel, newFile, overwrite, layout.NewSpacer())
			dia = dialog.NewCustomWithoutButtons(lang.X("export.mode.title", "New file or overwrite"), c, Gui.MainWindow)
			dia.Show()
			si := Gui.MainWindow.Canvas().Size()
			var windowScale float32 = 1.0
			dia.Resize(fyne.NewSize(si.Width*windowScale, dia.MinSize().Height))
		}
	}
}

func UIErrorHandler(err error) {
	UIErrorHandlerWithMessage(err, "")
}

func UIErrorHandlerWithMessage(err error, msg string) {
	fyne.Do(func() {
		if msg != "" {
			if msg[len(msg)-1] != '\n' {
				msg += "\n"
			}
			err = errors.Join(errors.New(msg), err)
		}
		dialog.ShowError(err, Gui.MainWindow)
	})
}
