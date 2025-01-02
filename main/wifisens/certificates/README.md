# Update WiFiNINA Certificates and Firmware

* WARN: Somehow this doesn't work for self-signed CA certs currently


## Steps

1. Flash the updater sketch to the microcontroller
   1. either via the official [How To](https://support.arduino.cc/hc/en-us/articles/360013896579-Use-the-Firmware-Updater-in-Arduino-IDE)
   2. or by directly flashing `fwupdater_nano_rp2040_connect.bin`
2. Get the `arduino-fwuploader`
   ```shell
   curl -s -L https://github.com/arduino/arduino-fwuploader/releases/download/2.4.1/arduino-fwuploader_2.4.1_Linux_64bit.tar.gz | tar xvz
   ```
3. Fetch the default root certificates
   ```shell
   curl -L -O https://raw.githubusercontent.com/arduino/nina-fw/refs/heads/master/data/roots.pem
   ```
4. Append any additional trust roots to that `roots.pem` file in PEM format
5. Copy the certificates to the WiFiNINA module
   IMPORTANT: Update the `address` accordingly! 
   ```shell
   ./arduino-fwuploader certificates  flash --fqbn arduino:mbed_nano:nanorp2040connect --address /dev/ttyACM1 --file all_certs.pem
   ```


# Kitchen Sink

**Generate a fresh CA cert as well as a server cert signed by it**
```shell
./gen_certs.sh
```

**Running a TLS server with a generated certificate**
```shell
openssl s_server -4 -accept 0.0.0.0:8443 -cert server.crt -key server.key -debug
```
