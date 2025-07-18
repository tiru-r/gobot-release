/*
 * Copyright (c) 2016-2017 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package i2c

// SHT2xDriver is a driver for the SHT2x based devices.
//
// This module was tested with Sensirion SHT21 Breakout.

import (
	"errors"
	"time"
)

const sht2xDefaultAddress = 0x40

const (
	// SHT2xAccuracyLow is the faster, but lower accuracy sample setting
	//  0/1 = 8bit RH, 12bit Temp
	SHT2xAccuracyLow = byte(0x01)

	// SHT2xAccuracyMedium is the medium accuracy and speed sample setting
	//  1/0 = 10bit RH, 13bit Temp
	SHT2xAccuracyMedium = byte(0x80)

	// SHT2xAccuracyHigh is the high accuracy and slowest sample setting
	//  0/0 = 12bit RH, 14bit Temp
	//  Power on default is 0/0
	SHT2xAccuracyHigh = byte(0x00)

	// SHT2xTriggerTempMeasureHold is the command for measuring temperature in hold controller mode
	SHT2xTriggerTempMeasureHold = 0xe3

	// SHT2xTriggerHumdMeasureHold is the command for measuring humidity in hold controller mode
	SHT2xTriggerHumdMeasureHold = 0xe5

	// SHT2xTriggerTempMeasureNohold is the command for measuring humidity in no hold controller mode
	SHT2xTriggerTempMeasureNohold = 0xf3

	// SHT2xTriggerHumdMeasureNohold is the command for measuring humidity in no hold controller mode
	SHT2xTriggerHumdMeasureNohold = 0xf5

	// SHT2xWriteUserReg is the command for writing user register
	SHT2xWriteUserReg = 0xe6

	// SHT2xReadUserReg is the command for reading user register
	SHT2xReadUserReg = 0xe7

	// SHT2xReadUserReg is the command for reading user register
	SHT2xSoftReset = 0xfe
)

// SHT2xDriver is a Driver for a SHT2x humidity and temperature sensor
type SHT2xDriver struct {
	*Driver
	Units    string
	accuracy byte
	crcTable []byte
}

// NewSHT2xDriver creates a new driver with specified i2c interface
// Params:
//
//	c Connector - the Adaptor to use with this Driver
//
// Optional params:
//
//	i2c.WithBus(int):	bus to use with this driver
//	i2c.WithAddress(int):	address to use with this driver
func NewSHT2xDriver(c Connector, options ...func(Config)) *SHT2xDriver {
	// From the document "CRC Checksum Calculation -- For Safe Communication with SHT2x Sensors":
	// CRC-8/SENSIRION-SHT2x with polynomial 0x31, init 0x00, no reflection, no xor
	d := &SHT2xDriver{
		Driver:   NewDriver(c, "SHT2x", sht2xDefaultAddress),
		Units:    "C",
		crcTable: makeCRC8Table(0x31),
	}
	d.afterStart = d.initialize

	for _, option := range options {
		option(d)
	}

	return d
}

func (d *SHT2xDriver) Accuracy() byte { return d.accuracy }

// SetAccuracy sets the accuracy of the sampling
func (d *SHT2xDriver) SetAccuracy(acc byte) error {
	d.accuracy = acc

	if d.connection != nil {
		return d.sendAccuracy()
	}

	return nil
}

// Reset does a software reset of the device
func (d *SHT2xDriver) Reset() error {
	if err := d.connection.WriteByte(SHT2xSoftReset); err != nil {
		return err
	}

	time.Sleep(15 * time.Millisecond) // 15ms delay (from the datasheet 5.5)

	return nil
}

// Temperature returns the current temperature, in celsius degrees.
func (d *SHT2xDriver) Temperature() (float32, error) {
	rawT, err := d.readSensor(SHT2xTriggerTempMeasureNohold)
	if err != nil {
		return 0, err
	}

	// From the datasheet 6.2:
	// T[C] = -46.85 + 175.72 * St / 2^16
	temp := -46.85 + 175.72/65536.0*float32(rawT)

	return temp, nil
}

// Humidity returns the current humidity in percentage of relative humidity
func (d *SHT2xDriver) Humidity() (float32, error) {
	rawH, err := d.readSensor(SHT2xTriggerHumdMeasureNohold)
	if err != nil {
		return 0, err
	}

	// From the datasheet 6.1:
	// RH = -6 + 125 * Srh / 2^16
	humidity := -6.0 + 125.0/65536.0*float32(rawH)

	return humidity, nil
}

// sendCommandDelayGetResponse is a helper function to reduce duplicated code
func (d *SHT2xDriver) readSensor(cmd byte) (uint16, error) {
	if err := d.connection.WriteByte(cmd); err != nil {
		return 0, err
	}

	// Hang out while measurement is taken. 85ms max, page 9 of datasheet.
	time.Sleep(85 * time.Millisecond)

	// Comes back in three bytes, data(MSB) / data(LSB) / Checksum
	buf := make([]byte, 3)
	counter := 0
	for {
		got, err := d.connection.Read(buf)
		counter++
		if counter > 50 {
			return 0, err
		}
		if err == nil {
			if got != 3 {
				return 0, ErrNotEnoughBytes
			}
			break
		}
		time.Sleep(1 * time.Millisecond)
	}

	// Store the result
	crc := crc8Checksum(buf[0:2], d.crcTable)
	if buf[2] != crc {
		return 0, errors.New("Invalid crc")
	}
	read := uint16(buf[0])<<8 | uint16(buf[1])
	read &= 0xfffc // clear two low bits (status bits)

	return read, nil
}

func (d *SHT2xDriver) initialize() error {
	if err := d.Reset(); err != nil {
		return err
	}

	return d.sendAccuracy()
}

func (d *SHT2xDriver) sendAccuracy() error {
	if err := d.connection.WriteByte(SHT2xReadUserReg); err != nil {
		return err
	}
	userRegister, err := d.connection.ReadByte()
	if err != nil {
		return err
	}

	userRegister &= 0x7e // Turn off the resolution bits
	acc := d.accuracy
	acc &= 0x81         // Turn off all other bits but resolution bits
	userRegister |= acc // Mask in the requested resolution bits

	// Request a write to user register
	if _, err := d.connection.Write([]byte{SHT2xWriteUserReg, userRegister}); err != nil {
		return err
	}

	_, err = d.connection.ReadByte()
	return err
}

// makeCRC8Table creates a CRC8 lookup table for the given polynomial
func makeCRC8Table(poly byte) []byte {
	table := make([]byte, 256)
	for i := range 256 {
		crc := byte(i)
		for range 8 {
			if crc&0x80 != 0 {
				crc = (crc << 1) ^ poly
			} else {
				crc = crc << 1
			}
		}
		table[i] = crc
	}
	return table
}

// crc8Checksum calculates CRC8 checksum for the given data using the lookup table
func crc8Checksum(data []byte, table []byte) byte {
	crc := byte(0x00)
	for _, b := range data {
		crc = table[crc^b]
	}
	return crc
}
