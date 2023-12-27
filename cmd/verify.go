package main

import (
	"log"

	"github.com/alexflint/go-arg"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/cmd/verify"
)

type args struct {
	File      string `arg:"required,-f,--file" help:"Verification config file path"`
	FupConfig string `arg:"required,-c,--fup-config" help:"Fup config file path"`
}

func main() {
	var parsed args
	arg.MustParse(&parsed)

	fupConfig, err := base.ReadConfig(parsed.FupConfig)
	if err != nil {
		log.Fatalf("error reading fup config: %v", err)
	}

	err = verify.Verify(parsed.File, fupConfig)
	if err != nil {
		log.Fatalf("Error during verification: %v\n", err)
	}
}
