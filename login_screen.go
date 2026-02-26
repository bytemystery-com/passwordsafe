package main

import (
	"errors"
	"fmt"

	"bytemystery-com/passwordsafe/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type LoginMode int

const (
	LOGIN_LOGIN LoginMode = iota
	LOGIN_CHANGE_PASSWORD
	LOGIN_NEW
)

type LoginView struct {
	content   *fyne.Container
	pass1     *widget.Entry
	pass2     *widget.Entry
	label1    *widget.Label
	label2    *widget.Label
	ok        *widget.Button
	cancel    *widget.Button
	loginMode LoginMode
}

var _ UpdateToolbarInterface = (*LoginView)(nil)

func NewLoginView() *LoginView {
	l := LoginView{}
	icon := canvas.NewImageFromResource(Gui.Icon)
	icon.FillMode = canvas.ImageFillContain
	si := Gui.Theme.GetSpecialSize("login_icon_size")

	icon.SetMinSize(fyne.NewSize(si, si))
	icon.Refresh()
	l.label1 = widget.NewLabel(lang.X("login.pass1", "Password"))
	l.pass1 = widget.NewPasswordEntry()
	l.pass1.OnSubmitted = func(string) {
		l.doLogIn()
	}
	l.label2 = widget.NewLabel(lang.X("login.pass2", "Confirm password"))
	l.pass2 = widget.NewPasswordEntry()
	l.pass2.OnSubmitted = func(string) {
		l.doLogIn()
	}
	l.ok = widget.NewButton(lang.X("login", "Log in"), func() {
		l.doLogIn()
	})
	l.ok.Importance = widget.HighImportance
	l.cancel = widget.NewButton(lang.X("cancel", "Cancel"), func() {
		l.loginMode = LOGIN_LOGIN
		Gui.logInMode = l.loginMode
		SetLoginView()
	})

	s1 := Gui.Theme.GetSpecialSize("login_space_logo_label")
	s2 := Gui.Theme.GetSpecialSize("login_space_field_ok")
	if Gui.IsDesktop {
		l.content = container.NewVBox(layout.NewSpacer(), container.NewCenter(icon), util.NewVFiller(s1),
			l.label1, l.pass1, l.label2, l.pass2, util.NewVFiller(s2), l.ok, l.cancel, layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer())
	} else {
		l.content = container.NewVBox(container.NewCenter(icon), util.NewVFiller(s1),
			l.label1, l.pass1, l.label2, l.pass2, util.NewVFiller(s2), l.ok, l.cancel)
	}
	return &l
}

func (l *LoginView) doLogIn() {
	switch l.loginMode {
	case LOGIN_CHANGE_PASSWORD, LOGIN_NEW:
		if len(l.pass1.Text) < Gui.Settings.MinPassLength {
			UIErrorHandler(fmt.Errorf(lang.X("changepasswd.tooshort", "The passwort is too short !\nAt least it should be %d characters long !"), Gui.Settings.MinPassLength))
		} else if l.pass1.Text != l.pass2.Text {
			UIErrorHandler(errors.New(lang.X("changepasswd.notequal", "The both passwords are not equal")))
		} else {
			DoChangePasswd([]byte(l.pass1.Text))
		}
	case LOGIN_LOGIN:
		DoLogin([]byte(l.pass1.Text))
	}
}

func (l *LoginView) Reset(loginMode LoginMode) {
	l.pass1.SetText("")
	l.pass2.SetText("")
	switch loginMode {
	case LOGIN_CHANGE_PASSWORD:
		l.label1.SetText(lang.X("changepasswd.pass1", "New password"))
		l.ok.SetText(lang.X("changepasswd", "Change password"))
		l.pass2.Show()
		l.label2.Show()
		l.cancel.Show()
	case LOGIN_LOGIN:
		l.label1.SetText(lang.X("login.pass1", "Password"))
		l.ok.SetText(lang.X("login", "Anmelden"))
		l.pass2.Hide()
		l.label2.Hide()
		l.cancel.Hide()
	case LOGIN_NEW:
		l.label1.SetText(lang.X("login.pass1", "Password"))
		l.ok.SetText(lang.X("create", "Create database"))
		l.pass2.Show()
		l.label2.Show()
		l.cancel.Hide()
	}
	l.loginMode = loginMode
}

func (l *LoginView) GetContent() *fyne.Container {
	l.pass1.SetText("")
	l.pass2.SetText("")
	Gui.MainWindow.Canvas().Focus(l.pass1)
	return l.content
}

func (l *LoginView) UpdateToolBar() {
	Gui.Toolbar.Items = []widget.ToolbarItem{Gui.toolToggleTheme, widget.NewToolbarSpacer(), Gui.toolInfo}
	Gui.Toolbar.Refresh()
}

func (l *LoginView) ThemeChanged() {
	l.content.Refresh()
}
