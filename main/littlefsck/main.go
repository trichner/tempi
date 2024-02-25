package main

import (
	"fmt"
	"io"
	"machine"
	"os"
	"time"

	"tinygo.org/x/drivers/sdcard"
	"tinygo.org/x/tinyfs/littlefs"
)

const testFile = "test.txt"

func main() {
	machine.InitSerial()
	time.Sleep(2 * time.Second)

	err := run()
	if err != nil {
		panic(err)
	}
}

func run() error {
	spi := machine.SPI0

	fmt.Println("configuring SD card")
	// https://learn.adafruit.com/adafruit-adalogger-featherwing/pinouts
	sd := sdcard.New(spi, machine.SPI0_SCK_PIN, machine.SPI0_SDO_PIN, machine.SPI0_SDI_PIN, machine.GPIO10)
	err := sd.Configure()
	if err != nil {
		return err
	}

	stat(&sd)

	fs := littlefs.New(&sd)

	fs.Configure(&littlefs.Config{
		CacheSize:     256,
		LookaheadSize: 256,
		BlockCycles:   100,
	})

	fmt.Println("mounting FS")
	err = fs.Mount()
	if err != nil {
		fmt.Printf("failed to mount, re-formatting: %s\n", err)
		if err = fs.Format(); err != nil {
			return err
		}
		if err = fs.Mount(); err != nil {
			return err
		}
	}

	fmt.Println("append line")
	err = appendToTestFile(fs, "test?")
	if err != nil {
		return err
	}

	fmt.Println("unmount")
	return fs.Unmount()
}

func stat(sd *sdcard.Device) {
	fmt.Printf("write block size: %d\n", sd.WriteBlockSize())
	fmt.Printf("erase block size: %d\n", sd.EraseBlockSize())
	fmt.Printf("size: %d\n", sd.Size())
	fmt.Printf("block count: %d\n", sd.Size()/sd.EraseBlockSize())
}

func appendToTestFile(fs *littlefs.LFS, s string) error {
	f, err := fs.OpenFile(testFile, os.O_RDWR|os.O_CREATE)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.WriteString(f, s+"\n")
	return err
}
