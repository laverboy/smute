package main

import (
	"github.com/laverboy/smute/app"
)

func main() {
	app.CLI([]string{"", "github.com/laverboy/smute-templates", "basic", "/tmp/basic"})
}
