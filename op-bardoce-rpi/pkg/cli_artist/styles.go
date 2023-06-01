package cli_artist

import "gopkg.in/gookit/color.v1"

func BoldRed() color.Style {
	return color.Style{color.FgRed, color.OpBold}
}
func ItalicWhite() color.Style {
	return color.Style{color.FgLightWhite, color.OpItalic}
}
