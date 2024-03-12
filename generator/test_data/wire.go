//go:build wireinject
// +build wireinject

package test_data

import wire "github.com/google/wire"

var DefaultSet = wire.NewSet(MakeHTTPHandler)
