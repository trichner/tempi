package main

import (
	"crypto/tls"
	"log/slog"
	"machine"
	"net/netip"
	"time"

	"github.com/soypat/seqs/stacks"
	"github.com/trichner/tempi/pkg/netstack"
)

func main() {
	// one-off, generate server cert & keys:
	// openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 365 -nodes

	// start TLS server:
	//  openssl s_server -key key.pem -cert cert.pem -accept 4433 -state -msg

	logger := slog.New(slog.NewTextHandler(machine.Serial, &slog.HandlerOptions{
		Level: slog.LevelInfo,
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
	dstport := uint16(4433)
	hwdst, err := netstack.ResolveHardwareAddr(stack, dstip)
	if err != nil {
		panic("resolve ip '" + dstip.String() + "':" + err.Error())
	}

	logger.Info("dialing tcp", slog.String("remote_ip", dstip.String()))

	randomLocalPort := uint16(1337)
	err = socket.OpenDialTCP(randomLocalPort, hwdst, netip.AddrPortFrom(dstip, dstport), 200)
	if err != nil {
		panic("tcp dial fail '" + dstip.String() + "':" + err.Error())
	}

	logger.Info("wrapping in tls", slog.String("remote_ip", dstip.String()), slog.Int("dst_port", int(dstport)))
	tlsConn := tls.Client(socket, &tls.Config{
		InsecureSkipVerify: true,
	})

	logger.Info("tls handshake", slog.String("remote_ip", dstip.String()), slog.Int("dst_port", int(dstport)))
	err = tlsConn.Handshake()
	if err != nil {
		panic("tls dial fail '" + dstip.String() + "':" + err.Error())
	}

	for {
		tlsConn.Write([]byte("Hello World!"))
		time.Sleep(time.Second)
	}
}
