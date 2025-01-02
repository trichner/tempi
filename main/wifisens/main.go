//go:build rp2040

package main

import (
	"fmt"
	"machine"
	"strconv"
	"time"
	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"

	"github.com/trichner/tempi/pkg/sht4x"
	"github.com/trichner/tempi/pkg/toggler"
)

const watchDogMillis = 20_000

func main() {
	machine.InitSerial()

	log("setting up watchdog")
	wd := machine.Watchdog
	wd.Configure(machine.WatchdogConfig{watchDogMillis})

	time.Sleep(2 * time.Second)
	log("ready to go")

	log("setup i2c")
	bus := machine.I2C0
	err := bus.Configure(machine.I2CConfig{})
	if err != nil {
		panic(err)
	}

	log("setup temp")
	sht := sht4x.New(bus, 0)

	log("ready for blink")
	led := toggler.SetupToggler(machine.LED)

	log("setting up netstack")
	link, _ := probe.Probe()

	log("connecting to WiFi AP")
	err = link.NetConnect(&netlink.ConnectParams{
		Ssid:       ssid,
		Passphrase: pass,
	})
	if err != nil {
		panic(err)
	}
	defer link.NetDisconnect()

	hwaddr, err := link.GetHardwareAddr()
	if err != nil {
		panic(err)
	}
	log("hwaddr: " + hwaddr.String())
	deviceId := hwaddr.String()

	log("starting watchdog")
	err = wd.Start()
	if err != nil {
		panic(err)
	}
	log("starting loop")

	errorStreak := 0

	sleepTime := time.Second * 5
	sampleTime := time.Minute

	nextMeasurement := time.Now()
	for {
		wd.Update()
		led.Toggle()

		now := time.Now()
		if now.Before(nextMeasurement) {
			time.Sleep(sleepTime)
			continue
		}
		nextMeasurement = now.Add(sampleTime)

		log("> appending record")
		temp, hum, err := sht.ReadTemperatureHumidity()
		if err != nil {
			panic(err)
		}

		if err := postMeasurement(deviceId, temp, hum); err != nil {
			errorStreak++
			log("ERROR posting a measurement, skipping: " + err.Error() + " this is the " + strconv.Itoa(errorStreak) + " try")
			if errorStreak > 16 {
				panic(err)
			}
		}
		log("< appended record")
		time.Sleep(sleepTime)
	}
}

func postMeasurement(deviceId string, temperatureMilliCelsius int32, relativeHumidityMilliPercent int32) error {

	data := []byte(fmt.Sprintf(`{"temperature_milli_celsius":%d,"relative_humidity_milli_percent":%d}`, temperatureMilliCelsius, relativeHumidityMilliPercent))

	log("posting record")
	host := "events-236347963523.europe-north1.run.app"
	port := "443"
	path := "/events/" + deviceId

	return postJsonViaTls(host, port, path, data)
}

func log(s string) {
	_, err := machine.Serial.Write([]byte(s + "\n\r"))
	if err != nil {
		panic(err)
	}
}
