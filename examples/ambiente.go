package main

import (
	"fmt"
	"os"
	"colors"
)

var p = fmt.Println

func main() {
	// Obtém o valor da variável de ambiente "PATH"
	path := os.Getenv("PATH")
	shell := os.Getenv("SHELL")
	p("PATH               :", path)
	p("SHELL              :", shell)
	p("USER               :", os.Getenv("USER"))
	p( colors.Red + "XDG_SESSION_DESKTOP:", colors.Cyan + os.Getenv("XDG_SESSION_DESKTOP"))
	p( colors.Red + "XDG_SESSION_TYPE   :", colors.Cyan + os.Getenv("XDG_SESSION_TYPE"))

}
