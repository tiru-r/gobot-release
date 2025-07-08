package gobot

import (
	"context"
	"log/slog"
	"reflect"
)

// JSONDevice is a JSON representation of a Device.
type JSONDevice struct {
	Name       string   `json:"name"`
	Driver     string   `json:"driver"`
	Connection string   `json:"connection"`
	Commands   []string `json:"commands"`
}

// NewJSONDevice returns a JSONDevice given a Device.
func NewJSONDevice(device Device) *JSONDevice {
	jsonDevice := &JSONDevice{
		Name:       device.Name(),
		Driver:     reflect.TypeOf(device).String(),
		Commands:   []string{},
		Connection: "",
	}
	if device.Connection() != nil {
		jsonDevice.Connection = device.Connection().Name()
	}
	if commander, ok := device.(Commander); ok {
		for command := range commander.Commands() {
			jsonDevice.Commands = append(jsonDevice.Commands, command)
		}
	}
	return jsonDevice
}

// A Device is an instnace of a Driver
type Device Driver

// Devices represents a collection of Device
type Devices []Device

// Len returns devices length
func (d *Devices) Len() int {
	return len(*d)
}

// Each enumerates through the Devices and calls specified callback function.
func (d *Devices) Each(f func(Device)) {
	for _, device := range *d {
		f(device)
	}
}

// All returns an iterator over all devices using range-over-func
func (d *Devices) All() func(func(Device) bool) {
	return func(yield func(Device) bool) {
		for _, device := range *d {
			if !yield(device) {
				return
			}
		}
	}
}

// Start calls Start on each Device in d
func (d *Devices) Start() error {
	slog.Info("Starting devices...")
	var err error
	for _, device := range *d {
		attrs := []slog.Attr{
			slog.String("device", device.Name()),
		}

		if pinner, ok := device.(Pinner); ok {
			attrs = append(attrs, slog.String("pin", pinner.Pin()))
		}

		slog.LogAttrs(context.TODO(), slog.LevelInfo, "Starting device", attrs...)
		if derr := device.Start(); derr != nil {
			err = AppendError(err, derr)
		}
	}
	return err
}

// Halt calls Halt on each Device in d
func (d *Devices) Halt() error {
	var err error
	for _, device := range *d {
		if derr := device.Halt(); derr != nil {
			err = AppendError(err, derr)
		}
	}
	return err
}
