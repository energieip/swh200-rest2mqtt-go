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
	dump, err := tools.ToJSON(driverHello)
	if err != nil {
		rlog.Errorf("Could not dump HVAC %v status %v", driver.Mac, err.Error())
		return
	}

	err = s.local.SendCommand("/read/hvac/"+driver.Mac+"/"+pconst.UrlHello, dump)
	if err != nil {
		rlog.Errorf("Could not send hello to the server %v status %v", driver.Mac, err.Error())
		return
	}
}

func (s *Service) sendDump(status dhvac.Hvac) {
	token, err := s.hvacLogin(status.IP)
	if err != nil {
		status.Error = 1
		s.hvacs.Set(strings.ToUpper(status.Mac), status)
		dump, _ := status.ToJSON()
		s.local.SendCommand("/read/hvac/"+status.Mac+"/"+pconst.UrlStatus, dump)
		return
	}

	info, err := s.hvacGetStatus(status.IP, token)
	if err != nil {
		status.Error = 2
		s.hvacs.Set(strings.ToUpper(status.Mac), status)
		dump, _ := status.ToJSON()
		s.local.SendCommand("/read/hvac/"+status.Mac+"/"+pconst.UrlStatus, dump)
		return
	}
	status.LinePower = 10
	status.SpaceCO2 = info.AirRegister.SpaceCO2
	status.OADamper = info.AirRegister.OADamper
	status.OccManCmd1 = info.Regulation.OccManCmd
	status.DewSensor1 = info.Regulation.DewSensor
	status.SpaceTemp1 = int(info.Regulation.SpaceTemp * 10)
	status.HeatCool1 = info.Regulation.HeatCool
	status.CoolOutput1 = info.Regulation.CoolOuput
	status.EffectSetPoint1 = int(info.Regulation.EffectifSetPoint * 10)
	status.HoldOff1 = info.Regulation.WindowHoldOff

	infoSetpoint, err := s.getHvacSetpoints(status.IP, token)
	if err != nil {
		status.Error = 2
		s.hvacs.Set(strings.ToUpper(status.Mac), status)
		dump, _ := status.ToJSON()
		s.local.SendCommand("/read/hvac/"+status.Mac+"/"+pconst.UrlStatus, dump)
		return
	}

	infoRegul, err := s.getHvacSetupRegulation(status.IP, token)
	if err != nil {
		status.Error = 2
		s.hvacs.Set(strings.ToUpper(status.Mac), status)
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
	status.TemperatureOffsetStep = int(infoRegul.TemperOffsetStep * 10)
	status.Shift = int((float32(info.Regulation.OffsetTemp) * infoRegul.TemperOffsetStep) * 10)
	status.Error = 0

	s.hvacs.Set(strings.ToUpper(status.Mac), status)
	s.driversSeen.Set(strings.ToUpper(status.Mac), time.Now().UTC())

	dump, _ := status.ToJSON()
	s.local.SendCommand("/read/hvac/"+status.Mac+"/"+pconst.UrlStatus, dump)
}

func (s *Service) receivedHvacSetup(setup dhvac.HvacSetup) {
	d, errGet := s.hvacs.Get(strings.ToUpper(setup.Mac))
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

	err = s.setHvacSetupRegulation(setup, hvac.IP, token)
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
	s.hvacs.Set(strings.ToUpper(hvac.Mac), hvac)
}

func (s *Service) receivedHvacUpdate(conf dhvac.HvacConf) {
	d, errGet := s.hvacs.Get(strings.ToUpper(conf.Mac))
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
	s.hvacs.Set(strings.ToUpper(hvac.Mac), hvac)

	token, err := s.hvacLogin(hvac.IP)
	if err != nil {
		return
	}
	s.setHvacRuntime(conf, *hvac, hvac.IP, token)
}

func (s *Service) newHvac(new interface{}) error {
	driver, err := core.ToDevice(new)
	if err != nil || driver == nil {
		return err
	}
	rlog.Info("New HVAC plugged ", driver.Mac)
	var token string
	token, err = s.hvacLogin(driver.IP)
	if err != nil {
		// wait for device to be up and ready
		time.Sleep(120 * time.Second)
		rlog.Info("Retry connection to HVAC ", driver.Mac)
		token, err = s.hvacLogin(driver.IP)
		if err != nil {
			return err
		}
	}

	info, err := s.hvacGetVersion(driver.IP, token)
	if err != nil || info == nil {
		return err
	}
	hvac := dhvac.Hvac{
		Mac:             driver.Mac,
		SwitchMac:       s.Mac,
		Protocol:        "REST",
		IsConfigured:    false,
		FriendlyName:    driver.Mac,
		IP:              driver.IP,
		SoftwareVersion: info.SoftwareVersion,
	}

	s.hvacs.Set(strings.ToUpper(hvac.Mac), hvac)
	s.driversSeen.Set(strings.ToUpper(driver.Mac), time.Now().UTC())
	return nil
}

func (s *Service) hvacGetStatus(IP string, token string) (*core.HvacLoop1, error) {
	url := "https://" + IP + "/api/runtime/hvac/loop1"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	req.Close = true
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
	req.Close = true
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
	req.Close = true
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
	req.Close = true
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

func (s *Service) setHvacRuntime(conf dhvac.HvacConf, status dhvac.Hvac, IP string, token string) error {
	url := "https://" + IP + "/api/runtime/hvac/loop1"

	param := core.HvacLoopCtrl{}

	if conf.WindowStatus != nil {
		if param.Regulation == nil {
			airReg := core.HvacRegulationCtrl{}
			param.Regulation = &airReg
		}
		value := 0
		if *conf.WindowStatus == true {
			value = 1
		}
		param.Regulation.WindowHoldOff = &value
	}
	if conf.Temperature != nil {
		if param.Regulation == nil {
			airReg := core.HvacRegulationCtrl{}
			param.Regulation = &airReg
		}
		spaceTemp := float32(*conf.Temperature) / 10.0
		param.Regulation.SpaceTemp = &spaceTemp
	}
	if conf.CO2 != nil {
		if param.AirRegister == nil {
			airRegister := core.HvacAirRegisterCtrl{}
			param.AirRegister = &airRegister
		}
		co2 := *conf.CO2 / 10
		param.AirRegister.SpaceCO2 = &co2
	}
	if conf.Shift != nil {
		if param.Regulation == nil {
			airReg := core.HvacRegulationCtrl{}
			param.Regulation = &airReg
		}
		step := float32(status.TemperatureOffsetStep) / 10.0
		if step > 0 {
			offsetTemp := (float32(*conf.Shift) / 10.0) / step
			offset := int(offsetTemp)
			param.Regulation.OffsetTemp = &offset
		}
	}

	if conf.Presence != nil {
		if param.Regulation == nil {
			airReg := core.HvacRegulationCtrl{}
			param.Regulation = &airReg
		}
		pres := *conf.Presence
		if pres {
			mode := dhvac.OCCUPANCY_COMFORT
			param.Regulation.OccManCmd = &mode
		} else {
			mode := dhvac.OCCUPANCY_ECONOMY
			param.Regulation.OccManCmd = &mode
		}
	}

	if conf.HeatCool != nil {
		if param.Regulation == nil {
			airReg := core.HvacRegulationCtrl{}
			param.Regulation = &airReg
		}
		param.Regulation.HeatCool = conf.HeatCool
	}

	requestBody, err := json.Marshal(param)
	if err != nil {
		return err
	}
	rlog.Infof("Send new parameters to HVAC %v: %v", status.Mac, string(requestBody))

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+token)
	req.Close = true
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)

	if err != nil {
		rlog.Error(err.Error())
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return NewError("Incorrect Status code")
	}

	return nil
}

func (s *Service) setHvacSetupRegulation(setup dhvac.HvacSetup, IP string, token string) error {
	url := "https://" + IP + "/api/setup/hvac/regulation"

	if setup.TemperatureOffsetStep == nil {
		return nil
	}

	config := core.HvacSetupRegulation{
		TemperOffsetStep: float32(*setup.TemperatureOffsetStep) / 10.0,
	}

	requestBody, err := json.Marshal(config)
	if err != nil {
		return err
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+token)
	req.Close = true
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)

	if err != nil {
		rlog.Error(err.Error())
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return NewError("Incorrect Status code")
	}

	return nil
}

func (s *Service) getHvacSetupRegulation(IP string, token string) (*core.HvacSetupRegulation, error) {
	url := "https://" + IP + "/api/setup/hvac/regulation"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+token)
	req.Close = true
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

	status := core.HvacSetupRegulation{}
	err = json.Unmarshal(body, &status)
	if err != nil {
		rlog.Error("Cannot parse body: " + err.Error())
		return nil, err
	}
	return &status, nil
}

func (s *Service) hvacInit(setup dhvac.HvacSetup, IP string, token string) error {
	url := "https://" + IP + "/api/setup/hvac/setpoint/loop1"

	OccCool := float32(19)
	if setup.SetpointCoolOccupied != nil {
		OccCool = float32(*setup.SetpointCoolOccupied) / 10
	}
	OccHeat := float32(26)
	if setup.SetpointHeatOccupied != nil {
		OccHeat = float32(*setup.SetpointHeatOccupied) / 10
	}
	UnoccHeat := float32(30)
	if setup.SetpointHeatInoccupied != nil {
		UnoccHeat = float32(*setup.SetpointHeatInoccupied) / 10
	}
	UnoccCool := float32(15)
	if setup.SetpointCoolInoccupied != nil {
		UnoccCool = float32(*setup.SetpointCoolInoccupied) / 10
	}
	CoolStandby := float32(28)
	if setup.SetpointCoolStandby != nil {
		CoolStandby = float32(*setup.SetpointCoolStandby) / 10
	}
	HeatStandby := float32(17)
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
	req.Close = true
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
