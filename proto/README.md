# AI service contract generation

Run from the repository root. Checked-in bindings allow ordinary builds without code generation. Never hand-edit generated files.

Python uses the locked grpcio-tools in `services/ai/uv.lock`:

```sh
(cd services/ai && uv sync --frozen)
services/ai/.venv/bin/python -m grpc_tools.protoc -I proto --python_out=services/ai/generated --grpc_python_out=services/ai/generated proto/hexta/ai/v1/service.proto
```

Go generation used protoc 34.1, protoc-gen-go v1.36.11 and protoc-gen-go-grpc 1.6.1. Install these Go plugins if regeneration is needed:

```sh
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.1
PATH="$(go env GOPATH)/bin:$PATH" protoc -I proto --go_out=services/api --go_opt=module=github.com/TruongHoang2004/Hexta/services/api --go-grpc_out=services/api --go-grpc_opt=module=github.com/TruongHoang2004/Hexta/services/api proto/hexta/ai/v1/service.proto
```

Add compatible fields with new field numbers; reserve removed names/numbers. Introduce a new package version for breaking changes. The v1 contract currently covers offline script simulation; richer provider/run metadata will require additive fields before live integration.
