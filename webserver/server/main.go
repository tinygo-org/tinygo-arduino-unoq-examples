package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.bug.st/serial"
)

const (
	listenAddr = ":8080"
	serialDev  = "/dev/ttyHS1"
	baudRate   = 115200
)

var (
	port     serial.Port
	serialMu sync.Mutex

	latestValue string
	latestTime  time.Time
	latestMu    sync.RWMutex
)

func openPort() error {
	var err error
	port, err = serial.Open(serialDev, &serial.Mode{
		BaudRate: baudRate,
	})
	if err != nil {
		return err
	}
	port.SetReadTimeout(100 * time.Millisecond)
	return nil
}

func drainPort() {
	var buf [64]byte
	for {
		n, _ := port.Read(buf[:])
		if n == 0 {
			break
		}
	}
}

func readSensor() (string, error) {
	serialMu.Lock()
	defer serialMu.Unlock()

	// Drain any stale data before sending command.
	drainPort()

	_, err := port.Write([]byte("READ\r\n"))
	if err != nil {
		return "", fmt.Errorf("serial write error: %w", err)
	}

	// Small delay to let sensor process command and start responding.
	time.Sleep(50 * time.Millisecond)

	var buf [64]byte
	var line []byte
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		n, err := port.Read(buf[:])
		if err != nil {
			return "", fmt.Errorf("serial read error: %w", err)
		}
		if n > 0 {
			line = append(line, buf[:n]...)
			if idx := strings.IndexByte(string(line), '\n'); idx >= 0 {
				return strings.TrimSpace(string(line[:idx])), nil
			}
		}
	}
	return "", fmt.Errorf("serial read timeout (got %d bytes so far: %q)", len(line), line)
}

func pollSensor() {
	for {
		value, err := readSensor()
		if err != nil {
			log.Printf("sensor poll error: %v", err)
		} else {
			latestMu.Lock()
			latestValue = value
			latestTime = time.Now().UTC()
			latestMu.Unlock()
		}
		time.Sleep(1 * time.Second)
	}
}

func handleSensor(w http.ResponseWriter, r *http.Request) {
	latestMu.RLock()
	value := latestValue
	t := latestTime
	latestMu.RUnlock()

	if value == "" {
		http.Error(w, "no sensor data yet", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"analog_a0":%s,"time":"%s"}`, value, t.Format(time.RFC3339))
}

func main() {
	if err := openPort(); err != nil {
		log.Fatalf("failed to open serial port %s: %v", serialDev, err)
	}
	defer port.Close()

	go pollSensor()

	http.HandleFunc("/sensor", handleSensor)
	http.Handle("/", http.FileServer(http.Dir("/home/arduino/www")))

	log.Printf("listening on %s", listenAddr)
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
				continue
			}
			addrs, _ := iface.Addrs()
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
					log.Printf("  http://%s%s/", ipnet.IP, listenAddr)
				}
			}
		}
	}
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
