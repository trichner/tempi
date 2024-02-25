#!/bin/bash

set -e

docker run --rm -it -p 0.0.0.0:1883:1883 -p 9001:9001 -v ./mosquitto.conf:/mosquitto/config/mosquitto.conf eclipse-mosquitto
