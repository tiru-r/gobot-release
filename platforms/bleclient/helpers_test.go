package bleclient

import (
	"context"
	"fmt"
	"time"

	"gobot.io/x/gobot/v2/bluetooth"
)

type btTestAdapter struct {
	deviceAddress       string
	rssi                int16
	scanDelay           time.Duration
	payload             *btTestPayload
	simulateEnableErr   bool
	simulateScanErr     bool
	simulateStopScanErr bool
	simulateConnectErr  bool
}

func (bta *btTestAdapter) Enable(ctx context.Context) error {
	if bta.simulateEnableErr {
		return fmt.Errorf("adapter enable error")
	}
	return nil
}

func (bta *btTestAdapter) Scan(ctx context.Context, params bluetooth.ScanParams, callback func(bluetooth.Advertisement)) error {
	if bta.simulateScanErr {
		return fmt.Errorf("adapter scan error")
	}

	time.Sleep(bta.scanDelay)

	// Parse MAC address manually
	var addr bluetooth.Address
	// Simple parsing for test address like "12:34:56:78:9A:BC"
	if len(bta.deviceAddress) == 17 {
		for i := range 6 {
			byteStr := bta.deviceAddress[i*3 : i*3+2]
			val := 0
			for _, c := range byteStr {
				val *= 16
				if c >= '0' && c <= '9' {
					val += int(c - '0')
				} else if c >= 'a' && c <= 'f' {
					val += int(c - 'a' + 10)
				} else if c >= 'A' && c <= 'F' {
					val += int(c - 'A' + 10)
				}
			}
			addr.MAC[5-i] = byte(val)
		}
	}

	adv := bluetooth.Advertisement{
		Address:   addr,
		RSSI:      bta.rssi,
		LocalName: bta.payload.name,
	}
	callback(adv)

	return nil
}

func (bta *btTestAdapter) StopScan(ctx context.Context) error {
	if bta.simulateStopScanErr {
		return fmt.Errorf("adapter stop scan error")
	}
	return nil
}

func (bta *btTestAdapter) Connect(ctx context.Context, addr bluetooth.Address, params bluetooth.ConnectionParams) (bluetooth.Device, error) {
	if bta.simulateConnectErr {
		return nil, fmt.Errorf("adapter connect error")
	}

	return &btTestDevice{}, nil
}

type btTestPayload struct {
	name string
}

func (ptp *btTestPayload) LocalName() string { return ptp.name }

func (*btTestPayload) HasServiceUUID(bluetooth.UUID) bool { return true }

func (*btTestPayload) Bytes() []byte { return nil }

func (*btTestPayload) ManufacturerData() map[uint16][]byte { return nil }

func (*btTestPayload) ServiceData() map[bluetooth.UUID][]byte { return nil }

type btTestDevice struct {
	simulateDiscoverServicesErr bool
	simulateDisconnectErr       bool
	services                    []bluetooth.Service
}

func (btd *btTestDevice) DiscoverServices(ctx context.Context, uuids []bluetooth.UUID) error {
	if btd.simulateDiscoverServicesErr {
		return fmt.Errorf("device discover services error")
	}
	return nil
}

func (btd *btTestDevice) Services() []bluetooth.Service {
	return btd.services
}

func (btd *btTestDevice) GetService(uuid bluetooth.UUID) (bluetooth.Service, error) {
	for _, service := range btd.services {
		if service.UUID() == uuid {
			return service, nil
		}
	}
	return nil, fmt.Errorf("service not found")
}

func (btd *btTestDevice) Disconnect(ctx context.Context) error {
	if btd.simulateDisconnectErr {
		return fmt.Errorf("device disconnect error")
	}
	return nil
}

func (btd *btTestDevice) Connected() bool {
	return true
}

func (btd *btTestDevice) Address() bluetooth.Address {
	return bluetooth.Address{}
}

func (btd *btTestDevice) Name() string {
	return "Test Device"
}

func (btd *btTestDevice) RSSI() int16 {
	return -60
}

func (btd *btTestDevice) RequestMTU(ctx context.Context, mtu uint16) error {
	return nil
}

func (btd *btTestDevice) GetMTU() uint16 {
	return 247
}

type btTestChara struct {
	readData         []byte
	writtenData      []byte
	notificationFunc func(buf []byte)
	uuid             bluetooth.UUID
}

func (btc *btTestChara) Read(ctx context.Context) ([]byte, error) {
	return btc.readData, nil
}

func (btc *btTestChara) Write(ctx context.Context, data []byte) error {
	btc.writtenData = append(btc.writtenData, data...)
	return nil
}

func (btc *btTestChara) WriteWithoutResponse(ctx context.Context, data []byte) error {
	btc.writtenData = append(btc.writtenData, data...)
	return nil
}

func (btc *btTestChara) Subscribe(ctx context.Context, callback func([]byte)) error {
	btc.notificationFunc = callback
	return nil
}

func (btc *btTestChara) Unsubscribe(ctx context.Context) error {
	btc.notificationFunc = nil
	return nil
}

func (btc *btTestChara) UUID() bluetooth.UUID {
	return btc.uuid
}
