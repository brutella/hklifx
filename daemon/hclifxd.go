package main

import (
    "log"
    "fmt"
    
    "github.com/wolfeidau/lifx"
    "github.com/brutella/hap/app"
    "github.com/brutella/hap/server"
    "github.com/brutella/hap/model/accessory"
    "github.com/brutella/hap/model"
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
    
    acc := accessoryForBulb(bulb)
    
    log.Println("LIFX is now", on)
    acc.SetOn(on)
}

func accessoryForBulb(bulb *lifx.Bulb)model.Switch {
    label := bulb.GetLabel()
    switch_service, found := switches[label]
    if found == true {
        return switch_service
    }
    
    fmt.Println("Create new switch for blub")
    
    info := model.Info{
        Name: label,
        SerialNumber: "001",
        Manufacturer: "LIFX",
        Model: "LIFX",
    }
    
    sw := accessory.NewSwitch(info)
    sw.OnStateChanged(func(on bool) {
        if on == true {
            client.LightOn(bulb)
            fmt.Println("Switch is on")
        } else {
            client.LightOff(bulb)
            fmt.Println("Switch is off")
        }
    })
    
    application.AddAccessory(sw.Accessory)
    switches[label] = sw
    
    return sw
}

var application *app.App
var switches map[string]model.Switch
var client *lifx.Client

func main() {
    switches = map[string]model.Switch{}
    
    conf := app.NewConfig()
    conf.DatabaseDir = "./data"
    conf.BridgeName = "TestBridge" // default "GoBridge"
    
    pwd, _ := server.NewPassword("11122333")
    conf.BridgePassword = pwd // default "001-02-003"
    conf.BridgeManufacturer = "Matthias Hochgatterer" // default "brutella"
    
    var err error
    application, err = app.NewApp(conf)
    if err != nil {
        log.Fatal(err)
    }
    
    ConnectLIFX()
        
    application.Run()
}
