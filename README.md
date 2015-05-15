# hklifx

This project is an implementation of a HomeKit bridge for [LIFX](http://www.lifx.com) light bulbs using [HomeControl](https://github.com/brutella/hc) and [lifx](https://github.com/wolfeidau/lifx).

The official [LIFX app](http://www.lifx.com/pages/go) for iOS or Android is required to initially setup the light bulbs. After that you can use the `hklifx` bridge to control your lights via HomeKit by using [Home](todo) or any other HomeKit-compatible app.

## Build

 Build `hklifxd.go` using `go build hklifxd.go` 

or 

- Use the Makefile to build for Beaglebone Black
    
        make bbb
    
    or Raspberry Pi

        make rpi