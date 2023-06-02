package cli_artist

import (
	"fmt"
	"gopkg.in/gookit/color.v1"
)

func PrintStyledText(style color.Style, text string) {
	style.Print(text)
}

func PrintSpaces(numSpaces int) {
	for i := 0; i < numSpaces; i++ {
		fmt.Print("\n")
	}
}

func ClearTerminal() {
	fmt.Print("\033[H\033[2J")
}
