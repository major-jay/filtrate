package main

import (
	"fmt"

	"github.com/major-jay/filtrate/filtrate"
)

func main() {
	filtrate.InitTree("./config/sensitiveWord")
	word := filtrate.Search("你sssdsfjs一貫道kla说夜总会新义安ss")
	fmt.Print(word)
	return
}
