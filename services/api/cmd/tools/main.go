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
		&model.Tenant{},
		&model.TenantMember{},
		&model.Product{},
		&model.ProductVariant{},
		&model.InventoryItem{},
		&model.StockMovement{},
		&model.Customer{},
		&model.Order{},
		&model.OrderItem{},
		&model.AuditLog{},
		&model.AIDraft{},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}
	io.WriteString(os.Stdout, stmts)
}
