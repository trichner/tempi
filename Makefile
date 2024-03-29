
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
