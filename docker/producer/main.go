package main

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

const (
	TOPIC                 = "floridis/sensor/temp"
	QOS                   = 1
	SERVER_ADDRESS_BROKER = "tcp://localhost:1883"
	DELAY                 = 1 * time.Second
	INITIAL_CLIENT_ID     = "mqtt_producer"
	WRITETOLOG            = true
	USERNAME              = "floridis"
	PASSWORD              = "password"
)

func main() {
	mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	mqtt.CRITICAL = log.New(os.Stdout, "[CRITICAL] ", 0)
	mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)
	// mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)
	opts := mqtt.NewClientOptions()

	opts.AddBroker(SERVER_ADDRESS_BROKER)

	uuid := uuid.New()
	clientId := INITIAL_CLIENT_ID + "_" + uuid.String()

	opts.SetClientID(clientId)
	opts.SetOrderMatters(true) // Allow out of order messages if set to false
	opts.SetUsername(USERNAME)
	opts.SetPassword(PASSWORD)
	opts.ConnectTimeout = 1 * time.Second // Minimal delays on connect
	opts.WriteTimeout = 1 * time.Second   // Minimal delays on writes
	opts.KeepAlive = 10                   // Keepalive every 10 seconds
	opts.PingTimeout = 1 * time.Second    // local broker so response should be quick

	tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}

	opts.SetTLSConfig(tlsConfig)
	// keep trying to connect and will reconnect if network drops
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)

	// Log events
	opts.OnConnect = func(client mqtt.Client) {
		log.Println("Connected.")
	}
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		log.Printf("Connection lost: %v.", err)
	}
	opts.OnReconnecting = func(mqtt.Client, *mqtt.ClientOptions) {
		log.Println("Attempting to reconnect...")
	}
	// Create the client using the options above
	client := mqtt.NewClient(opts)

	// Throw an error if the connection isn't successfull
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	log.Println("Connection is up.")

	// Publish messages until a signal is received
	publish(client, clientId)
}

func publish(client mqtt.Client, clientId string) {
	done := make(chan struct{})
	var wg sync.WaitGroup

	type msg struct {
		Count    uint64
		Temp     float64
		Date     time.Time
		ClientId string
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
						Count:    count,
						Temp:     temp,
						Date:     time.Now(),
						ClientId: clientId,
					})
				if err != nil {
					panic(err)
				}
				if WRITETOLOG {
					log.Printf("Sending message: %s\n", msg)
				}
				t := client.Publish(TOPIC, QOS, false, msg)
				// Handle the token in a go routine so this loop keeps sending messages regardless of delivery status
				go func() {
					<-t.Done()
					if t.Error() != nil {
						log.Printf("Error publishing: %s\n", t.Error())
					}
				}()
			case <-done:
				log.Println("Producer done.")
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
	log.Println("signal caught - exiting")

	close(done)
	wg.Wait()
	log.Println("Shutdown complete.")
}
