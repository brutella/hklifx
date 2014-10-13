package main

import (
    "log"

    "github.com/wolfeidau/lifx"
    
    "github.com/brutella/virtual-device"
    "github.com/brutella/homecontrol"
)

var switches map[string]homecontrol.Switch
var filesystem virtual.FileSystem

func main() {
    filesystem := virtual.NewFileSystem()
    switches := map[string]homecontrol.Switch{}
    
    c := lifx.NewClient()
    err := c.StartDiscovery()

    if err != nil {
        log.Fatalf("Could not find bulb %s", err)
    }

    go func() {
        sub := c.Subscribe()
        for {
            event := <-sub.Events

            switch event := event.(type) {
            case *lifx.Gateway:
                log.Printf("Gateway Update %v", event)
            case *lifx.Bulb:
                label := event.GetLabel()
                on := true
                if event.GetPower() == 0 {
                    on = false
                }
                
                sw :=  switches[label]
                if sw == nil {
                    sw = homecontrol.NewSwitch(label, event.GetLifxAddress(), "LIFX", "LIFX", on)
                    err := filesystem.Mount(sw, "./" + label)
                    if err != nil {
                        log.Println("Could not mount switch.", err)
                    }
                    filesystem.Watch(sw, func() {
                        if sw.IsOn() {
                            c.LightOn(event)
                        } else {
                            c.LightOff(event)
                        }
                    })
                    switches[label] = sw
                }
                
                sw.SetOn(on)
                log.Printf("Update switch")
                filesystem.Write(sw)
            default:
                log.Printf("Event %v", event)
            }

        }
    }()

    select {
        
    }
}