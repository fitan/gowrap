package setup

import (
	"github.com/fitan/gowrap/generator/test_data"
)

// goverter:converter
type Copygen interface {
	//ModelsToDomain(nest.Account) test_data.Account
	//SliceModelsToDomain([]nest.Account) []test_data.Account
	Struct2Struct(account test_data.NestAccount) test_data.TestAccount
}
