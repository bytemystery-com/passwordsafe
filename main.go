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
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"bytemystery-com/passwordsafe/crypt"
	"bytemystery-com/passwordsafe/database"
	"bytemystery-com/passwordsafe/omap"
	"bytemystery-com/passwordsafe/passsafetheme"
	"bytemystery-com/passwordsafe/util"

	_ "net/http/pprof"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type GUI struct {
	App           fyne.App
	MainWindow    fyne.Window
	Toolbar       *widget.Toolbar
	IsDesktop     bool
	Icon          *fyne.StaticResource
	FyneSettings  fyne.Settings
	Settings      *Preferences
	Theme         *passsafetheme.PassSafeTheme
	DarkMenuItem  *fyne.MenuItem
	LightMenuItem *fyne.MenuItem

	IconPlusUp   *fyne.StaticResource
	IconPlusDown *fyne.StaticResource
	IconPlusX    *fyne.StaticResource

	IconLockUp   *fyne.StaticResource
	IconLockDown *fyne.StaticResource
	IconLockX    *fyne.StaticResource

	IconEditUp   *fyne.StaticResource
	IconEditDown *fyne.StaticResource
	IconEditX    *fyne.StaticResource

	IconViewUp   *fyne.StaticResource
	IconViewDown *fyne.StaticResource
	IconViewX    *fyne.StaticResource

	IconBackUp   *fyne.StaticResource
	IconBackDown *fyne.StaticResource
	IconBackX    *fyne.StaticResource

	IconSearchUp   *fyne.StaticResource
	IconSearchDown *fyne.StaticResource
	IconSearchX    *fyne.StaticResource

	IconEmpty *fyne.StaticResource

	IconImport *fyne.StaticResource
	IconExport *fyne.StaticResource

	IconHooverBottom *fyne.StaticResource

	IconCollection omap.OMap[string, *fyne.StaticResource]

	categoryView  *CategoryView
	entryEditView *EntryEditView
	entryView     *EntryView
	loginView     *LoginView
	settingsView  *SettingsView

	oldContent            fyne.CanvasObject
	beforeSettingsContent fyne.CanvasObject
	lockTimer             *time.Timer
	DatabaseFile          string
	logInMode             LoginMode
	content               *fyne.Container
	busy                  *fyne.Container

	toolAdd          *widget.ToolbarAction
	toolEdit         *widget.ToolbarAction
	toolRemove       *widget.ToolbarAction
	toolChangePasswd *widget.ToolbarAction
	toolSettings     *widget.ToolbarAction
	toolExport       *widget.ToolbarAction
	toolImport       *widget.ToolbarAction
	toolInfo         *widget.ToolbarAction
	toolToggleTheme  *widget.ToolbarAction
	toolDelEntry     *widget.ToolbarAction
}

//go:embed assets/*
var assets embed.FS

var (
	Gui            = GUI{}
	Database       *database.Db
	Crypt          *crypt.Crypt
	MasterPasswort []byte
)

func forceLanguage() {
	if *Flags.language == "" {
		return
	}
	// Hack. Ongoing discussion in https://github.com/fyne-io/fyne/issues/5333
	lcontent, err := assets.ReadFile("assets/lang/" + *Flags.language + ".json")
	if err != nil {
		return
	}
	lang.AddTranslationsForLocale(lcontent, lang.SystemLocale())
}

type FlagsType struct {
	language *string
}

var Flags FlagsType

type UpdateToolbarInterface interface {
	UpdateToolBar()
	ThemeChanged()
}

func main() {
	Flags.language = flag.String("l", "", "language (en, de ....)")
	flag.Parse()

	loadTranslations(assets, "assets/lang")
	forceLanguage()

	// go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
	// top
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	Gui.App = app.NewWithID("com.bytemystery.passwordsafe2")
	loadIcons()
	loadPreferences()
	Gui.FyneSettings = Gui.App.Settings()
	var tv fyne.ThemeVariant
	switch Gui.Settings.ThemeVariant {
	case -1:
		tv = fyne.CurrentApp().Settings().ThemeVariant()
	case 0:
		tv = theme.VariantDark
	case 1:
		tv = theme.VariantLight
	}

	Gui.Theme = passsafetheme.NewPassSafeTheme(tv)
	Gui.App.Settings().SetTheme(Gui.Theme)

	if _, ok := Gui.App.(desktop.App); ok {
		Gui.IsDesktop = true
	}
	Gui.MainWindow = Gui.App.NewWindow("PasswordSafe")
	Gui.MainWindow.SetIcon(Gui.Icon)

	Gui.toolAdd = widget.NewToolbarAction(theme.ContentAddIcon(), func() { Gui.categoryView.Add() })
	Gui.toolEdit = widget.NewToolbarAction(theme.DocumentCreateIcon(), func() { Gui.categoryView.Edit() })
	Gui.toolRemove = widget.NewToolbarAction(theme.DeleteIcon(), func() { Gui.categoryView.Delete() })
	Gui.toolDelEntry = widget.NewToolbarAction(theme.DeleteIcon(), func() { Gui.entryView.DelEntry() })

	Gui.toolChangePasswd = widget.NewToolbarAction(theme.AccountIcon(), func() { doChangePassword() })
	Gui.toolSettings = widget.NewToolbarAction(theme.SettingsIcon(), func() { SetSettingsView() })
	Gui.toolExport = widget.NewToolbarAction(theme.NewThemedResource(Gui.IconExport), func() { doExport() })
	Gui.toolImport = widget.NewToolbarAction(theme.NewThemedResource(Gui.IconImport), func() { doImport() })
	Gui.toolInfo = widget.NewToolbarAction(theme.InfoIcon(), func() { showInfoDialog() })

	Gui.toolToggleTheme = widget.NewToolbarAction(theme.BrokenImageIcon(), func() { toggleTheme() })

	Gui.Toolbar = widget.NewToolbar(Gui.toolAdd, Gui.toolEdit, Gui.toolRemove, widget.NewToolbarSeparator(), Gui.toolChangePasswd, widget.NewToolbarSeparator(), widget.NewToolbarSpacer(), Gui.toolInfo)

	scaling := theme.Size("text") / 14.0

	Gui.categoryView = NewCategoryView()
	Gui.entryEditView = NewEntryEditView()
	Gui.entryView = NewEntryView()
	Gui.loginView = NewLoginView()
	Gui.settingsView = NewSettingsView()

	Gui.content = container.NewStack()
	backGround := canvas.NewRectangle(Gui.Theme.GetSpecialColor("wait_background"))
	Gui.busy = container.NewStack(backGround, container.NewVBox(layout.NewSpacer(), widget.NewProgressBarInfinite(), layout.NewSpacer()))
	Gui.MainWindow.SetContent(container.NewBorder(Gui.Toolbar, nil, nil, nil, Gui.content))

	Gui.MainWindow.Resize(fyne.NewSize(400*scaling, 700*scaling))
	Gui.MainWindow.CenterOnScreen()

	var err error
	Gui.DatabaseFile, err = database.GetDBFile("passwordsafe")
	if err != nil {
		log.Println(err)
		return
	}
	Database = database.NewDb()
	defer Database.Close()

	Gui.App.Lifecycle().SetOnExitedForeground(func() {
		// if fyne.CurrentDevice().IsMobile() {
		Gui.lockTimer = time.AfterFunc(time.Duration(Gui.Settings.AutoLogOut)*time.Second, func() {
			fyne.Do(func() {
				DoLock()
			})
		})
	})
	Gui.App.Lifecycle().SetOnEnteredForeground(func() {
		if Gui.lockTimer != nil {
			Gui.lockTimer.Stop()
			Gui.lockTimer = nil
		}
	})

	_, err = os.Stat(Gui.DatabaseFile)
	if err != nil {
		Gui.logInMode = LOGIN_NEW
	}

	fyne.CurrentApp().Settings().AddListener(func(settings fyne.Settings) {
		updateTheme()
	})

	SetLoginView()

	Gui.MainWindow.Show()
	fyne.Do(func() {
		Gui.loginView.GetContent()
	})

	Gui.App.Run()
}

func SetBusy(busy bool) {
	fyne.Do(func() {
		if busy {
			Gui.busy.Show()
		} else {
			Gui.busy.Hide()
		}
	})
}

func setContent(c fyne.CanvasObject) {
	Gui.content.RemoveAll()
	Gui.content.Add(c)
	Gui.content.Add(Gui.busy)
	Gui.busy.Hide()
	if c == Gui.categoryView.GetContent() {
		Gui.categoryView.UpdateToolBar()
	} else if c == Gui.entryEditView.GetContent() {
		Gui.entryEditView.UpdateToolBar()
	} else if c == Gui.entryView.GetContent() {
		Gui.entryView.UpdateToolBar()
	} else if c == Gui.loginView.GetContent() {
		Gui.loginView.UpdateToolBar()
	}
}

func AddEntry() {
	id, _ := Gui.categoryView.GetCategoryId()
	Gui.entryEditView.NewEntry(id)
	setContent(Gui.entryEditView.GetContent())
	Gui.MainWindow.Canvas().Focus(Gui.entryEditView.name)
}

func SetCatgeoryView(categoryId int64) {
	setContent(Gui.categoryView.GetContent())
	Gui.categoryView.Update(categoryId)
}

func SetEntryView(id int64) {
	setContent(Gui.entryView.GetContent())
	Gui.entryView.Update(id)
}

func SetEditView(id int64) {
	Gui.entryEditView.SetEntry(id)
	setContent(Gui.entryEditView.GetContent())
}

func SetLoginView() {
	Gui.loginView.Reset(Gui.logInMode)
	setContent(Gui.loginView.GetContent())
}

func SetSettingsView() {
	Gui.beforeSettingsContent = Gui.content.Objects[0]
	Gui.settingsView.Init()
	setContent(Gui.settingsView.GetContent())
}

func DoLogin(pass []byte) {
	if Gui.logInMode == LOGIN_NEW {
		Crypt = crypt.NewCrypt(nil)
	} else {
		err := Database.Open(Gui.DatabaseFile, nil, nil)
		if err != nil {
			log.Println(err)
			return
		}
		cfg, err := Database.GetCryptCfg()
		Database.Close()
		if err != nil {
			UIErrorHandler(err)
			return
		}
		Crypt = crypt.NewCrypt(cfg)
	}
	err := Database.Open(Gui.DatabaseFile, pass, Crypt)
	if err != nil {
		UIErrorHandler(err)
		return
	}
	k, err := Database.GetMasterKey(pass, Crypt)
	if err != nil {
		UIErrorHandlerWithMessage(err, lang.X("pass.err", "Wrong password !"))
		crypt.ErasePassword(MasterPasswort)
		Database.Close()
		return
	}
	crypt.ErasePassword(k)
	MasterPasswort = pass
	// err = Database.ConverToJson(MasterPasswort, Crypt)
	switch Gui.logInMode {
	case LOGIN_CHANGE_PASSWORD:
		SetLoginView()
		Gui.logInMode = LOGIN_LOGIN
	case LOGIN_NEW:
		Gui.logInMode = LOGIN_LOGIN
		SetCatgeoryView(-1)
	case LOGIN_LOGIN:
		if Gui.oldContent != nil {
			setContent(Gui.oldContent)
			Gui.oldContent = nil
		} else {
			SetCatgeoryView(int64(Gui.Settings.LastCategoryId))
		}
	}
	if Gui.Settings.AutoUpdateCheck {
		CheckForUpdate(true)
	}
}

func DoLock() {
	Gui.oldContent = Gui.content.Objects[0]
	Database.Close()
	Crypt = nil
	crypt.ErasePassword(MasterPasswort)
	SetLoginView()

	if Gui.Settings.AutoExportPath != "" {
		f := func() error {
			data, err := os.ReadFile(Gui.DatabaseFile)
			if err != nil {
				return err
			}
			return util.WriteFile(Gui.Settings.AutoExportPath, data)
		}
		err := f()
		if err != nil {
			UIErrorHandler(err)
		}
	}
}

func GetDefaultIcon() *fyne.StaticResource {
	icon, ok := Gui.IconCollection.GetByKey("0")
	if !ok {
		return nil
	}
	return icon
}

func doChangePassword() {
	DoLock()
	Gui.logInMode = LOGIN_CHANGE_PASSWORD
}

func RestoreBeforeSettings() {
	if Gui.beforeSettingsContent != nil {
		setContent(Gui.beforeSettingsContent)
		Gui.beforeSettingsContent = nil
		return
	}
	SetLoginView()
}

func DoChangePasswd(pass []byte) {
	if Gui.logInMode == LOGIN_NEW {
		DoLogin(pass)
	} else {
		err := Database.ChangePassword(MasterPasswort, pass, Crypt)
		if err != nil {
			UIErrorHandler(err)
			crypt.ErasePassword(pass)
			return
		}
		crypt.ErasePassword(MasterPasswort)
		MasterPasswort = pass
		SetLoginView()
	}
}

func doExport() {
	ExportDatabase(func(err error) {
		if err != nil {
			UIErrorHandler(err)
		} else {
			SendNotification(lang.X("export.ok.title", "Export"), lang.X("export.ok.msg", "Database was successfully exported !"))
		}
	})
	SetLoginView()
}

func ExportDatabase(fDone func(error)) {
	Database.Close()

	w := func(writer fyne.URIWriteCloser) {
		f, err := os.Open(Gui.DatabaseFile)
		if err != nil {
			if fDone != nil {
				fDone(err)
			}
			return
		}
		defer f.Close()
		_, err = io.Copy(writer, f)
		if fDone != nil {
			fDone(err)
		}
	}

	ShowExportTypeDialog(func(overwrite bool) {
		var dia *dialog.FileDialog
		if overwrite {
			dia = dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
				if err != nil {
					if fDone != nil {
						fDone(err)
					}
					return
				}
				if reader == nil {
					if fDone != nil {
						fDone(errors.New("Reader == nil"))
					}
					return
				}
				defer reader.Close()
				writer, err := storage.Writer(reader.URI())
				if err != nil {
					if fDone != nil {
						fDone(err)
					}
					return
				}
				if writer == nil {
					if fDone != nil {
						fDone(errors.New("Writer2 == nil"))
					}
					return
				}
				defer writer.Close()
				w(writer)
			}, Gui.MainWindow)
		} else {
			dia = dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
				if err != nil {
					if fDone != nil {
						fDone(err)
					}
					return
				}
				if writer == nil {
					/*
						if fDone != nil {
							fDone(errors.New("Writer1 == nil"))
						}
					*/
					return
				}
				defer writer.Close()
				w(writer)
			}, Gui.MainWindow)
		}
		dia.SetView(dialog.ListView)
		if !overwrite {
			fName := path.Base(Gui.DatabaseFile)
			dia.SetFileName(fName)
		}
		if Gui.IsDesktop {
			filter := storage.NewExtensionFileFilter([]string{".db"})
			dia.SetFilter(filter)
		}
		dia.Show()
		if Gui.IsDesktop {
			si := Gui.MainWindow.Canvas().Size()
			var windowScale float32 = 1.0
			dia.Resize(fyne.NewSize(si.Width*windowScale, si.Height*windowScale))
		}
	})
}

func doImport() {
	ImportDatabase(func(err error) {
		if err != nil {
			UIErrorHandler(err)
		} else {
			SendNotification(lang.X("import.ok.title", "Import"), lang.X("import.ok.msg", "Database was successfully imported !"))
		}
	})
	SetLoginView()
}

func ImportDatabase(fDone func(error)) {
	thisTs, err := Database.GetLastWrite()
	if err != nil {
		if fDone != nil {
			fDone(err)
		}
		return
	}
	Database.Close()
	dia := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			if fDone != nil {
				fDone(err)
			}
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		newFile, err := database.GetDBFile("new")
		if err != nil {
			if fDone != nil {
				fDone(err)
			}
			return
		}
		newDb, err := os.OpenFile(newFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o664)
		if err != nil {
			if fDone != nil {
				fDone(err)
			}
			return
		}
		_, err = io.Copy(newDb, reader)
		if err != nil {
			newDb.Close()
			if fDone != nil {
				fDone(err)
			}
			return
		}
		newDb.Close()

		err = Database.Open(newFile, MasterPasswort, Crypt)
		if err != nil {
			if fDone != nil {
				fDone(err)
			}
			return
		}
		newTs, err := Database.GetLastWrite()
		Database.Close()
		if err != nil {
			if fDone != nil {
				fDone(err)
			}
			return
		}

		replaceDb := func() error {
			backupFile, err := database.GetDBFile("backup")
			if err != nil {
				if fDone != nil {
					fDone(err)
				}
				return err
			}
			err = util.RenameFile(Gui.DatabaseFile, backupFile)
			if err != nil {
				if fDone != nil {
					fDone(err)
				}
				return err
			}
			err = util.RenameFile(newFile, Gui.DatabaseFile)
			if err != nil {
				if fDone != nil {
					fDone(err)
				}
				return err
			}
			return nil
		}

		if newTs.Before(thisTs) {
			msg := widget.NewLabel(fmt.Sprintf(lang.X("import.isolder.msg", "Imported database is older\n(%s)\nthan actual\n(%s) !\n\nProceed importing nevertheless ?"),
				util.FormatDateTime(newTs, true), util.FormatDateTime(thisTs, true)))
			msg.Importance = widget.WarningImportance
			msg.Wrapping = fyne.TextWrapWord
			msg.TextStyle = fyne.TextStyle{
				Bold: true,
			}
			msg.Alignment = fyne.TextAlignCenter
			dia := dialog.NewCustomConfirm(lang.X("import.isolder.title", "Datenbank ist älter"),
				lang.X("import.import", "Import"), lang.X("cancel", "Cancel"), msg,
				func(ok bool) {
					if !ok {
						return
					}
					err = replaceDb()
					if fDone != nil {
						fDone(err)
					}
				}, Gui.MainWindow)
			dia.Show()
			dia.Resize(fyne.NewSize(Gui.MainWindow.Canvas().Size().Width, dia.MinSize().Height))

		} else {
			err = replaceDb()
			if fDone != nil {
				fDone(err)
			}
		}
	}, Gui.MainWindow)
	dia.SetView(dialog.ListView)
	if Gui.IsDesktop {
		filter := storage.NewExtensionFileFilter([]string{".db"})
		dia.SetFilter(filter)
	}
	dia.Show()
	if Gui.IsDesktop {
		si := Gui.MainWindow.Canvas().Size()
		var windowScale float32 = 1.0
		dia.Resize(fyne.NewSize(si.Width*windowScale, si.Height*windowScale))
	}
}

func toggleTheme() {
	if Gui.Theme.GetVariant() == theme.VariantDark {
		Gui.Theme.SetVariant(theme.VariantLight)
	} else {
		Gui.Theme.SetVariant(theme.VariantDark)
	}
	Gui.Settings.ThemeVariant = int(Gui.Theme.GetVariant())
	Gui.Settings.Store()
	Gui.App.Settings().SetTheme(Gui.Theme)
	updateTheme()
}

func updateTheme() {
	Gui.toolAdd.ToolbarObject().Refresh()
	Gui.toolEdit.ToolbarObject().Refresh()
	Gui.toolRemove.ToolbarObject().Refresh()
	Gui.toolChangePasswd.ToolbarObject().Refresh()
	Gui.toolSettings.ToolbarObject().Refresh()
	Gui.toolExport.ToolbarObject().Refresh()
	Gui.toolImport.ToolbarObject().Refresh()
	Gui.toolInfo.ToolbarObject().Refresh()
	Gui.toolToggleTheme.ToolbarObject().Refresh()
	Gui.toolDelEntry.ToolbarObject().Refresh()

	Gui.loginView.ThemeChanged()
	Gui.categoryView.ThemeChanged()
	Gui.entryView.ThemeChanged()
	Gui.entryEditView.ThemeChanged()
	Gui.settingsView.ThemeChanged()
}

func ExportToJson(writer fyne.URIWriteCloser, done func(string, error)) (string, error) {
	str, err := Database.ExportToJson(MasterPasswort, Crypt)
	if err != nil {
		if done != nil {
			done("", err)
		}
		return "", err
	}
	if writer != nil {
		defer writer.Close()
		err = util.WriteFileToStorage(writer, []byte(str))
		if err != nil {
			if done != nil {
				done("", err)
			}
			return "", err
		}
		path := writer.URI().String()
		if done != nil {
			done(path, nil)
		}
		return path, nil
	} else {
		dir := filepath.Dir(Gui.DatabaseFile)
		path := filepath.Join(dir, "passwordsafe.json")
		err = os.WriteFile(path, []byte(str), 0o644)
		if err != nil {
			if done != nil {
				done("", err)
			}
			return "", err
		}
		if done != nil {
			done(path, nil)
		}
		return path, nil
	}
}

func CheckForUpdate(notify bool) {
	if notify {
		now := time.Now().Unix()
		if now-Gui.Settings.LastUpdatecheck < int64(Gui.Settings.UpdateCheckInterval)*3600 {
			return
		}
	}
	go func() {
		m := Gui.App.Metadata()
		type Version struct {
			maj   int
			min   int
			patch int
		}
		thisVersion := Version{}
		gitVersion := Version{}
		web, newVer, err := util.CheckForUpdate()
		if err != nil {
			return
		}
		n, err := fmt.Sscanf(m.Version, "%d.%d.%d", &thisVersion.maj, &thisVersion.min, &thisVersion.patch)
		if n != 3 || err != nil {
			return
		}
		n, err = fmt.Sscanf(newVer, "v%d.%d.%d", &gitVersion.maj, &gitVersion.min, &gitVersion.patch)
		if n != 3 || err != nil {
			return
		}
		if thisVersion.maj < gitVersion.maj || (thisVersion.maj == gitVersion.maj && thisVersion.min < gitVersion.min) ||
			(thisVersion.maj == gitVersion.maj && thisVersion.min == gitVersion.min && thisVersion.patch < gitVersion.patch) {
			link, err := url.Parse(web)
			if err != nil {
				return
			}
			fyne.Do(func() {
				if notify {
					SendNotification(lang.X("update.notify.title", "New version"), fmt.Sprintf(lang.X("update.notify.msg", "New version %s is available"), newVer))
					Gui.Settings.LastUpdatecheck = time.Now().Unix()
					Gui.Settings.Store()
				} else {
					msg := widget.NewHyperlinkWithStyle(fmt.Sprintf(lang.X("update.msg", "A new version %s is available !"), newVer),
						link, fyne.TextAlignCenter, fyne.TextStyle{
							Bold: true,
						})
					var dia *dialog.CustomDialog
					ok := widget.NewButton(lang.X("ok", "Ok"), func() {
						dia.Hide()
					})
					dia = dialog.NewCustomWithoutButtons(lang.X("update.title", "Update"),
						container.NewVBox(msg, util.NewVFiller(2), ok), Gui.MainWindow)
					dia.Show()
					dia.Resize(fyne.NewSize(Gui.MainWindow.Canvas().Size().Width, dia.MinSize().Height))
				}
			})
		} else {
			if !notify {
				fyne.Do(func() {
					dialog.ShowInformation(lang.X("update.title", "Update"), lang.X("update.nonew", "You are already running the latest version."), Gui.MainWindow)
				})
			}
		}
	}()
}
