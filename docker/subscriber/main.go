package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

const (
	TOPIC                 = "floridis/sensor/temp"
	QOS                   = 1
	SERVER_ADDRESS_BROKER = "tcp://localhost:1883"
	INITIAL_CLIENT_ID     = "mqtt_subscriber"
	WRITETOLOG            = true
	WRITETODISK           = false
	OUTPUTFILE            = "/binds/receivedMessages.txt"
	USERNAME              = "floridis"
	PASSWORD              = "password"
)

type handler struct {
	f *os.File
}

// handler is a simple struct that provides a function to be called when a message is received.
//The message is parsed and the count followed by the raw message is written to the file
func NewHandler() *handler {
	var f *os.File
	if WRITETODISK {
		var err error
		f, err = os.Create(OUTPUTFILE)
		if err != nil {
			panic(err)
		}
	}
	return &handler{f: f}
}

func (o *handler) Close() {
	if o.f != nil {
		if err := o.f.Close(); err != nil {
			log.Printf("Error occured when trying to close the file: %s", err)
		}
		o.f = nil
	}
}

type Message struct {
	Count    uint64
	Temp     float64
	Date     time.Time
	ClientId string
}

func (o *handler) handle(_ mqtt.Client, msg mqtt.Message) {
	var m Message
	if err := json.Unmarshal(msg.Payload(), &m); err != nil {
		log.Printf("Message could not be parsed (%s): %s", msg.Payload(), err)
	}
	if o.f != nil {
		if _, err := o.f.WriteString(fmt.Sprintf("%09d %s\n", m.Count, msg.Payload())); err != nil {
			log.Printf("Error writing to file: %s", err)
		}
	}
	if WRITETOLOG {
		log.Printf("Received message with id=%v and payload:\n %s\n", msg.MessageID(), msg.Payload())
	}
}

func main() {
	mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	mqtt.CRITICAL = log.New(os.Stdout, "[CRITICAL] ", 0)
	mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)
	// mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)
	h := NewHandler()
	defer h.Close()

	opts := mqtt.NewClientOptions()
	opts.AddBroker(SERVER_ADDRESS_BROKER)

	uuid := uuid.New()
	clientId := INITIAL_CLIENT_ID + "_" + uuid.String()

	opts.SetClientID(clientId)
	opts.SetOrderMatters(true) // Allow out of order messages if set to false
	opts.SetUsername(USERNAME)
	opts.SetPassword(PASSWORD)
	opts.ConnectTimeout = time.Second // Minimal delays on connect
	opts.WriteTimeout = time.Second   // Minimal delays on writes
	opts.KeepAlive = 10               // Keepalive every 10 seconds
	opts.PingTimeout = time.Second    // local broker so response should be quick

	tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}

	opts.SetTLSConfig(tlsConfig)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)

	// Log events
	opts.DefaultPublishHandler = func(_ mqtt.Client, msg mqtt.Message) {
		log.Printf("Unexpected message: %s\n", msg)
	}
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		log.Printf("Connection lost: %v.", err)
	}
	opts.OnConnect = func(client mqtt.Client) {
		log.Println("Connected.")
		t := client.Subscribe(TOPIC, QOS, h.handle)
		go func() {
			<-t.Done()
			if t.Error() != nil {
				log.Printf("Error subscribing: %s\n", t.Error())
			} else {
				log.Println("Subscribed to:", TOPIC)
			}
		}()
	}

	opts.OnReconnecting = func(mqtt.Client, *mqtt.ClientOptions) {
		log.Println("Attempting to reconnect...")
	}
	// create the client using the options above
	client := mqtt.NewClient(opts)
	client.AddRoute(TOPIC, h.handle)
	// throw an error if the connection isn't successfull
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	log.Println("Connection is up.")
	// Use signals to exit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	<-sig
	log.Println("signal caught - exiting")
	client.Disconnect(1000)
	log.Println("Shutdown complete.")
}
