package main

import (
	"log"

	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
)

func onStateChanged(device gatt.Device, s gatt.State) {
	log.Println("State:", s)
	switch s {
	case gatt.StatePoweredOn:
		log.Println("Scanning for Broadcasts...")
		device.Scan([]gatt.UUID{}, true)
		return
	default:
		device.StopScanning()
	}
}

func onPeripheralDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	log.Printf("\nPeripheral ID:%s, NAME:(%s)\n", p.ID(), p.Name())
	log.Println("  Local Name        =", a.LocalName)
	log.Println("  TX Power Level    =", a.TxPowerLevel)
	log.Println("  Manufacturer Data =", a.ManufacturerData)
	log.Println("  Service Data      =", a.ServiceData)
}

func main() {
	device, err := gatt.NewDevice(option.DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}
	device.Handle(gatt.PeripheralDiscovered(onPeripheralDiscovered))
	device.Init(onStateChanged)
	select {}
}
