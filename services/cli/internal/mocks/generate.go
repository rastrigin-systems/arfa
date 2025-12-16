// Package mocks provides mock implementations for testing.
//
// To regenerate mocks, run:
//
//	go generate ./...
//
// or from the services/cli directory:
//
//	make mocks
package mocks

//go:generate go run go.uber.org/mock/mockgen -source=../interfaces.go -destination=interfaces_mock.go -package=mocks
