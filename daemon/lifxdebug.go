package main

import (
    "github.com/wolfeidau/lifx"
    "log"
    "sync"
    "time"
)

var mutex sync.Mutex
var bulbs map[string]*lifx.Bulb
var added chan *lifx.Bulb
var removed chan *lifx.Bulb

func addBulb(bulb *lifx.Bulb) {
    addr := bulb.GetLifxAddress()
    if _, ok := bulbs[addr]; ok == false {
        mutex.Lock()
        bulbs[addr] = bulb
        mutex.Unlock()
        
        added <- bulb
    }
}

func keepAliveAll(c *lifx.Client) {
    interval := 10 * time.Second
    ticker := time.NewTicker(5 * time.Second)
	for _ = range ticker.C {
        mutex.Lock()
		for addr, bulb := range bulbs {
			delta := time.Now().UnixNano() - bulb.LastSeen.UnixNano()
            if time.Duration(delta) > interval {
                log.Println("Bulb alive?")
			    err := c.GetBulbState(bulb)
                if err != nil {
                    log.Println(err)
                }
			}
		}
        mutex.Unlock()
	}
}

func main() {
    bulbs = map[string]*lifx.Bulb{}
    added = make(chan *lifx.Bulb)
    removed = make(chan *lifx.Bulb)
    
    c := lifx.NewClient()
    err := c.StartDiscovery()

    if err != nil {
        log.Fatalf("Could not start discovery %s", err)
    }
    
    go func() {
        sub := c.Subscribe()
        for {
            event := <-sub.Events
            switch event := event.(type) {
            case *lifx.Gateway:
                log.Println("Gateway", event)
            case *lifx.Bulb:  
                log.Println("Bulb event", event)
                addBulb(event)
                event.SetStateHandler(func(newState *lifx.BulbState) {
                    log.Println("Updated", newState)
                    addBulb(event)
                })
            default:
                log.Println("Event", event)
            }
        }
    }()
    
    go keepAliveAll(c)
    
    for {
        select {
        case bulb := <- added:
            log.Println("Added", bulb)
        case bulb := <- removed:
            log.Println("Removed", bulb)
        }
    }
}