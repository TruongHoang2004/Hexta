package main

import (
	"fmt"
	"io"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
)

func main() {
	stmts, err := gormschema.New("postgres").Load(
		&model.AuthIdentities{},
		&model.Sessions{},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}
	io.WriteString(os.Stdout, stmts)
}
