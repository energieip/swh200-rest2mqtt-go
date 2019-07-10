package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/energieip/common-components-go/pkg/service"
	firm "github.com/energieip/swh200-rest2mqtt-go/internal/service"
)

func main() {
	var confFile string
	var service service.IService

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.StringVar(&confFile, "config", "", "Specify an alternate configuration file.")
	flag.StringVar(&confFile, "c", "", "Specify an alternate configuration file.")
	flag.Parse()

	s := firm.Service{}
	service = &s
	err := service.Initialize(confFile)
	if err != nil {
		log.Println("Error during service connexion " + err.Error())
		os.Exit(1)
	}
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Received SIGTERM")
		service.Stop()
		os.Exit(0)
	}()

	err = service.Run()
	if err != nil {
		log.Println("Error during service execution " + err.Error())
		os.Exit(1)
	}
}
