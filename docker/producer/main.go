package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	TOPIC                 = "topic/temps"
	QOS                   = 1
	SERVER_ADDRESS_BROKER = "tcp://localhost:1883"
	DELAY                 = 1 * time.Second
	CLIENT_ID             = "mqtt_producer"
	WRITETOLOG            = true
)

func main() {
	opts := mqtt.NewClientOptions()

	// Add mosquitto broker address
	//opts.AddBroker(SERVER_ADDRESS_MOSQ)
	// Add hive broker address
	opts.AddBroker(SERVER_ADDRESS_BROKER)
	// Add emqx broker address
	//opts.AddBroker(SERVER_ADDRESS_EMQX)
	opts.SetClientID(CLIENT_ID)

	opts.SetOrderMatters(true)            // Allow out of order messages
	opts.ConnectTimeout = 1 * time.Second // Minimal delays on connect
	opts.WriteTimeout = 1 * time.Second   // Minimal delays on writes
	opts.KeepAlive = 10                   // Keepalive every 10 seconds
	opts.PingTimeout = 1 * time.Second    // local broker so response should be quick

	// keep trying to connect and will reconnect if network drops
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)

	// Log events
	opts.OnConnect = func(client mqtt.Client) {
		log.Println("Connected.")
	}
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		fmt.Printf("Connection lost: %v.", err)
	}
	opts.OnReconnecting = func(mqtt.Client, *mqtt.ClientOptions) {
		fmt.Println("Attempting to reconnect...")
	}
	// Create the client using the options above
	client := mqtt.NewClient(opts)

	// Throw an error if the connection isn't successfull
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	fmt.Println("Connection is up.")

	// Publish messages until a signal is received
	publish(client)
}

func publish(client mqtt.Client) {
	done := make(chan struct{})
	var wg sync.WaitGroup

	type msg struct {
		Count uint64
		Temp  float64
	}

	wg.Add(1)
	go func() {
		var count uint64
		var temp float64
		for {
			select {
			case <-time.After(DELAY):
				count++
				temp = rand.Float64() + 3
				msg, err := json.Marshal(
					msg{
						Count: count,
						Temp:  temp,
					})
				if err != nil {
					panic(err)
				}
				if WRITETOLOG {
					fmt.Printf("Sending message: %s\n", msg)
				}
				t := client.Publish(TOPIC, QOS, false, msg)
				// Handle the token in a go routine so this loop keeps sending messages regardless of delivery status
				go func() {
					<-t.Done()
					if t.Error() != nil {
						fmt.Printf("Error publishing: %s\n", t.Error())
					}
				}()
			case <-done:
				fmt.Println("Producer done.")
				wg.Done()
				return
			}
		}
	}()
	// Use signals to exit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	<-sig
	fmt.Println("signal caught - exiting")

	close(done)
	wg.Wait()
	fmt.Println("Shutdown complete.")
}
