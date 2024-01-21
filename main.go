package main

import (
	"flag"
	"fmt"
	"log"
	"time"
)

const version = "0.1.0"

var (
	h bool
	v bool

	scan_net  string
	scan_port string
	dname     string

	interval time.Duration
)

func init() {
	log.SetFlags(log.LstdFlags)

	flag.BoolVar(&h, "h", false, "")
	flag.BoolVar(&v, "v", false, "")

	flag.StringVar(&scan_net, "network", "", "Scanning target network")
	flag.StringVar(&scan_port, "port", "5055", "Scanning target port")
	flag.StringVar(&dname, "dname", "", "The hostname mapped from ZSpace IP to /etc/hosts")
	flag.DurationVar(&interval, "interval", 5*time.Minute, "Run interval, if 0, run once (in minute)")
}

func main() {
	flag.Parse()
	if h {
		flag.Usage()
	} else if v {
		fmt.Println(version)
	} else {
		if scan_net == "" || scan_port == "" || dname == "" {
			log.Fatalln("invalid flag options")
		}
		if interval == 0*time.Minute {
			update_zspace_hostip()
		} else {
			for {
				update_zspace_hostip()
				time.Sleep(interval)
			}
		}
	}
}
