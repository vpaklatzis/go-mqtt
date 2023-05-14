package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	TOPIC                 = "topic/temps"
	QOS                   = 1
	SERVER_ADDRESS_BROKER = "tcp://localhost:1883"
	CLIENT_ID             = "mqtt_consumer"
	WRITETOLOG            = true
	WRITETODISK           = false
	OUTPUTFILE            = "/binds/receivedMessages.txt"
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
			fmt.Printf("Error occured when trying to close the file: %s", err)
		}
		o.f = nil
	}
}

type Message struct {
	Count uint64
	Temp  float64
}

func (o *handler) handle(_ mqtt.Client, msg mqtt.Message) {
	var m Message
	if err := json.Unmarshal(msg.Payload(), &m); err != nil {
		fmt.Printf("Message could not be parsed (%s): %s", msg.Payload(), err)
	}
	if o.f != nil {
		if _, err := o.f.WriteString(fmt.Sprintf("%09d %s\n", m.Count, msg.Payload())); err != nil {
			fmt.Printf("Error writing to file: %s", err)
		}
	}
	if WRITETOLOG {
		fmt.Printf("Received message: %s\n", msg.Payload())
	}
}

func main() {
	h := NewHandler()
	defer h.Close()

	opts := mqtt.NewClientOptions()
	opts.AddBroker(SERVER_ADDRESS_BROKER)
	opts.SetClientID(CLIENT_ID)

	opts.SetOrderMatters(true)        // Allow out of order messages
	opts.ConnectTimeout = time.Second // Minimal delays on connect
	opts.WriteTimeout = time.Second   // Minimal delays on writes
	opts.KeepAlive = 10               // Keepalive every 10 seconds
	opts.PingTimeout = time.Second    // local broker so response should be quick

	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)

	// Log events
	opts.DefaultPublishHandler = func(_ mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Unexpected message: %s\n", msg)
	}
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		fmt.Printf("Connection lost: %v.", err)
	}
	opts.OnConnect = func(client mqtt.Client) {
		log.Println("Connected.")
		t := client.Subscribe(TOPIC, QOS, h.handle)
		go func() {
			<-t.Done()
			if t.Error() != nil {
				fmt.Printf("Error subscribing: %s\n", t.Error())
			} else {
				fmt.Println("Subscribed to: ", TOPIC)
			}
		}()
	}

	opts.OnReconnecting = func(mqtt.Client, *mqtt.ClientOptions) {
		fmt.Println("Attempting to reconnect...")
	}
	// create the client using the options above
	client := mqtt.NewClient(opts)
	client.AddRoute(TOPIC, h.handle)
	// throw an error if the connection isn't successfull
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	fmt.Println("Connection is up.")
	// Use signals to exit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	<-sig
	fmt.Println("signal caught - exiting")
	client.Disconnect(1000)
	fmt.Println("Shutdown complete.")
}
