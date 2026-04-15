package main

import (
	"machine"
	"strconv"
	"sync/atomic"
	"time"
)

var (
	uart        = machine.Serial
	latestValue atomic.Uint32
)

func main() {
	time.Sleep(2 * time.Second)

	uart.Configure(machine.UARTConfig{})

	adc := machine.ADC{Pin: machine.A0}
	machine.InitADC()
	adc.Configure(machine.ADCConfig{})

	// Continuously sample analog reading in background.
	go func() {
		for {
			latestValue.Store(uint32(adc.Get()))
			time.Sleep(100 * time.Millisecond)
		}
	}()

	input := make([]byte, 64)
	i := 0
	for {
		if uart.Buffered() > 0 {
			data, _ := uart.ReadByte()
			if data == '\n' || data == '\r' {
				cmd := string(input[:i])
				i = 0
				if cmd == "READ" {
					value := latestValue.Load()
					uart.Write([]byte(strconv.FormatUint(uint64(value), 10)))
					uart.Write([]byte("\r\n"))
				} else if len(cmd) > 0 {
					uart.Write([]byte("ERR: unknown command\r\n"))
				}
			} else {
				if i < len(input) {
					input[i] = data
					i++
				}
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}
