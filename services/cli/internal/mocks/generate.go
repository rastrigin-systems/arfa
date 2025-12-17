// Package mocks provides mock implementations for testing.
//
// To regenerate mocks, run:
//
//	go generate ./...
//
// or from the services/cli directory:
//
//	make mocks
//
// Note: We generate from sync/interfaces.go because it contains all the
// dependency interfaces (ConfigManagerInterface, APIClientInterface, etc.)
// that are used across tests. This avoids name collisions from generating
// mocks from multiple packages that define interfaces with the same names.
package mocks

//go:generate go run go.uber.org/mock/mockgen -source=../sync/interfaces.go -destination=interfaces_mock.go -package=mocks
