package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/ilikebits/jeb/krpc"
)

func main() {
	// Configure logging.
	log.SetFlags(log.Lmicroseconds | log.Llongfile)

	// Parse flags.
	addr := flag.String("addr", "127.0.0.1:50000", "server TCP address")
	flag.Parse()

	// Dial client.
	c, err := krpc.Dial(*addr)
	if err != nil {
		panic(err)
	}

	// Call KRPC.GetStatus()
	stat, err := c.Status()
	if err != nil {
		panic(err)
	}
	log.Printf("%#v", stat)

	// Call KRPC.SpaceCenter.ActiveVessel()
	v, err := c.Vessel()
	if err != nil {
		panic(err)
	}
	log.Printf("%#v", v)

	// Call vessel.Flight()
	f, err := v.Flight()
	if err != nil {
		panic(err)
	}
	log.Printf("%#v", f)

	for {
		// Call flight.SurfaceAltitude()
		a, err := f.SurfaceAltitude()
		if err != nil {
			panic(err)
		}
		fmt.Printf("%#v\n", a)

		time.Sleep(1 * time.Millisecond)
	}
}
