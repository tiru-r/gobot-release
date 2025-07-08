package bleclient

import (
	"context"
	"fmt"
	"time"

	"gobot.io/x/gobot/v2/bluetooth"
)

// bluetoothExtDevicer is the interface usually implemented by bluetooth.Device
type bluetoothExtDevicer interface {
	DiscoverServices(ctx context.Context, uuids []bluetooth.UUID) error
	Services() []bluetooth.Service
	GetService(uuid bluetooth.UUID) (bluetooth.Service, error)
	Disconnect(ctx context.Context) error
	Connected() bool
	Address() bluetooth.Address
	Name() string
}

// bluetoothExtAdapterer is the interface usually implemented by bluetooth.Central
type bluetoothExtAdapterer interface {
	Enable(ctx context.Context) error
	Scan(ctx context.Context, params bluetooth.ScanParams, callback func(bluetooth.Advertisement)) error
	StopScan(ctx context.Context) error
	Connect(ctx context.Context, address bluetooth.Address, params bluetooth.ConnectionParams) (bluetooth.Device, error)
}

type bluetoothExtCharacteristicer interface {
	Read(ctx context.Context) ([]byte, error)
	Write(ctx context.Context, data []byte) error
	WriteWithoutResponse(ctx context.Context, data []byte) error
	Subscribe(ctx context.Context, callback func([]byte)) error
	Unsubscribe(ctx context.Context) error
	UUID() bluetooth.UUID
}

// btAdptCreatorFunc is just a convenience type, used in the BLE client to ensure testability
type btAdptCreatorFunc func(bluetoothExtAdapterer, bool) *btAdapter

// btAdapter is the wrapper for an external adapter implementation
type btAdapter struct {
	extAdapter      bluetoothExtAdapterer
	btDeviceCreator func(bluetoothExtDevicer, string, string) *btDevice
	debug           bool
}

// newBtAdapter creates a new wrapper around the given external implementation
func newBtAdapter(a bluetoothExtAdapterer, debug bool) *btAdapter {
	bta := btAdapter{
		extAdapter:      a,
		btDeviceCreator: newBtDevice,
		debug:           debug,
	}

	return &bta
}

// Enable configures the BLE stack. It must be called before any Bluetooth-related calls (unless otherwise indicated).
// It pass through the function of the external implementation.
func (bta *btAdapter) enable() error {
	ctx := context.Background()
	return bta.extAdapter.Enable(ctx)
}

// StopScan stops any in-progress scan. It can be called from within a Scan callback to stop the current scan.
// If no scan is in progress, an error will be returned.
func (bta *btAdapter) stopScan() error {
	ctx := context.Background()
	return bta.extAdapter.StopScan(ctx)
}

// Connect starts a connection attempt to the given peripheral device address.
//
// On Linux and Windows, the IsRandom part of the address is ignored.
func (bta *btAdapter) connect(address bluetooth.Address, devName string) (*btDevice, error) {
	ctx := context.Background()
	extDev, err := bta.extAdapter.Connect(ctx, address, bluetooth.DefaultConnectionParams())
	if err != nil {
		return nil, err
	}

	return bta.btDeviceCreator(extDev, address.String(), devName), nil
}

// Scan starts a BLE scan for the given identifier (address or name).
func (bta *btAdapter) scan(identifier string, scanTimeout time.Duration) (*bluetooth.Advertisement, error) {
	resultChan := make(chan bluetooth.Advertisement, 1)
	errChan := make(chan error)

	ctx, cancel := context.WithTimeout(context.Background(), scanTimeout)
	defer cancel()

	go func() {
		params := bluetooth.DefaultScanParams()
		params.Timeout = scanTimeout
		
		callback := func(adv bluetooth.Advertisement) {
			if bta.debug {
				fmt.Printf("[scan result]: address: '%s', rssi: %d, name: '%s'\n",
					adv.Address.String(), adv.RSSI, adv.LocalName)
			}
			if adv.Address.String() == identifier || adv.LocalName == identifier {
				resultChan <- adv
			}
		}
		
		err := bta.extAdapter.Scan(ctx, params, callback)
		if err != nil {
			errChan <- err
		}
	}()

	select {
	case result := <-resultChan:
		if err := bta.stopScan(); err != nil {
			return nil, err
		}
		return &result, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		_ = bta.stopScan()
		return nil, fmt.Errorf("scan timeout (%s) elapsed", scanTimeout)
	}
}

// btDevice is the wrapper for an external device implementation
type btDevice struct {
	extDevice  bluetoothExtDevicer
	devAddress string
	devName    string
}

// newBtDevice creates a new wrapper around the given external implementation
func newBtDevice(d bluetoothExtDevicer, address, name string) *btDevice {
	return &btDevice{extDevice: d, devAddress: address, devName: name}
}

func (btd *btDevice) name() string { return btd.devName }

func (btd *btDevice) address() string { return btd.devAddress }

func (btd *btDevice) discoverServices(uuids []bluetooth.UUID) ([]bluetooth.Service, error) {
	ctx := context.Background()
	err := btd.extDevice.DiscoverServices(ctx, uuids)
	if err != nil {
		return nil, err
	}
	return btd.extDevice.Services(), nil
}

// Disconnect from the BLE device. This method is non-blocking and does not wait until the connection is fully gone.
func (btd *btDevice) disconnect() error {
	ctx := context.Background()
	return btd.extDevice.Disconnect(ctx)
}

func readFromCharacteristic(chara bluetoothExtCharacteristicer) ([]byte, error) {
	ctx := context.Background()
	return chara.Read(ctx)
}

func writeToCharacteristicWithoutResponse(chara bluetoothExtCharacteristicer, data []byte) error {
	ctx := context.Background()
	return chara.WriteWithoutResponse(ctx, data)
}

func enableNotificationsForCharacteristic(chara bluetoothExtCharacteristicer, f func(data []byte)) error {
	ctx := context.Background()
	return chara.Subscribe(ctx, f)
}
