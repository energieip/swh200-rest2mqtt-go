package service

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/energieip/common-components-go/pkg/dhvac"

	"github.com/energieip/swh200-rest2mqtt-go/internal/api"
	"github.com/energieip/swh200-rest2mqtt-go/internal/core"
	net "github.com/energieip/swh200-rest2mqtt-go/internal/network"

	pkg "github.com/energieip/common-components-go/pkg/service"
	"github.com/energieip/common-components-go/pkg/tools"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/romana/rlog"
)

const (
	DefaultTimerDump = 1000
)

type systemError struct {
	s string
}

func (e *systemError) Error() string {
	return e.s
}

// NewError raise an error
func NewError(text string) error {
	return &systemError{text}
}

//Service content
type Service struct {
	local        net.ServerNetwork //local broker for drivers
	Mac          string            //Switch mac address
	label        string
	events       chan string
	timerDump    time.Duration //in seconds
	ip           string
	isConfigured bool
	hvacs        cmap.ConcurrentMap
	conf         pkg.ServiceConfig
	clientID     string
	driversSeen  cmap.ConcurrentMap
	api          *api.API
}

//Initialize service
func (s *Service) Initialize(confFile string) error {
	s.events = make(chan string)
	s.hvacs = cmap.New()
	s.driversSeen = cmap.New()

	conf, err := pkg.ReadServiceConfig(confFile)
	if err != nil {
		rlog.Error("Cannot parse configuration file " + err.Error())
		return err
	}
	s.conf = *conf

	mac, _ := tools.GetNetworkInfo()
	s.Mac = mac

	os.Setenv("RLOG_LOG_LEVEL", conf.LogLevel)
	os.Setenv("RLOG_LOG_NOTIME", "yes")
	os.Setenv("RLOG_TIME_FORMAT", "2006/01/06 15:04:05.000")
	rlog.UpdateEnv()
	rlog.Info("Starting rest2mqtt service")

	s.timerDump = DefaultTimerDump

	broker, err := net.CreateServerNetwork()
	if err != nil {
		rlog.Error("Cannot connect to broker " + conf.LocalBroker.IP + " error: " + err.Error())
		return err
	}
	s.local = *broker

	go s.local.Connect(*conf)
	web := api.InitAPI(*conf)
	s.api = web
	go s.coldBootStart()
	rlog.Info("rest2mqtt service started")
	return nil
}

func (s *Service) coldBootStart() {
	nmap := exec.Command("nmap", "-sP", "10.0.0.0/24")
	filter := exec.Command("awk", "/Nmap scan report for/{printf $5;}/MAC Address:/{print \" => \"$3;}")

	r, w := io.Pipe()
	nmap.Stdout = w
	filter.Stdin = r

	var result bytes.Buffer
	filter.Stdout = &result

	rlog.Info("Start coldBoot device Scan")

	nmap.Start()
	filter.Start()
	nmap.Wait()
	w.Close()
	filter.Wait()

	for _, line := range strings.Split(result.String(), "\n") {
		elts := strings.Split(line, " => ")
		if len(elts) != 2 {
			continue
		}
		mac := elts[1]
		ip := elts[0]
		if strings.HasPrefix(mac, "F2:23") {
			rlog.Info("Skip EIP device: ", mac)
			continue
		}
		device := core.Device{
			Mac: mac,
			IP:  ip,
		}
		rlog.Info("coldBoot found device : ", device)
		s.newHvac(device)
	}
	rlog.Info("End coldBoot device Scan")
}

//Stop service
func (s *Service) Stop() {
	rlog.Info("Stopping rest2mqtt service")
	s.local.Disconnect()
	rlog.Info("rest2mqtt service stopped")
}

func (s *Service) cronDump() {
	timerDump := time.NewTicker(s.timerDump * time.Millisecond)
	for {
		select {
		case <-timerDump.C:
			for _, v := range s.hvacs.Items() {
				driver, _ := dhvac.ToHvac(v)
				if driver.IsConfigured {
					s.sendDump(*driver)
				} else {
					s.sendHello(*driver)
				}
			}
		}
	}
}

//Run service mainloop
func (s *Service) Run() error {
	//TODO manage case restart with already connected device???
	go s.cronDump()
	for {
		select {
		case evtUpdate := <-s.local.EventsConf:
			for _, event := range evtUpdate {
				s.receivedHvacUpdate(event)
			}

		case evtSetup := <-s.local.EventsSetup:
			for _, event := range evtSetup {
				s.receivedHvacSetup(event)
			}

		case evtAPI := <-s.api.EventsToBackend:
			for evtType, content := range evtAPI {
				switch evtType {
				case "newDevice":
					go s.newHvac(content)
				}
			}
		}
	}
	return nil
}
