package main

import (
	"fmt"
	"os"
	"colors/colors"
)

var p = fmt.Println

func main() {
	// Obtém o valor da variável de ambiente "PATH"
	path := os.Getenv("PATH")
	shell := os.Getenv("SHELL")
	p("PATH               :", path)
	p("SHELL              :", shell)
	p("USER               :", os.Getenv("USER"))
	p("XDG_SESSION_DESKTOP:", os.Getenv("XDG_SESSION_DESKTOP"))
	p( Red + "XDG_SESSION_TYPE   :", os.Getenv("XDG_SESSION_TYPE"))

}
