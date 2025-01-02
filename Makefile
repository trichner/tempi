
TARGET?=./feather-rp2040-homebrew.json
#TARGET=./feather-rp2040.json

.PHONY: build
build:
	tinygo build -target $(TARGET)

.PHONY: flash
flash:
	tinygo flash -print-stacks -size full -target $(TARGET) -monitor ./main/tlogger

.PHONY: fmt
fmt:
	gofumpt -l -w .

.PHONY: flash.tlogger
flash.tlogger:
	tinygo flash -size short -print-stacks -target=feather-rp2040 -monitor ./main/tlogger

.PHONY: flash.wifitest
flash.wifitest:
	tinygo flash -stack-size=4KB -size short -print-stacks -target=nano-rp2040 -monitor ./main/wifitest

.PHONY: debug.wifisens
flash.wifisens:
	tinygo flash -stack-size=4KB -size short -print-stacks -target=nano-rp2040 -monitor ./main/wifisens

.PHONY: flash.alerty
flash.alerty:
	tinygo flash -size short -print-stacks -target=pico -monitor ./main/alerty
	#tinygo flash -size short -print-stacks -target=pico -stack-size=8kb -monitor ./main/alerty

.PHONY: openocd.alerty
openocd.alerty:
	openocd-rp2040 -f interface/cmsis-dap.cfg -f target/rp2040.cfg -c "adapter speed 5000"
	#openocd-rp2040 -f interface/cmsis-dap.cfg -f target/rp2040.cfg -c "adapter speed 5000" -c 'program alerty.elf verify reset exit'

.PHONY: flash.littlefsck
flash.littlefsck:
	tinygo flash -print-stacks -size full -target $(TARGET) -monitor ./main/littlefsck

.PHONY: flash.rtc
flash.rtc:
	tinygo flash -print-stacks -size full -target $(TARGET) -monitor ./main/rtcsetup

.PHONY: flash.soilsensor
flash.soilsensor:
	tinygo flash -print-stacks -size full -target $(TARGET) -monitor ./main/soilsensor

.PHONY: flash.xmasds
flash.xmasds:
	tinygo flash -print-stacks -size full -target $(TARGET) -monitor ./main/xmasds
