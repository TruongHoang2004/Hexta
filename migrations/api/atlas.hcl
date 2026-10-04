data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "github.com/TruongHoang2004/Hexta/services/api/cmd/tools"
  ]
}

env "gorm" {
  # Nguồn schema lấy từ GORM
  src = data.external_schema.gorm.url

  # Database dev (chỉ dùng để Atlas diff, không phải DB production)
  dev = "postgres://postgres:postgres@postgres:5432/dev?sslmode=disable"

  url = "postgres://postgres:postgres@postgres:5432/api?sslmode=disable"

  migration {
    dir = "file://migrations/api"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
