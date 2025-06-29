package cmd

import (
	"fmt"
)

const (
	reset       = "\033[0m"
	lightOrange = "\033[38;5;215m"
	orange      = "\033[38;5;209m"
	darkOrange  = "\033[38;5;203m"
	lightRed    = "\033[38;5;197m"
	red         = "\033[38;5;196m"
	darkRed     = "\033[38;5;124m"

	peach       = "\033[38;2;255;195;160m" // #FFC3A0
	lightCoral  = "\033[38;2;255;154;125m" // #FF9A7D
	coral       = "\033[38;2;255;107;107m" // #FF6B6B
	salmon      = "\033[38;2;255;78;80m"   // #FF4E50
	crimson     = "\033[38;2;200;50;70m"   // #C83246
	darkCrimson = "\033[38;2;150;30;50m"   //rgb(179, 57, 77)
)

const shortArangoCliLogo = peach + `
 █████╗ ██████╗  █████╗ ███╗   ██╗ ██████╗  ██████╗ ` + lightCoral + `      ██████╗██╗     ██╗
` + lightCoral + `██╔══██╗██╔══██╗██╔══██╗████╗  ██║██╔════╝ ██╔═══██╗` + coral + `     ██╔════╝██║     ██║
` + coral + `███████║██████╔╝███████║██╔██╗ ██║██║  ███╗██║   ██║` + salmon + `     ██║     ██║     ██║
` + salmon + `██╔══██║██╔══██╗██╔══██║██║╚██╗██║██║   ██║██║   ██║` + crimson + `     ██║     ██║     ██║
` + crimson + `██║  ██║██║  ██║██║  ██║██║ ╚████║╚██████╔╝╚██████╔╝` + darkCrimson + `     ╚██████╗███████╗██║
` + darkCrimson + `╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝  ╚═════╝ ` + darkCrimson + `      ╚═════╝╚══════╝╚═╝
` + reset

func PrintBanner() {
	fmt.Println(shortArangoCliLogo)
}
