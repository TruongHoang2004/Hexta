data "external_schema" "gorm" {
  program = [
    "go", "run",
    "ariga.io/atlas-provider-gorm", "load",
    "--path", "../../service/catalog/internal/core/model",
    "--dialect", "postgres"
  ]
}

env "gorm" {
  # Nguồn schema lấy từ GORM
  src = data.external_schema.gorm.url

  # Database dev (chỉ dùng để Atlas diff, không phải DB production)
  dev = "postgres://postgres:postgres@postgres:5432/dev?sslmode=disable"

  url = "postgres://postgres:postgres@postgres:5432/catalog?sslmode=disable"

  migration {
    dir = "file://."
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
