package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/energieip/swh200-rest2mqtt-go/internal/core"

	pkg "github.com/energieip/common-components-go/pkg/service"
	"github.com/romana/rlog"
)

type arrayString []string

func (i *arrayString) String() string {
	return strings.Join(*i, ",")
}

func (i *arrayString) Set(value string) error {
	if value == "" {
		return nil
	}

	var list []string
	for _, in := range strings.Split(value, ",") {
		list = append(list, in)
	}

	*i = arrayString(list)
	return nil
}

func (i *arrayString) Get() interface{} { return []string(*i) }

type arrayInt []int

func (i *arrayInt) Set(val string) error {
	if val == "" {
		return nil
	}

	var list []int
	for _, in := range strings.Split(val, ",") {
		i, err := strconv.Atoi(in)
		if err != nil {
			return err
		}

		list = append(list, i)
	}

	*i = arrayInt(list)
	return nil
}

func (i *arrayInt) Get() interface{} { return []int(*i) }

func (i *arrayInt) String() string {
	var list []string
	for _, in := range *i {
		list = append(list, strconv.Itoa(in))
	}
	return strings.Join(list, ",")
}

func main() {
	var confFile string
	var newIP string
	var newMac string

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flag.StringVar(&confFile, "config", "", "Specify an alternate configuration file.")
	flag.StringVar(&confFile, "c", "", "Specify an alternate configuration file.")
	flag.StringVar(&newIP, "ip", "", "IP")
	flag.StringVar(&newIP, "i", "", "IP")
	flag.StringVar(&newMac, "mac", "", "mac address")
	flag.StringVar(&newMac, "m", "", "mac address")
	flag.Parse()

	conf, err := pkg.ReadServiceConfig(confFile)
	if err != nil {
		rlog.Error("Cannot parse configuration file " + err.Error())
		os.Exit(1)
	}
	os.Setenv("RLOG_LOG_LEVEL", conf.LogLevel)
	os.Setenv("RLOG_LOG_NOTIME", "yes")
	rlog.UpdateEnv()

	user := core.Device{
		Mac: strings.ToUpper(newMac),
		IP:  newIP,
	}

	requestBody, err := json.Marshal(user)
	if err != nil {
		rlog.Error(err.Error())
		os.Exit(1)
	}

	url := "https://" + conf.InternalAPI.IP + ":" + conf.InternalAPI.Port + "/v1.0/driver/new"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	transCfg := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	req.Close = true
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		rlog.Error(err.Error())
		os.Exit(1)
	}

	body, err := ioutil.ReadAll(resp.Body)

	rlog.Info("Device " + newMac + " successfully added " + string(body))
}
