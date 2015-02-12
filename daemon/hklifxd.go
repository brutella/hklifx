package main

import (
	"fmt"
	"log"

	"github.com/brutella/hap/app"
	"github.com/brutella/hap/model"
	"github.com/brutella/hap/model/accessory"
	"github.com/brutella/hap/server"
	"github.com/wolfeidau/lifx"
	"math"
	"time"
)

const (
	// from https://github.com/LIFX/LIFXKit/blob/master/LIFXKit/Classes-Common/LFXHSBKColor.h
	HSBKKelvinDefault = uint16(3500)
	HSBKKelvinMin     = uint16(2500)
	HSBKKelvinMax     = uint16(9000)
)

func ConnectLIFX() {
	client = lifx.NewClient()
	err := client.StartDiscovery()

	if err != nil {
		log.Fatalf("Could not find bulb %s", err)
	}

	go func() {
		sub := client.Subscribe()
		for {
			event := <-sub.Events
			switch event := event.(type) {
			case *lifx.Gateway:
			case *lifx.Bulb:
				updateBulb(event)
			default:
				log.Printf("Event %v", event)
			}
		}
	}()
}

func updateBulb(bulb *lifx.Bulb) {
	on := true
	if bulb.GetPower() == 0 {
		on = false
	}

	state := bulb.GetState()

	light_bulb := lightForBulb(bulb).bulb

	light_bulb.SetOn(on)

	brightness := float64(state.Brightness) / float64(math.MaxUint16) * 100
	saturation := float64(state.Saturation) / float64(math.MaxUint16) * 100
	hue := float64(state.Hue) / float64(math.MaxUint16) * 360

	light_bulb.SetBrightness(int(brightness))
	light_bulb.SetSaturation(saturation)
	light_bulb.SetHue(hue)

	log.Println("LIFX is now", on)
	log.Println("Brightness", brightness)
	log.Println("Saturation", saturation)
	log.Println("Hue", hue)
}

func toggleBulb(bulb *lifx.Bulb) {
	if bulb.GetPower() == 0 {
		client.LightOn(bulb)
	} else {
		client.LightOff(bulb)
	}
}

func lightForBulb(bulb *lifx.Bulb) *lifxLight {
	label := bulb.GetLabel()
	light, found := lights[label]
	if found == true {
		return light
	}

	fmt.Println("Create new switch for blub")
	application.SetReachable(true)

	info := model.Info{
		Name:         label,
		SerialNumber: "001",
		Manufacturer: "LIFX",
		Model:        "LIFX",
	}

	light_bulb := accessory.NewLightBulb(info)
	light_bulb.OnIdentify(func() {
		timeout := 1 * time.Second
		toggleBulb(bulb)
		time.Sleep(timeout)
		toggleBulb(bulb)
		time.Sleep(timeout)
		toggleBulb(bulb)
		time.Sleep(timeout)
		toggleBulb(bulb)
	})

	light_bulb.OnStateChanged(func(on bool) {
		if on == true {
			client.LightOn(bulb)
			log.Println("Switch is on")
		} else {
			client.LightOff(bulb)
			log.Println("Switch is off")
		}
	})

	updateColors := func(client *lifx.Client, bulb *lifx.Bulb) {
		// TODO define max variables in Gohap

		// HAP: [0...360]
		// LIFX: [0...MAX_UINT16]
		hue := light_bulb.GetHue()

		// HAP: [0...100]
		// LIFX: [0...MAX_UINT16]
		saturation := light_bulb.GetSaturation()
		// HAP: [0...100]
		// LIFX: [0...MAX_UINT16]
		brightness := light_bulb.GetBrightness()
		// [2500..9000]
		kelvin := HSBKKelvinDefault

		lifx_brightness := math.MaxUint16 * float64(brightness) / 100
		lifx_saturation := math.MaxUint16 * float64(saturation) / 100
		lifx_hue := math.MaxUint16 * float64(hue) / 360

		log.Println("Brightness", lifx_brightness)
		log.Println("Hue", lifx_saturation)
		log.Println("Saturation", lifx_hue)
		client.LightColour(bulb, uint16(lifx_hue), uint16(lifx_saturation), uint16(lifx_brightness), uint16(kelvin), 0x0500)
	}

	light_bulb.OnBrightnessChanged(func(value int) {
		updateColors(client, bulb)
	})

	light_bulb.OnSaturationChanged(func(value float64) {
		updateColors(client, bulb)
	})

	light_bulb.OnHueChanged(func(value float64) {
		updateColors(client, bulb)
	})

	application.AddAccessory(light_bulb.Accessory)
	light = &lifxLight{light_bulb, light_bulb.Accessory}
	lights[label] = light

	return light
}

type lifxLight struct {
	bulb      model.LightBulb
	accessory *accessory.Accessory
}

var application *app.App
var lights map[string]*lifxLight
var client *lifx.Client

func main() {
	lights = map[string]*lifxLight{}

	conf := app.NewConfig()
	conf.DatabaseDir = "./data"
	conf.BridgeName = "LIFXBridge"

	pwd, _ := server.NewPassword("11122333")
	conf.BridgePassword = pwd
	conf.BridgeManufacturer = "Matthias Hochgatterer"

	var err error
	application, err = app.NewApp(conf)
	if err != nil {
		log.Fatal(err)
	}

	ConnectLIFX()

	application.RunAndPublish(false)
}