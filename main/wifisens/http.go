package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strconv"
	"strings"
)

func postJsonViaTls(host, port, path string, body []byte) error {
	log("dialing")
	conn, err := dial(host + ":" + port)
	if err != nil {
		return fmt.Errorf("failed dialing: %w", err)
	}
	defer conn.Close()

	log("sending")
	if err := sendRequest(conn, host, path, body); err != nil {
		return fmt.Errorf("failed sending request: %w", err)
	}

	log("receiving")
	statusCode, err := receiveResponse(conn)
	if err != nil {
		return fmt.Errorf("failed receiving response: %w", err)
	}

	log("checking")
	if statusCode != 200 {
		return errors.New("bad status code: " + strconv.Itoa(statusCode))
	}

	return nil
}

func dial(addr string) (*net.TLSConn, error) {

	// root certs are part of the firmware
	conn, err := tls.Dial("tcp", addr, nil)
	//TODO: deal with possible errors
	if err != nil {
		return nil, err
	}
	return conn, err
}

func sendRequest(writer io.Writer, host, path string, body []byte) error {
	var buf bytes.Buffer
	// there are no write errors in bytes.Buffer, hence ignoring the returned one
	fmt.Fprintln(&buf, "POST", path, "HTTP/1.1")
	fmt.Fprintln(&buf, "Host:", host)
	fmt.Fprintln(&buf, "User-Agent: TinyGo")
	fmt.Fprintln(&buf, "Connection: close")
	fmt.Fprintln(&buf, "Content-Type: application/json")
	fmt.Fprintf(&buf, "Content-Length: %d\n", len(body))
	buf.WriteByte('\n')

	_, err := buf.WriteTo(writer)
	if err != nil {
		return fmt.Errorf("failed to send headers: %w", err)
	}

	if len(body) > 0 {
		_, err := writer.Write(body)
		if err != nil {
			return fmt.Errorf("failed to send body: %w", err)
		}
	}
	return nil
}

func receiveResponse(reader io.Reader) (int, error) {

	// this is rather 'fat', though doing all that buffering etc. ourselves is a real pain
	responseReader := textproto.NewReader(bufio.NewReader(reader))

	statusLine, err := responseReader.ReadLine()
	if err != nil {
		return 0, err
	}

	proto, status, ok := strings.Cut(statusLine, " ")
	if !ok {
		return 0, errors.New("malformed HTTP response")
	}

	// don't care about it now
	_ = proto

	statusCode, _, _ := strings.Cut(status, " ")
	if len(statusCode) != 3 {
		return 0, fmt.Errorf("malformed HTTP status code: %s", statusCode)
	}

	code, err := strconv.Atoi(statusCode)
	if err != nil {
		return 0, fmt.Errorf("malformed HTTP status code: %s", statusCode)
	}

	// Yolo, the rest should be fine. Who needs response bodies anyways ¯\_(ツ)_/¯

	return code, nil
}
