#!/bin/bash

tinygo flash -target feather-rp2040 && sleep 2 && screen -L /dev/ttyACM0 9600
