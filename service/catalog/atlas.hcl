data "external_schema" "gorm" {
  program = [
    "go", "run", "-mod=mod",
    "ariga.io/atlas-provider-gorm", "load",
    "--path", "./internal/core/model",
    "--dialect", "postgres"
  ]
}

env "gorm" {
  # Nguồn schema lấy từ GORM
  src = data.external_schema.gorm.url

  # Database dev (chỉ dùng để Atlas diff, không phải DB production)
  dev = "postgres://postgres:postgres@localhost:5433/dev?sslmode=disable"

  url = "postgres://postgres:postgres@localhost:5433/catalog?sslmode=disable"

  migration {
    dir = "file://internal/infrastructure/database/migrations"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
