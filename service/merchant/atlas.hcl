data "external_schema" "gorm" {
  program = [
    "go", "run", "-mod=mod",
    "ariga.io/atlas-provider-gorm", "load",
    "--path", "./internal/core/model",
    "--dialect", "postgres"
  ]
}

env "gorm" {
  src = data.external_schema.gorm.url

  dev = "postgres://postgres:postgres@localhost:5433/dev?sslmode=disable"

  url = "postgres://postgres:postgres@localhost:5433/merchant?sslmode=disable"

  migration {
    dir = "file://internal/infrastructure/database/migrations"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
