package main

import (
	"example/service"
	"flag"
	"fmt"
	"os"
)

var listenPort *int

func init() {
	showHelp := flag.Bool("h", false, "Show usage")

	listenPort = flag.Int("p", 8080, "Specify the port")
	flag.IntVar(listenPort, "port", 8080, "Specify the port")

	flag.Parse()

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
		flag.PrintDefaults()
	}

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}
}

func main() {

	eurekaconfig := service.EurekaConfig{
		Url:         []string{"http://localhost:8761/eureka"},
		ServiceName: "GO-SERVICE",
		HostName:    "localhost",
		Port:        *listenPort,
	}

	err := service.Init(*listenPort, eurekaconfig)
	if err != nil {
		fmt.Println(err)
	}

	service.Start()
}
