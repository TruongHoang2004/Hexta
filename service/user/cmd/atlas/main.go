package main

import (
	"fmt"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"
	"gitlab.com/ecommercehub1/user/internal/core/model"
)

func main() {
	stmts, err := gormschema.New("postgres").Load(&model.User{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(stmts)
}
