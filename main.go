package main

import (
	"github.com/alexflint/go-arg"
	"github.com/femnad/fup/base"
	"github.com/femnad/fup/provision"
	"log"
)

var args struct {
	File string `arg:"required,-f,--file"`
}

func main() {
	arg.MustParse(&args)
	config, err := base.ReadConfig(args.File)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	p := provision.Provisioner{Config: config}
	p.Apply()
}
