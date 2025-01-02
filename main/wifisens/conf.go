package main

import _ "embed"

//go:embed wifi_ssid.txt
var ssid string

//go:embed wifi_psk.txt
var pass string
