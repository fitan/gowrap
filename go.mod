module github.com/fitan/gowrap

require (
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/davecgh/go-spew v1.1.1
	github.com/fitan/jennifer v0.0.0-20221025094417-113be729db13
	github.com/go-kit/kit v0.9.0
	github.com/gobeam/stringy v0.0.5
	github.com/google/wire v0.5.0
	github.com/gorilla/mux v1.8.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cast v1.5.0
	github.com/spf13/cobra v1.6.0
	go.opentelemetry.io/otel v1.0.1
	go.opentelemetry.io/otel/trace v1.0.1
	golang.org/x/tools v0.1.11
	gorm.io/gorm v1.24.0
)

require (
	github.com/dave/jennifer v1.6.0 // indirect
	github.com/go-logfmt/logfmt v0.4.0 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.4 // indirect
	github.com/kr/logfmt v0.0.0-20140226030751-b84e30acd515 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.8.0 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/sys v0.0.0-20211019181941-9d821ace8654 // indirect
)

//replace github.com/fitan/jennifer => ../jennifer

go 1.20
