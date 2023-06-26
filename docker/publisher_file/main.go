package main

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
	INITIAL_CLIENT_ID     = "mqtt_publisher"
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
		Temp float64 `json:"temperature"`
	}

	data, err := ioutil.ReadFile("a.txt")

	if err != nil {
		log.Println("File reading error", err)
		return
	}

	trimmedLeft := strings.TrimLeft(string(data), "{")

	trimmedRight := strings.TrimRight(trimmedLeft, "}")

	temps := strings.Split(trimmedRight, "}{")

	wg.Add(1)
	go func() {
		for _, v := range temps {
			select {
			case <-time.After(DELAY):

				n := strings.ReplaceAll(v, "'temperature': ", "")

				var temp float64

				temp, err = strconv.ParseFloat(n, 8)

				if err != nil {
					log.Println(err)
				}

				msg, err := json.Marshal(
					msg{
						Temp: temp,
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
				log.Println("Publisher done.")
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
