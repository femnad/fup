package main

import (
	"github.com/alexflint/go-arg"
	"github.com/femnad/fup/base"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/provision"
	"log"
)

var args struct {
	File     string `arg:"-f,--file" default:"~/.config/fup/fup.yml"`
	LogLevel int    `arg:"-l,--loglevel" default:"4"`
}

func main() {
	arg.MustParse(&args)
	internal.InitLogging(args.LogLevel)
	config, err := base.ReadConfig(args.File)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	p := provision.Provisioner{Config: config}
	p.Apply()
}
