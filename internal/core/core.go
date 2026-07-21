// Package core provides application level business logic
package core

type Database interface{}

type Core struct {
	storage Database
}
