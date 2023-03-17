module github.com/fitan/gowrap

require (
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/dave/jennifer v1.6.0 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/fitan/jennifer v0.0.0-20221025094417-113be729db13
	github.com/go-kit/kit v0.9.0
	github.com/go-logfmt/logfmt v0.4.0 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gobeam/stringy v0.0.5
	github.com/gorilla/mux v1.8.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cast v1.5.0
	github.com/spf13/cobra v1.6.0
	github.com/stretchr/testify v1.8.0 // indirect
	go.opentelemetry.io/otel v1.0.1
	go.opentelemetry.io/otel/trace v1.0.1
	golang.org/x/tools v0.1.11
	gorm.io/gorm v1.24.0
)

replace github.com/fitan/jennifer => ../jennifer

go 1.16
