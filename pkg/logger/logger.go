package logger

import (
	"fmt"
	"io"
	"machine"
	"os"
	"strconv"
	"time"

	"tinygo.org/x/drivers/sdcard"
	"tinygo.org/x/tinyfs/littlefs"
)

const (
	bootCountFileName = "boot_count"
	logFileName       = "log_file.jsonlines"
)

type Record struct {
	Timestamp                    time.Time
	MilliDegreeCelsius           int32
	MilliPercentRelativeHumidity int32
	SoilHumidity                 int32
}

type Logger struct {
	card *sdcard.Device
	fs   *littlefs.LFS
}

func New() (*Logger, error) {
	spi := machine.SPI0

	// https://learn.adafruit.com/adafruit-adalogger-featherwing/pinouts
	sd := sdcard.New(spi, machine.SPI0_SCK_PIN, machine.SPI0_SDO_PIN, machine.SPI0_SDI_PIN, machine.GPIO10)
	err := sd.Configure()
	if err != nil {
		return nil, err
	}

	fs := littlefs.New(&sd)

	fs.Configure(&littlefs.Config{
		CacheSize:     512,
		LookaheadSize: 512,
		BlockCycles:   100,
	})

	err = fs.Mount()
	if err != nil {
		print("re-formatting SD card: " + err.Error())
		if err = fs.Format(); err != nil {
			return nil, err
		}
		if err = fs.Mount(); err != nil {
			return nil, err
		}
	}

	return &Logger{
		card: &sd,
		fs:   fs,
	}, nil
}

func (l *Logger) IncrementBootCount() (int, error) {
	count, err := readBootCount(l.fs)
	if err != nil {
		return 0, err
	}
	count++
	return count, writeBootCount(l.fs, count)
}

func (l *Logger) AppendRecord(r *Record) error {
	line := fmt.Sprintf("{\"ts\":%d,\"temperature\":%d,\"humidity\":%d,\"soilhumidity\":%d}\n", r.Timestamp.Unix(), r.MilliDegreeCelsius, r.MilliPercentRelativeHumidity, r.SoilHumidity)

	f, err := l.fs.OpenFile(logFileName, os.O_RDWR|os.O_APPEND|os.O_CREATE)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.WriteString(f, line)
	return err
}

func writeBootCount(fs *littlefs.LFS, count int) error {
	f, err := fs.OpenFile(bootCountFileName, os.O_RDWR|os.O_CREATE)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.WriteString(f, strconv.Itoa(count))
	return err
}

func readBootCount(fs *littlefs.LFS) (int, error) {
	f, err := fs.OpenFile(bootCountFileName, os.O_RDONLY|os.O_CREATE)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	countRaw, err := io.ReadAll(f)
	if len(countRaw) == 0 {
		return 0, nil
	}
	return strconv.Atoi(string(countRaw))
}

func ls(fs *littlefs.LFS, path string) {
	dir, err := fs.Open(path)
	if err != nil {
		fmt.Printf("Could not open directory %s: %v\n", path, err)
		return
	}
	defer dir.Close()
	infos, err := dir.Readdir(0)
	if err != nil {
		fmt.Printf("Could not read directory %s: %v\n", path, err)
		return
	}
	for _, info := range infos {
		s := "f "
		if info.IsDir() {
			s = "d "
		}
		fmt.Printf("%s %5d %s\n", s, info.Size(), info.Name())
	}
}

func lsblk(dev *sdcard.Device) {
	csd := dev.CSD
	sectors, err := csd.Sectors()
	if err != nil {
		fmt.Printf("%s\r\n", err.Error())
		return
	}
	cid := dev.CID

	fmt.Printf(
		"\r\n-------------------------------------\r\n"+
			" Device Information:  \r\n"+
			"-------------------------------------\r\n"+
			" JEDEC ID: %v\r\n"+
			" Serial:   %v\r\n"+
			" Status 1: %02x\r\n"+
			" Status 2: %02x\r\n"+
			" \r\n"+
			" Max clock speed (MHz): %d\r\n"+
			" Has Sector Protection: %t\r\n"+
			" Supports Fast Reads:   %t\r\n"+
			" Supports QSPI Reads:   %t\r\n"+
			" Supports QSPI Write:   %t\r\n"+
			" Write Status Split:    %t\r\n"+
			" Single Status Byte:    %t\r\n"+
			"-Sectors:               %d\r\n"+
			"-Bytes (Sectors * 512)  %d\r\n"+
			"-ManufacturerID         %02X\r\n"+
			"-OEMApplicationID       %04X\r\n"+
			"-ProductName            %s\r\n"+
			"-ProductVersion         %s\r\n"+
			"-ProductSerialNumber    %08X\r\n"+
			"-ManufacturingYear      %02X\r\n"+
			"-ManufacturingMonth     %02X\r\n"+
			"-Always1                %d\r\n"+
			"-CRC                    %02X\r\n"+
			"-------------------------------------\r\n\r\n",
		"attrs.JedecID",         // attrs.JedecID,
		cid.ProductSerialNumber, // serialNumber1,
		0,                       // status1,
		0,                       // status2,
		csd.TRAN_SPEED,          // attrs.MaxClockSpeedMHz,
		false,                   // attrs.HasSectorProtection,
		false,                   // attrs.SupportsFastRead,
		false,                   // attrs.SupportsQSPI,
		false,                   // attrs.SupportsQSPIWrites,
		false,                   // attrs.WriteStatusSplit,
		false,                   // attrs.SingleStatusByte,
		sectors,
		csd.Size(),
		cid.ManufacturerID,
		cid.OEMApplicationID,
		cid.ProductName,
		cid.ProductVersion,
		cid.ProductSerialNumber,
		cid.ManufacturingYear,
		cid.ManufacturingMonth,
		cid.Always1,
		cid.CRC,
	)
}
