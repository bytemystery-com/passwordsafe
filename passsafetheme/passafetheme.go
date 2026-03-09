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

package passsafetheme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

const Scaling = 1.2

type PassSafeTheme struct {
	base    fyne.Theme
	variant fyne.ThemeVariant
}

func (p *PassSafeTheme) GetVariant() fyne.ThemeVariant {
	return p.variant
}

func (p *PassSafeTheme) SetVariant(variant fyne.ThemeVariant) {
	p.variant = variant
}

func NewPassSafeTheme(variant fyne.ThemeVariant) *PassSafeTheme {
	return &PassSafeTheme{
		base:    theme.DefaultTheme(),
		variant: variant,
	}
}

func (p *PassSafeTheme) Color(c fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if p.variant == theme.VariantDark {
		switch c {
		case theme.ColorNameBackground:
			return color.NRGBA{25, 25, 25, 255}
		case theme.ColorNameOverlayBackground:
			return color.NRGBA{R: 50, G: 50, B: 50, A: 255}
		case theme.ColorNameInputBackground:
			return color.NRGBA{90, 90, 90, 255}
		case theme.ColorNameFocus:
			return color.NRGBA{R: 50, G: 145, B: 245, A: 255}
		case theme.ColorNameButton:
			return color.NRGBA{R: 60, G: 60, B: 60, A: 255}
		case theme.ColorNamePrimary:
			return color.NRGBA{R: 87, G: 139, B: 255, A: 255}
			// return color.NRGBA{R: 41, G: 111, B: 246, A: 255}
		case theme.ColorNameSelection:
			return color.NRGBA{R: 132, G: 173, B: 255, A: 255}
		case theme.ColorNameError:
			// return color.NRGBA{{244, 67, 54, 255}
			return color.NRGBA{R: 255, G: 104, B: 82, A: 255}
		case theme.ColorNameForeground:
			return color.NRGBA{R: 230, G: 230, B: 230, A: 255}
		}
	} else {
		switch c {
		case theme.ColorNameBackground:
			return color.NRGBA{247, 247, 247, 255}
		case theme.ColorNameOverlayBackground:
			return color.NRGBA{R: 230, G: 230, B: 230, A: 255}
		case theme.ColorNameInputBackground:
			return color.NRGBA{210, 210, 210, 255}
		case theme.ColorNameFocus:
			return color.NRGBA{R: 160, G: 200, B: 242, A: 255}
		case theme.ColorNameButton:
			return color.NRGBA{R: 225, G: 225, B: 225, A: 255}
		case theme.ColorNameSelection:
			return color.NRGBA{R: 191, G: 225, B: 255, A: 255}

		}
	}
	/*
		val := p.base.Color(c, p.variant)
		if c == theme.ColorNameForeground {
			fmt.Println(val)
		}
		return val
	*/
	return p.base.Color(c, p.variant)
}

func (p *PassSafeTheme) Font(s fyne.TextStyle) fyne.Resource {
	return p.base.Font(s)
}

func (p *PassSafeTheme) Icon(i fyne.ThemeIconName) fyne.Resource {
	return p.base.Icon(i)
}

func (p *PassSafeTheme) Size(s fyne.ThemeSizeName) float32 {
	val := p.base.Size(s)
	switch s {
	case theme.SizeNameSubHeadingText, theme.SizeNameHeadingText, theme.SizeNameCaptionText, theme.SizeNameText:
		return val * Scaling
	}
	return val
}

func (p *PassSafeTheme) GetSpecialColor(c string) color.Color {
	if p.variant == theme.VariantDark {
		switch c {
		case "entry_title":
			return color.NRGBA{255, 249, 191, 241}
		case "entry_label":
			return color.NRGBA{21, 229, 249, 255}
		case "wait_background":
			return color.NRGBA{127, 127, 127, 180}
		}
	} else {
		switch c {
		case "entry_title":
			return color.NRGBA{200, 0, 0, 241}
		case "entry_label":
			return color.NRGBA{0, 0, 255, 255}
		case "wait_background":
			return color.NRGBA{127, 127, 127, 180}
		}
	}
	return theme.Color(theme.ColorNameForeground)
}

func (p *PassSafeTheme) GetSpecialSize(s string) float32 {
	switch s {
	case "category_view_text_scale":
		return 1.0
	case "entry_view_title_scale":
		return 1.70
	case "entry_view_datetime_scale":
		return 0.75
	case "entry_view_category_scale":
		return 0.75
	case "entry_view_label_scale":
		return 1.0
	case "category_view_icon_size":
		return 36
	case "entry_view_icon_size":
		return 48
	case "entry_edit_icon_size":
		return 48
	case "icon_select_icon_size":
		return 64
	case "login_icon_size":
		return 128
	case "login_space_logo_label":
		return 3
	case "login_space_field_ok":
		return 1
	}
	return 1.0
}
