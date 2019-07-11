package service

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/energieip/common-components-go/pkg/dhvac"
	"github.com/energieip/common-components-go/pkg/pconst"
	"github.com/energieip/common-components-go/pkg/tools"
	"github.com/energieip/swh200-rest2mqtt-go/internal/core"
	"github.com/romana/rlog"
)

func (s *Service) sendHello(driver dhvac.Hvac) {
	driverHello := core.HvacHello{
		Mac:             driver.Mac,
		IP:              driver.IP,
		SwitchMac:       driver.SwitchMac,
		IsConfigured:    false,
		Protocol:        "REST",
		FriendlyName:    driver.FriendlyName,
		DumpFrequency:   DefaultTimerDump,
		SoftwareVersion: driver.SoftwareVersion,
	}
	if driver.FullMac != nil {
		driverHello.FullMac = *driver.FullMac
	}
	dump, err := tools.ToJSON(driverHello)
	if err != nil {
		rlog.Errorf("Could not dump HVAC %v status %v", driver.FullMac, err.Error())
		return
	}

	err = s.local.SendCommand("/read/hvac/"+driver.Mac+"/"+pconst.UrlHello, dump)
	if err != nil {
		rlog.Errorf("Could not send hello to the server %v status %v", driver.FullMac, err.Error())
		return
	}
	rlog.Info("Hello " + driver.Mac + " sent to the server")
}

func (s *Service) sendDump(status dhvac.Hvac) {
	//update HVAC info
	token, err := s.hvacLogin(status.IP)
	if err != nil {
		status.Error = 1
		s.hvacs.Set(status.Mac, status)
		dump, _ := status.ToJSON()
		s.local.SendCommand("/read/hvac/"+status.Mac+"/"+pconst.UrlStatus, dump)
		return
	}

	info, err := s.hvacGetStatus(status.IP, token)
	if err != nil {
		status.Error = 2
		s.hvacs.Set(status.Mac, status)
		dump, _ := status.ToJSON()
		s.local.SendCommand("/read/hvac/"+status.Mac+"/"+pconst.UrlStatus, dump)
		return
	}
	status.SpaceCO2 = info.AirRegister.SpaceCO2
	status.OADamper = info.AirRegister.OADamper
	status.OccManCmd1 = info.Regulation.OccManCmd
	status.DewSensor1 = info.Regulation.DewSensor
	status.SpaceTemp1 = info.Regulation.SpaceTemp
	status.HeatCool1 = info.Regulation.HeatCool
	status.CoolOutput1 = info.Regulation.CoolOuput

	infoSetpoint, err := s.getHvacSetpoints(status.IP, token)
	if err != nil {
		status.Error = 2
		s.hvacs.Set(status.Mac, status)
		dump, _ := status.ToJSON()
		s.local.SendCommand("/read/hvac/"+status.Mac+"/"+pconst.UrlStatus, dump)
		return
	}
	status.SetpointUnoccupiedHeat1 = int(infoSetpoint.SetpointUnoccHeat * 10)
	status.SetpointUnoccupiedCool1 = int(infoSetpoint.SetpointUnoccCool * 10)
	status.SetpointOccupiedCool1 = int(infoSetpoint.SetpointOccCool * 10)
	status.SetpointOccupiedHeat1 = int(infoSetpoint.SetpointOccHeat * 10)
	status.SetpointStandbyCool1 = int(infoSetpoint.SetpointStanbyCool * 10)
	status.SetpointStandbyHeat1 = int(infoSetpoint.SetpointStanbyHeat * 10)

	s.hvacs.Set(status.Mac, status)
	s.driversSeen.Set(status.Mac, time.Now().UTC())

	dump, _ := status.ToJSON()
	s.local.SendCommand("/read/hvac/"+status.Mac+"/"+pconst.UrlStatus, dump)
}

func (s *Service) receivedHvacSetup(setup dhvac.HvacSetup) {
	d, errGet := s.hvacs.Get(setup.Mac)
	if !errGet {
		rlog.Error("Cannot find hvac  ", setup.Mac)
		return
	}
	hvac, err := dhvac.ToHvac(d)
	if err != nil {
		rlog.Error("Cannot parse config  ", err.Error())
		return
	}

	if hvac.IsConfigured {
		return
	}

	token, err := s.hvacLogin(hvac.IP)
	if err != nil {
		rlog.Error("Cannot login to: ", err.Error())
		return
	}
	err = s.hvacInit(setup, hvac.IP, token)
	if err != nil {
		rlog.Error("Cannot apply init config: ", err.Error())
		return
	}
	if setup.Group != nil {
		hvac.Group = *setup.Group
	}
	hvac.Label = setup.Label
	hvac.DumpFrequency = setup.DumpFrequency
	hvac.IsConfigured = true
	s.hvacs.Set(hvac.Mac, hvac)
}

func (s *Service) receivedHvacUpdate(conf dhvac.HvacConf) {
	d, errGet := s.hvacs.Get(conf.Mac)
	if !errGet {
		return
	}
	hvac, err := dhvac.ToHvac(d)
	if err != nil {
		return
	}

	if !hvac.IsConfigured {
		return
	}

	if conf.Group != nil {
		hvac.Group = *conf.Group
	}
	if conf.Label != nil {
		hvac.Label = conf.Label
	}
	s.hvacs.Set(hvac.Mac, hvac)

	// token, err := s.hvacLogin(hvac.IP)
	// if err != nil {
	// 	return
	// }
	//TODO

}

func (s *Service) newHvac(new interface{}) error {
	driver, err := core.ToDevice(new)
	if err != nil {
		return err
	}

	token, err := s.hvacLogin(driver.IP)
	if err != nil {
		return err
	}
	info, err := s.hvacGetVersion(driver.IP, token)
	if err != nil || info == nil {
		return err
	}
	hvac := dhvac.Hvac{
		FullMac:         &driver.Mac,
		SwitchMac:       s.Mac,
		Protocol:        "REST",
		IsConfigured:    false,
		FriendlyName:    driver.Mac,
		IP:              driver.IP,
		SoftwareVersion: info.MainAppfirmwVersion,
	}

	submac := strings.SplitN(driver.Mac, ":", 4)
	hvac.Mac = submac[len(submac)-1]

	s.hvacs.Set(hvac.Mac, hvac)
	s.driversSeen.Set(driver.Mac, time.Now().UTC())
	return nil
}

func (s *Service) hvacGetStatus(IP string, token string) (*core.HvacLoop1, error) {
	url := "https://" + IP + "/api/runtime/hvac/loop1"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)

	if err != nil {
		rlog.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	info := core.HvacLoop1{}
	err = json.Unmarshal(body, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (s *Service) hvacGetVersion(IP string, token string) (*core.HvacSysInfo, error) {
	url := "https://" + IP + "/api/systemInfos"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+token)
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)

	if err != nil {
		rlog.Error("Cannot get version: " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	info := core.HvacSysInfo{}
	err = json.Unmarshal(body, &info)
	if err != nil {
		rlog.Error("Cannot get version: " + err.Error())
		return nil, err
	}
	return &info, nil
}

func (s *Service) hvacLogin(IP string) (string, error) {
	url := "https://" + IP + "/api/login"

	user := core.HvacLogin{
		UserKey: s.conf.ClientAPI.Password,
	}

	requestBody, err := json.Marshal(user)
	if err != nil {
		rlog.Error("Cannot send request to: " + err.Error())
		return "", err
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	req.Header.Add("Content-Type", "application/json")
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)

	if err != nil {
		rlog.Error("Cannot send request to: " + err.Error())
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", NewError("Incorrect Status code")
	}

	body, err := ioutil.ReadAll(resp.Body)

	auth := core.HvacAuth{}
	err = json.Unmarshal(body, &auth)
	if err != nil {
		rlog.Error("Cannot parse body: " + err.Error())
		return "", err
	}
	return auth.AccessToken, nil
}

func (s *Service) getHvacSetpoints(IP string, token string) (*core.HvacSetPoints, error) {
	url := "https://" + IP + "/api/setup/hvac/setpoint/loop1"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+token)
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)

	if err != nil {
		rlog.Error(err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, NewError("Incorrect Status code")
	}

	body, err := ioutil.ReadAll(resp.Body)

	status := core.HvacSetPoints{}
	err = json.Unmarshal(body, &status)
	if err != nil {
		rlog.Error("Cannot parse body: " + err.Error())
		return nil, err
	}
	return &status, nil
}

func (s *Service) hvacInit(setup dhvac.HvacSetup, IP string, token string) error {
	url := "https://" + IP + "/api/setup/hvac/setpoint/loop1"

	OccCool := float32(20)
	if setup.SetpointCoolOccupied != nil {
		OccCool = float32(*setup.SetpointCoolOccupied) / 10
	}
	OccHeat := float32(20)
	if setup.SetpointHeatOccupied != nil {
		OccHeat = float32(*setup.SetpointHeatOccupied) / 10
	}
	UnoccHeat := float32(20)
	if setup.SetpointHeatInoccupied != nil {
		UnoccHeat = float32(*setup.SetpointHeatInoccupied) / 10
	}
	UnoccCool := float32(20)
	if setup.SetpointCoolInoccupied != nil {
		UnoccCool = float32(*setup.SetpointCoolInoccupied) / 10
	}
	CoolStandby := float32(20)
	if setup.SetpointCoolStandby != nil {
		CoolStandby = float32(*setup.SetpointCoolStandby) / 10
	}
	HeatStandby := float32(20)
	if setup.SetpointHeatStandby != nil {
		HeatStandby = float32(*setup.SetpointHeatStandby) / 10
	}

	config := core.HvacSetPoints{
		SetpointOccCool:    OccCool,
		SetpointOccHeat:    OccHeat,
		SetpointUnoccHeat:  UnoccHeat,
		SetpointUnoccCool:  UnoccCool,
		SetpointStanbyCool: CoolStandby,
		SetpointStanbyHeat: HeatStandby,
	}

	requestBody, err := json.Marshal(config)
	if err != nil {
		return err
	}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+token)
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}
	client := &http.Client{Transport: transCfg}
	_, err = client.Do(req)

	if err != nil {
		rlog.Error(err.Error())
		return err
	}
	return nil
}
