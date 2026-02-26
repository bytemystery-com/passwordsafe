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

type CopyToClipboardType int

const (
	COPYTOCLIPBOARD_NONE CopyToClipboardType = iota
	COPYTOCLIPBOARD_TAPPED
	COPYTOCLIPBOARD_SECONDARYTAPPED
	COPYTOCLIPBOARD_BOTH
)

type ExportModeType int

const (
	EXPORTMODE_NEW ExportModeType = iota
	EXPORTMODE_OVERWRITE
	EXPORTMODE_ASK
)

type ViewModeType int

const (
	VIEWMODE_CATEGORY ViewModeType = iota
	VIEWMODE_VIEW
)

const (
	PREF_THEMEVARIANT_KEY            = "theme"
	PREF_THEMEVARIANT_VALUE          = -1
	PREF_AUTOLOGOUT_KEY              = "autologout"
	PREF_AUTOLOGOUT_VALUE            = 30
	PREF_COPYTOCLIPBOARD_KEY         = "copytoclipboard"
	PREF_COPYTOCLIPBOARD_VALUE       = COPYTOCLIPBOARD_SECONDARYTAPPED
	PREF_LASTCATEGORYID_KEY          = "lastcategoryid"
	PREF_LASTCATEGORYID_VALUE        = -1
	PREF_AUTOEXPORTPATH_KEY          = "autoexportpath"
	PREF_AUTOEXPORTPATH_VALUE        = ""
	PREF_EXPORTMODE_KEY              = "exportmode"
	PREF_EXPORMODE_VALUE             = EXPORTMODE_ASK
	PREF_MINPASSLENGTH_KEY           = "minpasslength"
	PREF_MINPASSLENGTH_MOBILE_VALUE  = 4
	PREF_MINPASSLENGTH_DESKTOP_VALUE = 8
	PREF_VIEWMODE_AFTER_EDIT_KEY     = "viewafteredit"
	PREF_VIEWMODE_AFTER_EDIT_VALUE   = VIEWMODE_VIEW
	PREF_VIEWMODE_AFTER_ADD_KEY      = "viewafteradd"
	PREF_VIEWMODE_AFTER_ADD_VALUE    = VIEWMODE_CATEGORY
)

type Preferences struct {
	ThemeVariant      int
	AutoLogOut        int
	CopyToClipboard   CopyToClipboardType
	LastCategoryId    int
	AutoExportPath    string
	ExportMode        ExportModeType
	MinPassLength     int
	ViewModeAfterEdit ViewModeType
	ViewModeAfterAdd  ViewModeType
}

func NewPreferences() *Preferences {
	p := &Preferences{
		ThemeVariant:      Gui.App.Preferences().IntWithFallback(PREF_THEMEVARIANT_KEY, PREF_THEMEVARIANT_VALUE),
		AutoLogOut:        Gui.App.Preferences().IntWithFallback(PREF_AUTOLOGOUT_KEY, PREF_AUTOLOGOUT_VALUE),
		CopyToClipboard:   CopyToClipboardType(Gui.App.Preferences().IntWithFallback(PREF_COPYTOCLIPBOARD_KEY, int(PREF_COPYTOCLIPBOARD_VALUE))),
		LastCategoryId:    Gui.App.Preferences().IntWithFallback(PREF_LASTCATEGORYID_KEY, PREF_LASTCATEGORYID_VALUE),
		AutoExportPath:    Gui.App.Preferences().StringWithFallback(PREF_AUTOEXPORTPATH_KEY, PREF_AUTOEXPORTPATH_VALUE),
		ExportMode:        ExportModeType(Gui.App.Preferences().IntWithFallback(PREF_EXPORTMODE_KEY, int(PREF_EXPORMODE_VALUE))),
		ViewModeAfterEdit: ViewModeType(Gui.App.Preferences().IntWithFallback(PREF_VIEWMODE_AFTER_EDIT_KEY, int(PREF_VIEWMODE_AFTER_EDIT_VALUE))),
		ViewModeAfterAdd:  ViewModeType(Gui.App.Preferences().IntWithFallback(PREF_VIEWMODE_AFTER_ADD_KEY, int(PREF_VIEWMODE_AFTER_ADD_VALUE))),
	}
	if Gui.IsDesktop {
		p.MinPassLength = Gui.App.Preferences().IntWithFallback(PREF_MINPASSLENGTH_KEY, PREF_MINPASSLENGTH_DESKTOP_VALUE)
	} else {
		p.MinPassLength = Gui.App.Preferences().IntWithFallback(PREF_MINPASSLENGTH_KEY, PREF_MINPASSLENGTH_MOBILE_VALUE)
	}
	return p
}

func (p *Preferences) Store() {
	pref := Gui.App.Preferences()
	pref.SetInt(PREF_THEMEVARIANT_KEY, p.ThemeVariant)
	pref.SetInt(PREF_AUTOLOGOUT_KEY, p.AutoLogOut)
	pref.SetInt(PREF_COPYTOCLIPBOARD_KEY, int(p.CopyToClipboard))
	pref.SetInt(PREF_LASTCATEGORYID_KEY, p.LastCategoryId)
	pref.SetString(PREF_AUTOEXPORTPATH_KEY, p.AutoExportPath)
	pref.SetInt(PREF_EXPORTMODE_KEY, int(p.ExportMode))
	pref.SetInt(PREF_MINPASSLENGTH_KEY, p.MinPassLength)
	pref.SetInt(PREF_VIEWMODE_AFTER_EDIT_KEY, int(p.ViewModeAfterEdit))
	pref.SetInt(PREF_VIEWMODE_AFTER_ADD_KEY, int(p.ViewModeAfterAdd))
}
