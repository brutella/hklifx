package main

import (
	"flag"
	"math"
	"os"
	"time"

	"github.com/brutella/log"

	"github.com/brutella/hc/hap"
	"github.com/brutella/hc/model"
	"github.com/brutella/hc/model/accessory"
	"github.com/brutella/hc/model/characteristic"

	"github.com/pdf/golifx"
	"github.com/pdf/golifx/common"
	"github.com/pdf/golifx/protocol"
)

const (
	// from https://github.com/LIFX/LIFXKit/blob/master/LIFXKit/Classes-Common/LFXHSBKColor.h
	HSBKKelvinDefault = uint16(3500)
	HSBKKelvinMin     = uint16(2500)
	HSBKKelvinMax     = uint16(9000)
)

type HKLight struct {
	accessory *accessory.Accessory
	sub       *common.Subscription
	transport hap.Transport

	light model.LightBulb
}

var (
	lights map[uint64]*HKLight
	pin    string
)

func Connect() {
	client, err := golifx.NewClient(&protocol.V2{Reliable: true})
	if err != nil {
		log.Fatalf("[ERR] Failed to initiliaze the client: %s", err)
	}

	client.SetDiscoveryInterval(30 * time.Second)

	sub, _ := client.NewSubscription()
	for {
		event := <-sub.Events()
		switch event.(type) {
		case common.EventNewLocation:
			log.Printf("[INFO] Discovered Location %s", event.(common.EventNewLocation).Location.GetLabel())
		case common.EventNewGroup:
			log.Printf("[INFO] Discovered Group %s", event.(common.EventNewGroup).Group.GetLabel())
		case common.EventNewDevice:
			label, _ := event.(common.EventNewDevice).Device.GetLabel()
			log.Printf("[INFO] Discovered Device %s", label)

			go NewDevice(event.(common.EventNewDevice).Device)

		case common.EventExpiredLocation:
			log.Printf("[INFO] Expired Location %s", event.(common.EventExpiredLocation).Location.GetLabel())
		case common.EventExpiredGroup:
			log.Printf("[INFO] Expired Group %s", event.(common.EventExpiredGroup).Group.GetLabel())
		case common.EventExpiredDevice:
			label, _ := event.(common.EventExpiredDevice).Device.GetLabel()
			log.Printf("[INFO] Expired Device %s", label)

			ExpireDevice(event.(common.EventExpiredDevice).Device)

		default:
			log.Printf("[INFO] Unknown Client Event: %T", event)
		}
	}
}

func NewDevice(device common.Device) {
	if light, ok := device.(common.Light); ok {
		hkLight := GetHKLight(light)

		hkLight.sub, _ = light.NewSubscription()
		for {
			event := <-hkLight.sub.Events()
			switch event.(type) {
			case common.EventUpdateLabel:
				log.Printf("[INFO] Updated Label for %s to %s", hkLight.accessory.Name(), event.(common.EventUpdateLabel).Label)
				// TODO Add support for label changes to HomeControl
				log.Printf("[INFO] Unsupported by HomeControl")
			case common.EventUpdatePower:
				log.Printf("[INFO] Updated Power for %s", hkLight.accessory.Name())
				hkLight.light.SetOn(event.(common.EventUpdatePower).Power)
			case common.EventUpdateColor:
				log.Printf("[INFO] Updated Color for %s", hkLight.accessory.Name())

				hue, saturation, brightness := ConvertLIFXColor(event.(common.EventUpdateColor).Color)

				hkLight.light.SetHue(hue)
				hkLight.light.SetSaturation(saturation)
				hkLight.light.SetBrightness(int(brightness))

			default:
				log.Printf("[INFO] Unknown Device Event: %T", event)
			}
		}
	} else {
		log.Println("[INFO] Unsupported Device")
	}
}

func ExpireDevice(device common.Device) {
	if light, ok := device.(common.Light); ok {
		hkLight, _ := lights[light.ID()]
		light.CloseSubscription(hkLight.sub)
		hkLight.transport.Stop()

		delete(lights, light.ID())
	} else {
		log.Println("[INFO] Unsupported Device")
	}
}

func GetHKLight(light common.Light) *HKLight {
	hkLight, found := lights[light.ID()]
	if found {
		return hkLight
	}

	label, _ := light.GetLabel()
	log.Printf("[INFO] Creating New HKLight for %s", label)

	info := model.Info{
		Name:         label,
		Manufacturer: "LIFX",
	}

	lightBulb := accessory.NewLightBulb(info)

	power, _ := light.GetPower()
	lightBulb.SetOn(power)

	color, _ := light.GetColor()
	hue, saturation, brightness := ConvertLIFXColor(color)

	lightBulb.SetBrightness(int(brightness))
	lightBulb.SetSaturation(saturation)
	lightBulb.SetHue(hue)

	transport, err := hap.NewIPTransport(pin, lightBulb.Accessory)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		transport.Start()
	}()

	hkLight = &HKLight{lightBulb.Accessory, nil, transport, lightBulb}

	lightBulb.OnIdentify(func() {
		timeout := 1 * time.Second

		for i := 0; i < 4; i++ {
			ToggleLight(light)
			time.Sleep(timeout)
		}
	})

	lightBulb.OnStateChanged(func(power bool) {
		log.Printf("[INFO] Changed State for %s", label)
		light.SetPower(power)
	})

	updateColor := func(light common.Light) {
		// HAP: [0...360]
		// LIFX: [0...MAX_UINT16]
		hue := lightBulb.GetHue()

		// HAP: [0...100]
		// LIFX: [0...MAX_UINT16]
		saturation := lightBulb.GetSaturation()

		// HAP: [0...100]
		// LIFX: [0...MAX_UINT16]
		brightness := lightBulb.GetBrightness()

		// [HSBKKelvinMin..HSBKKelvinMax]
		kelvin := HSBKKelvinDefault

		lifxHue := math.MaxUint16 * float64(hue) / float64(characteristic.MaxHue)
		lifxSaturation := math.MaxUint16 * float64(saturation) / float64(characteristic.MaxSaturation)
		lifxBrightness := math.MaxUint16 * float64(brightness) / float64(characteristic.MaxBrightness)

		color := common.Color{
			uint16(lifxHue),
			uint16(lifxSaturation),
			uint16(lifxBrightness),
			kelvin,
		}

		light.SetColor(color, 500*time.Millisecond)
	}

	lightBulb.OnHueChanged(func(value float64) {
		log.Printf("[INFO] Changed Hue for %s to %d", label, value)
		updateColor(light)
	})

	lightBulb.OnSaturationChanged(func(value float64) {
		log.Printf("[INFO] Changed Saturation for %s to %d", label, value)
		updateColor(light)
	})

	lightBulb.OnBrightnessChanged(func(value int) {
		log.Printf("[INFO] Changed Brightness for %s to %d", label, value)
		updateColor(light)
	})

	return hkLight
}

func ConvertLIFXColor(color common.Color) (float64, float64, float64) {
	hue := float64(color.Hue) / float64(math.MaxUint16) * float64(characteristic.MaxHue)
	saturation := float64(color.Saturation) / float64(math.MaxUint16) * float64(characteristic.MaxSaturation)
	brightness := float64(color.Brightness) / float64(math.MaxUint16) * float64(characteristic.MaxBrightness)

	return hue, saturation, brightness
}

func ToggleLight(light common.Light) {
	power, _ := light.GetPower()
	light.SetPower(!power)
}

func main() {
	lights = map[uint64]*HKLight{}

	pinArg := flag.String("pin", "", "PIN used to pair the LIFX bulbs with HomeKit")
	verboseArg := flag.Bool("v", false, "Whether or not log output is displayed")

	flag.Parse()

	pin = *pinArg

	if !*verboseArg {
		log.Info = false
		log.Verbose = false
	}

	hap.OnTermination(func() {
		for _, light := range lights {
			light.transport.Stop()
		}

		time.Sleep(100 * time.Millisecond)
		os.Exit(1)
	})

	Connect()
}
