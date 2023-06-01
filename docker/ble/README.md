# BLE Signal Scanner on Raspberry Pi 

## Commands

Per the Gatt instructions, Gatt needs complete control of the bluetooth. For this reason, we need to disable functionality within the OS. The following commands will bring down the bluetooth device and stop the built-in bluetooth server:

* `sudo hciconfig hci0 down`
* `sudo service bluetooth stop`

Then, execute the scanner:

* `cd /ble`
* `sudo go run main.go`
