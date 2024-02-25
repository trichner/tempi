package main

import (
	"context"
	"encoding/hex"
	mqtt "github.com/soypat/natiu-mqtt"
	"github.com/soypat/seqs/stacks"
	"github.com/trichner/tempi/pkg/netstack"
	"io"
	"log/slog"
	"machine"
	"net/netip"
	"time"
)

func main() {
	// one-off, generate server cert & keys:
	// openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 365 -nodes

	// start TLS server:
	//  openssl s_server -key key.pem -cert cert.pem -accept 4433 -state -msg

	logger := slog.New(slog.NewTextHandler(machine.Serial, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	_, stack, _, err := netstack.SetupWithDHCP(netstack.SetupConfig{
		Hostname: "alerty",
		Logger:   logger,
		TCPPorts: 1,

		// configure in secrets.go
		SSID: ssid,
		PSK:  pass,
	})
	if err != nil {
		logger.Error("netstack setup failed: %v", err)
		panic(err)
	}

	const socketBuf = 256

	socket, err := stacks.NewTCPConn(stack, stacks.TCPConnConfig{TxBufSize: socketBuf, RxBufSize: socketBuf})
	if err != nil {
		panic("socket create:" + err.Error())
	}

	dstip := netip.AddrFrom4([...]byte{192, 168, 16, 94})
	//dstport := uint16(1883)
	dstport := uint16(4433)

	hwdst, err := netstack.ResolveHardwareAddr(stack, dstip)
	if err != nil {
		panic("resolve ip '" + dstip.String() + "':" + err.Error())
	}

	logger.Info("dialing tcp", slog.String("dst_ip", dstip.String()), slog.Int("dst_port", int(dstport)), slog.String("dst_hw", hex.EncodeToString(hwdst[:])))

	randomLocalPort := uint16(1337)
	err = socket.OpenDialTCP(randomLocalPort, hwdst, netip.AddrPortFrom(dstip, dstport), 200)
	if err != nil {
		panic("tcp dial fail '" + dstip.String() + "':" + err.Error())
	}

	// works with `nc -l 0.0.0.0 4433`
	//for {
	//	socket.Write([]byte("Hello World!\n"))
	//	time.Sleep(3 * time.Second)
	//}

	logger.Info("create mqtt client")

	// Create new client.
	client := mqtt.NewClient(mqtt.ClientConfig{
		Decoder: mqtt.DecoderNoAlloc{make([]byte, 4*1024)},
		OnPub: func(_ mqtt.Header, _ mqtt.VariablesPublish, r io.Reader) error {
			message, _ := io.ReadAll(r)
			logger.Info("mqtt rx", slog.String("msg", string(message)))
			return nil
		},
	})

	logger.Info("connect mqtt client")
	// Prepare for CONNECT interaction with server.
	var varConn mqtt.VariablesConnect

	varConn.SetDefaultMQTT([]byte("pico"))
	ctx := context.Background()
	err = client.Connect(ctx, socket, &varConn)

	if err != nil {
		// Error or loop until connect success.
		logger.Error("connect attempt failed", slog.Any("error", err))
		panic(err)
	}

	// Ping forever until error.
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		pingErr := client.Ping(ctx)
		cancel()
		if pingErr != nil {
			logger.Error("ping error", slog.Any("error", pingErr), slog.Any("reason", client.Err()))
			panic(pingErr)
		}
		logger.Info("ping success")
	}
}
