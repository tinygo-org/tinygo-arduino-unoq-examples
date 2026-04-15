package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	mqtt "github.com/soypat/natiu-mqtt"
	"go.bug.st/serial"
)

const (
	broker    = "broker.hivemq.com:1883"
	topic     = "tinygo/arduino-unoq/sensor"
	serialDev = "/dev/ttyHS1"
	baudRate  = 115200
)

func main() {
	// Open serial connection to the microcontroller.
	port, err := serial.Open(serialDev, &serial.Mode{
		BaudRate: baudRate,
	})
	if err != nil {
		log.Fatalf("failed to open serial port %s: %v", serialDev, err)
	}
	defer port.Close()
	port.SetReadTimeout(2 * time.Second)

	// Create MQTT client.
	client := mqtt.NewClient(mqtt.ClientConfig{
		Decoder: mqtt.DecoderNoAlloc{UserBuffer: make([]byte, 1500)},
		OnPub: func(_ mqtt.Header, _ mqtt.VariablesPublish, r io.Reader) error {
			message, _ := io.ReadAll(r)
			log.Println("received:", string(message))
			return nil
		},
	})

	// Connect to MQTT broker.
	conn, err := net.Dial("tcp", broker)
	if err != nil {
		log.Fatalf("failed to connect to broker %s: %v", broker, err)
	}
	defer conn.Close()

	hostname, _ := os.Hostname()
	clientID := fmt.Sprintf("arduino-unoq-%s-%d", hostname, time.Now().UnixMilli()%10000)

	var varConn mqtt.VariablesConnect
	varConn.SetDefaultMQTT([]byte(clientID))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx, conn, &varConn)
	cancel()
	if err != nil {
		log.Fatalf("MQTT connect failed: %v", err)
	}
	log.Println("connected to MQTT broker at", broker)

	pubFlags, _ := mqtt.NewPublishFlags(mqtt.QoS0, false, false)

	// Main loop: read sensor and publish.
	reader := bufio.NewReader(port)
	var packetID uint16
	for {
		// Request a reading from the sensor.
		_, err := port.Write([]byte("READ\n"))
		if err != nil {
			log.Printf("serial write error: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Read the response line.
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("serial read error: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		value := strings.TrimSpace(line)
		if value == "" {
			continue
		}

		// Publish to MQTT.
		payload := []byte(fmt.Sprintf(`{"analog_a0":%s,"time":"%s"}`, value, time.Now().UTC().Format(time.RFC3339)))
		packetID++
		varPub := mqtt.VariablesPublish{TopicName: []byte(topic), PacketIdentifier: packetID}
		err = client.PublishPayload(pubFlags, varPub, payload)
		if err != nil {
			log.Printf("MQTT publish error: %v", err)
			// Attempt reconnect on next iteration.
			break
		}
		log.Printf("published: %s", payload)

		time.Sleep(5 * time.Second)
	}
}
