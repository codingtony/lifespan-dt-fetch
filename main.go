package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
)

var (
	name = flag.String("name", "LifeSpan", "name of remote peripheral")
	addr = flag.String("addr", "", "address of remote peripheral (MAC on Linux, UUID on OS X)")
	sd   = flag.Duration("sd", 30*time.Second, "scanning duration, 0 for indefinitely")
)

var characteristic055 = &ble.Characteristic{ValueHandle: 0x55}
var characteristic052 = &ble.Characteristic{ValueHandle: 0x52, CCCD: &ble.Descriptor{Handle: 0x53}}

var yellow = color.New(color.FgHiYellow).SprintfFunc()
var lightWhite = color.New(color.FgWhite).SprintfFunc()
var white = color.New(color.FgHiWhite).SprintfFunc()
var black = color.New(color.FgHiBlack).SprintfFunc()

// some part of this program is inspired from examples provided as part of go-ble/ble
func main() {
	flag.Parse()
	fmt.Println("Initializing Bluetooth")
	d, err := dev.NewDevice("default")
	if err != nil {
		log.Fatalf("can't new device : %s", err)
	}
	ble.SetDefaultDevice(d)

	// Default to search device with name of Gopher (or specified by user).
	filter := func(a ble.Advertisement) bool {
		return strings.ToUpper(a.LocalName()) == strings.ToUpper(*name)
	}

	// If addr is specified, search for addr instead.
	if len(*addr) != 0 {
		filter = func(a ble.Advertisement) bool {
			return strings.ToUpper(a.Addr().String()) == strings.ToUpper(*addr)
		}
	}

	// Scan for specified durantion, or until interrupted by user.
	fmt.Printf("Scanning for %s...\n", *sd)
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), *sd))
	cln, err := ble.Connect(ctx, filter)
	if err != nil {
		log.Fatalf("can't connect : %s", err)
	}

	// Make sure we had the chance to print out the message.
	done := make(chan struct{})

	go func() {
		<-cln.Disconnected()
		fmt.Printf("[ %s ] is disconnected \n", cln.Addr())
		close(done)
	}()

	notif := make(chan bool)
	durationChan := make(chan time.Duration, 1)
	caloriesChan := make(chan int, 1)
	distanceChan := make(chan float64, 1)
	stepChan := make(chan int, 1)

	err = cln.Subscribe(characteristic052, false, distanceHandle(notif, distanceChan))
	if err != nil {
		log.Fatalf("can't subscribe : %s", err)
	}
	write(cln, "a185000000", notif) // Distance

	err = cln.Subscribe(characteristic052, false, caloriesHandle(notif, caloriesChan))
	if err != nil {
		log.Fatalf("can't subscribe : %s", err)
	}
	write(cln, "a187000000", notif) // Calories

	err = cln.Subscribe(characteristic052, false, stepHandle(notif, stepChan))
	if err != nil {
		log.Fatalf("can't subscribe : %s", err)
	}
	write(cln, "a188000000", notif) // Steps

	err = cln.Subscribe(characteristic052, false, timeHandle(notif, durationChan))
	if err != nil {
		log.Fatalf("can't subscribe : %s", err)
	}
	write(cln, "a189000000", notif) // Time

	err = cln.Subscribe(characteristic052, false, speedHandle(notif))
	if err != nil {
		log.Fatalf("can't subscribe : %s", err)
	}
	write(cln, "a182000000", notif) // Speed

	distance := <-distanceChan
	calories := <-caloriesChan
	duration := <-durationChan
	steps := <-stepChan
	//fmt.Println(t.Format("2006-01-02 15:04:05"))
	fmt.Printf("%s distance : %.2f, calories: %d, duration : %s, steps : %d\n", time.Now().UTC().Format("2006-01-02T15:04:05Z"), distance, calories, duration, steps)

	// debug(cln, "0200000000", notif)
	// debug(cln, "0400000000", notif)
	// debug(cln, "a17F000000", notif)
	// debug(cln, "a180000000", notif)
	// debug(cln, "a181000000", notif)
	// debug(cln, "a183000000", notif)
	// debug(cln, "a184000000", notif)
	// debug(cln, "a18A000000", notif)

	/*
		0200000000:     [ 02 AA 11 18 00 00 ] // 4376 or 17.24
		0400000000:     [ 04 AA 00 00 00 00 ]
		a181000000:     [ A1 AA AA 00 00 00 ]
		a182000000:     [ A1 AA 01 3C 00 00 ] // 316 or 1.60 // speed
		a183000000:     [ A1 AA 00 00 00 00 ]

		On pause :
		0200000000:     [ 02 AA 11 18 00 00 ]
		0400000000:     [ 04 AA 00 00 00 00 ]
		a181000000:     [ A1 AA AA 00 00 00 ]
		a182000000:     [ A1 AA 00 00 00 00 ]
		a183000000:     [ A1 AA 00 00 00 00 ]


		a182000000 = Speed
		02 32 = 2.50km/h
		03 00 = 3.00km/h
		01 3C = 1.60km/h


	*/
	// write(cln, "e000000000") /// <-- this pause the threadmill
	// write(cln, "e200000000") /// <-- this reset the threadmill

	// Disconnect the connection. (On OS X, this might take a while.)
	fmt.Printf("Disconnecting [ %s ]... (this might take up to few seconds on OS X)\n", cln.Addr())
	cln.CancelConnection()
	<-done
}

// Great help for hex conversion : https://www.scadacore.com/tools/programming-calculators/online-hex-converter/

func caloriesHandle(notif chan (bool), caloriesChan chan (int)) func([]byte) {
	//A1 AA 00 AC 00 00 : 172
	return func(req []byte) {
		buffer := bytes.NewReader(req[2:])
		var val uint16
		binary.Read(buffer, binary.BigEndian, &val)

		fmt.Printf("%s %s %s\n", lightWhite("Calories:"), yellow("%d", val), black("[ % X ]", req))
		caloriesChan <- int(val)
		notif <- true
	}
}

func distanceHandle(notif chan (bool), distanceChan chan float64) func([]byte) {
	return func(req []byte) {
		buffer := bytes.NewReader(req[2:3])
		var km uint8
		binary.Read(buffer, binary.BigEndian, &km)

		buffer = bytes.NewReader(req[3:4])
		var decimal uint8
		binary.Read(buffer, binary.BigEndian, &decimal)

		fmt.Printf("%s %s %s\n", lightWhite("Distance:"), yellow("%d.%d", km, decimal), black("[ % X ]", req))
		distance := float64(km) + float64(decimal)/100.0
		distanceChan <- distance
		notif <- true
	}
}

func speedHandle(notif chan (bool)) func([]byte) {
	return func(req []byte) {
		buffer := bytes.NewReader(req[2:3])
		var km uint8
		binary.Read(buffer, binary.BigEndian, &km)

		buffer = bytes.NewReader(req[3:4])
		var decimal uint8
		binary.Read(buffer, binary.BigEndian, &decimal)

		fmt.Printf("%s %s %s\n", lightWhite("Speed:   "), yellow("%d.%d", km, decimal), black("[ % X ]", req))
		notif <- true
	}
}

func stepHandle(notif chan (bool), stepChan chan (int)) func([]byte) {
	//  A1 AA 01 B5 00 00  -> ~436 steps
	//  A1 AA 01 D9 00 00  -> ~475 steps
	//  A1 AA 01 FF 00 00  -> ~512 steps
	return func(req []byte) {

		buffer := bytes.NewReader(req[2:])
		var val uint16
		binary.Read(buffer, binary.BigEndian, &val)

		fmt.Printf("%s %s %s\n", lightWhite("Steps:   "), yellow("%d", val), black("[ % X ]", req))
		stepChan <- int(val)
		notif <- true
	}
}

func timeHandle(notif chan (bool), durationChan chan (time.Duration)) func([]byte) {
	// A1 AA 01 FF 00 00  (9:45, 633s 24 cal, 0.28k)
	// A1 AA 00 0A 00 00 (10:00, 654s, 25 cal, 0.30k)
	// A1 AA 00 0B 00 00 (11:00, 758s, 29 cal, 0.37k)
	// A1 AA 00 0B 1E 00 (11:30)
	// A1 AA 00 38 22 00  (56:34)
	// A1 AA 01 03 37 00  1:03:55

	return func(req []byte) {
		// this work under 60min

		buffer := bytes.NewReader(req[2:3])
		var hour uint8
		binary.Read(buffer, binary.BigEndian, &hour)

		buffer = bytes.NewReader(req[3:4])
		var min uint8
		binary.Read(buffer, binary.BigEndian, &min)

		// this seems good for seconds, even after 60min
		buffer = bytes.NewReader(req[4:])
		var sec uint16
		binary.Read(buffer, binary.LittleEndian, &sec)

		fmt.Printf("%s %s %s\n", lightWhite("Time:    "), yellow("%d:%02d:%02d", hour, min, sec), black("[ % X ]", req))
		t := time.Duration(hour)*time.Hour + time.Duration(min)*time.Minute + time.Duration(sec)*time.Second
		durationChan <- t
		notif <- true
	}
}

func debug(cln ble.Client, value055 string, notif chan (bool)) {
	err := cln.Subscribe(characteristic052, false, debugHandle(value055, notif))
	if err != nil {
		log.Fatalf("can't subscribe : %s", err)
	}
	write(cln, value055, notif)
}

func debugHandle(value055 string, notif chan (bool)) func([]byte) {
	return func(req []byte) {
		fmt.Printf("%s %s\n", lightWhite("%s:    ", value055), yellow("[ % X ]", req))
		notif <- true
	}
}

func write(cln ble.Client, hexstring string, notif chan (bool)) {
	data, _ := hex.DecodeString(hexstring)
	err := cln.WriteCharacteristic(characteristic055, data, true)
	if err != nil {
		log.Fatalf(err.Error())
	}
	<-notif
	cln.ClearSubscriptions()
}
