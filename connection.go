package gobot

import (
	"context"
	"log/slog"
	"reflect"
)

// JSONConnection is a JSON representation of a Connection.
type JSONConnection struct {
	Name    string `json:"name"`
	Adaptor string `json:"adaptor"`
}

// NewJSONConnection returns a JSONConnection given a Connection.
func NewJSONConnection(connection Connection) *JSONConnection {
	return &JSONConnection{
		Name:    connection.Name(),
		Adaptor: reflect.TypeOf(connection).String(),
	}
}

// A Connection is an instance of an Adaptor
type Connection Adaptor

// Connections represents a collection of Connection
type Connections []Connection

// Len returns connections length
func (c *Connections) Len() int {
	return len(*c)
}

// Each enumerates through the Connections and calls specified callback function.
func (c *Connections) Each(f func(Connection)) {
	for _, connection := range *c {
		f(connection)
	}
}

// All returns an iterator over all connections using range-over-func
func (c *Connections) All() func(func(Connection) bool) {
	return func(yield func(Connection) bool) {
		for _, connection := range *c {
			if !yield(connection) {
				return
			}
		}
	}
}

// Start calls Connect on each Connection in c
func (c *Connections) Start() error {
	slog.Info("Starting connections...")
	var err error
	for _, connection := range *c {
		attrs := []slog.Attr{
			slog.String("connection", connection.Name()),
		}

		if porter, ok := connection.(Porter); ok {
			attrs = append(attrs, slog.String("port", porter.Port()))
		}

		slog.LogAttrs(context.TODO(), slog.LevelInfo, "Starting connection", attrs...)

		if cerr := connection.Connect(); cerr != nil {
			err = AppendError(err, cerr)
		}
	}
	return err
}

// Finalize calls Finalize on each Connection in c
func (c *Connections) Finalize() error {
	var err error
	for _, connection := range *c {
		if cerr := connection.Finalize(); cerr != nil {
			err = AppendError(err, cerr)
		}
	}
	return err
}
