package setup

import (
	"github.com/fitan/gowrap/generator/test_data"
)

type Copygen interface {
	Struct2Struct(account *test_data.Account) *test_data.CopyAccount
}
