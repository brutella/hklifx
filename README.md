# hklifx

This project is an implementation of a HomeKit bridge for [LIFX](http://www.lifx.com) light bulbs using [HomeControl](https://github.com/brutella/hc) and [lifx](https://github.com/wolfeidau/lifx).

The official [LIFX app](http://www.lifx.com/pages/go) for iOS or Android is required to initially setup the light bulbs. After that you can use the `hklifxd` daemon to control your lights via HomeKit by using [Home][home] or any other HomeKit-compatible app.

# Installation

## Build

Build `hklifxd.go` using `go build hklifxd.go` or use the Makefile to build for Beaglebone Black
    
    make bbb
    
or Raspberry Pi

    make rpi

## Run

You need to provide the accessory pin as argument when running the `hklifxd` daemon.

    hklifxd -pin=32112321


## iOS

You need an iOS app to access HomeKit accessories. You can use [Home][home] which runs on iPhone, iPad and Apple Watch.

When you pair the accessory with HomeKit using an app, you have to enter the 8-digit pin (see *pin* argument).

[home]: http://selfcoded.com/home

# Contact

Matthias Hochgatterer

Github: [https://github.com/brutella](https://github.com/brutella)

Twitter: [https://twitter.com/brutella](https://twitter.com/brutella)

# License

hklifx is available under a non-commercial license. See the LICENSE file for more info.