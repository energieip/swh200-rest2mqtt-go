package service

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
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

func (s *Service) sendRefresh(status dhvac.Hvac) {
	token, err := s.hvacLogin(status.IP)
	if err != nil {
		rlog.Error("Cannot Login to " + status.Mac)
		status.Error = 1
		s.hvacs.Set(strings.ToUpper(status.Mac), status)
		return
	}
	time.Sleep(50 * time.Millisecond)
	s.driversSeen.Set(strings.ToUpper(status.Mac), time.Now().UTC())

	info, err := s.hvacGetStatus(status.IP, token)
	if err != nil {
		rlog.Error("Cannot get status from " + status.Mac)
		status.Error = 2
		s.hvacs.Set(strings.ToUpper(status.Mac), status)
		return
	}
	status.LinePower = 10
	status.SpaceCO2 = info.AirRegister.SpaceCO2
	status.OADamper = info.AirRegister.OADamper
	status.SpaceHygro = int(info.AirRegister.SpaceHygroRel * 10)
	status.OccManCmd1 = info.Regulation.OccManCmd
	status.DewSensor1 = info.Regulation.DewSensor
	status.SpaceTemp1 = int(info.Regulation.SpaceTemp * 10)
	status.HeatCool1 = info.Regulation.HeatCool
	time.Sleep(50 * time.Millisecond)
	maintenance, _ := s.getHvacMaintenanceMode(status.IP, token)
	if maintenance == nil {
		status.Error = 2
		rlog.Error("Cannot get maintenance info from " + status.Mac)
		s.hvacs.Set(strings.ToUpper(status.Mac), status)
		return
	}
	if maintenance.Running != true {
		status.HeatCool1 = dhvac.HVAC_MODE_TEST
	}
	status.CoolOutput1 = info.Regulation.CoolOutput
	status.HeatOutput1 = info.Regulation.HeatOutput
	status.EffectSetPoint1 = int(info.Regulation.EffectifSetPoint * 10)
	status.HoldOff1 = info.Regulation.WindowHoldOff
	time.Sleep(50 * time.Millisecond)
	infoSetpoint, err := s.getHvacSetpoints(status.IP, token)
	if err != nil {
		status.Error = 2
		rlog.Error("Cannot get hvacSetpoints info from " + status.Mac)
		s.hvacs.Set(strings.ToUpper(status.Mac), status)
		return
	}
	time.Sleep(50 * time.Millisecond)
	infoRegul, err := s.getHvacSetupRegulation(status.IP, token)
	if err != nil {
		status.Error = 2
		rlog.Error("Cannot get hvacSetupRegulation info from " + status.Mac)
		s.hvacs.Set(strings.ToUpper(status.Mac), status)
		return
	}
	time.Sleep(50 * time.Millisecond)
	inputValues, err := s.getHvacSetupInputs(status.IP, token)
	if err != nil {
		status.Error = 2
		rlog.Error("Cannot get hvacSetupInputs info from " + status.Mac)
		s.hvacs.Set(strings.ToUpper(status.Mac), status)
		return
	}
	time.Sleep(50 * time.Millisecond)
	outputValues, err := s.getHvacSetupOutputs(status.IP, token)
	if err != nil {
		status.Error = 2
		rlog.Error("Cannot get hvacSetupOutputs info from " + status.Mac)
		s.hvacs.Set(strings.ToUpper(status.Mac), status)
		return
	}

	status.SetpointUnoccupiedHeat1 = int(infoSetpoint.SetpointUnoccHeat * 10)
	status.SetpointUnoccupiedCool1 = int(infoSetpoint.SetpointUnoccCool * 10)
	status.SetpointOccupiedCool1 = int(infoSetpoint.SetpointOccCool * 10)
	status.SetpointOccupiedHeat1 = int(infoSetpoint.SetpointOccHeat * 10)
	status.SetpointStandbyCool1 = int(infoSetpoint.SetpointStanbyCool * 10)
	status.SetpointStandbyHeat1 = int(infoSetpoint.SetpointStanbyHeat * 10)
	status.TemperatureOffsetStep = int(infoRegul.TemperOffsetStep * 10)
	status.InputE1 = inputValues.InputE1
	status.InputE2 = inputValues.InputE2
	status.InputE3 = inputValues.InputE3
	status.InputE4 = inputValues.InputE4
	status.InputE5 = inputValues.InputE5
	status.InputE6 = inputValues.InputE6
	status.InputC1 = inputValues.InputC1
	status.InputC2 = inputValues.InputC2
	status.OutputY5 = outputValues.OutputY5
	status.OutputY6 = outputValues.OutputY6
	status.OutputY7 = outputValues.OutputY7
	status.OutputY8 = outputValues.OutputY8
	status.OutputYa = outputValues.OutputYa
	status.OutputYb = outputValues.OutputYb

	time.Sleep(50 * time.Millisecond)
	testValues, err := s.getHvacMaintenanceOutputs(status.IP, token)
	if err != nil {
		status.Error = 2
		rlog.Error("Cannot get maintenance output info from " + status.Mac)
		s.hvacs.Set(strings.ToUpper(status.Mac), status)
		return
	}
	time.Sleep(100 * time.Millisecond)
	infoVersion, err := s.hvacGetVersion(status.IP, token)
	if err != nil || infoVersion == nil {
		status.Error = 2
		rlog.Error("Cannot get info version from " + status.Mac)
		s.hvacs.Set(strings.ToUpper(status.Mac), status)
		return
	}
	status.SoftwareVersion = infoVersion.SoftwareVersion

	status.Forcing6WaysValve = testValues.OutputY5
	status.ForcingDamper = testValues.OutputY6
	status.Shift = int((float32(info.Regulation.OffsetTemp) * infoRegul.TemperOffsetStep) * 10)
	status.TemperatureSelect = int(info.Regulation.EffectifSetPoint*10) + status.Shift
	status.Error = 0

	s.hvacs.Set(strings.ToUpper(status.Mac), status)
}

func (s *Service) sendDump(status dhvac.Hvac) {
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

	err = s.setHvacSetupRegulation(setup, hvac.IP, token)
	if err != nil {
		rlog.Error("Cannot apply init config: ", err.Error())
		return
	}
	err = s.setHvacSetupInputs(setup, hvac.IP, token)
	if err != nil {
		rlog.Error("Cannot apply inputs config: ", err.Error())
		return
	}
	err = s.setHvacSetupOutputs(setup, hvac.IP, token)
	if err != nil {
		rlog.Error("Cannot apply outputs config: ", err.Error())
		return
	}
	err = s.hvacInit(setup, hvac.IP, token)
	if err != nil {
		rlog.Error("Cannot apply init config: ", err.Error())
		return
	}
	err = s.setHvacSetupAirRegister(setup, hvac.IP, token)
	if err != nil {
		rlog.Error("Cannot airRegister config: ", err.Error())
		return
	}
	if setup.Group != nil {
		hvac.Group = *setup.Group
	}
	hvac.Label = setup.Label
	hvac.DumpFrequency = DefaultTimerDump
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
		rlog.Error("Cannot get token info from " + conf.Mac)
		return
	}
	s.setHvacRuntime(conf, *hvac, hvac.IP, token)
	s.hvacSetAFConfig(conf, hvac.IP, token)
}

func (s *Service) updateHvac(IP string) error {
	url := "http://" + IP + ":3000/api/updateParam"

	config := core.HvacUpdateParams{
		TftpServerIP: "10.0.0.2",
		StartUpdate:  true,
	}

	requestBody, err := json.Marshal(config)
	if err != nil {
		return err
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("x-access-token", s.conf.ClientAPI.URLToken)
	req.Close = true
	transCfg := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		rlog.Error(err.Error())
		return err
	}
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)

		rlog.Errorf("%v Received UpdateHvac status code %v, body %v", IP, resp.StatusCode, string(body))
		return NewError("Incorrect Status code: " + strconv.Itoa((resp.StatusCode)))
	}
	rlog.Info("Update finished successfully")
	return nil
}

func (s *Service) updateHvacNewAPI(IP string, token string) error {
	url := "https://" + IP + "/api/updateParam"

	config := core.HvacUpdateParams{
		TftpServerIP: "10.0.0.2",
		StartUpdate:  true,
	}

	requestBody, err := json.Marshal(config)
	if err != nil {
		return err
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("x-access-token", token)
	req.Header.Set("authorization", "Bearer "+token)
	req.Close = true
	transCfg := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		rlog.Error(err.Error())
		return err
	}
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)

		rlog.Errorf("%v Received updateHvacNewAPI status code %v, body %v", IP, resp.StatusCode, string(body))
		return NewError("Incorrect Status code: " + strconv.Itoa((resp.StatusCode)))
	}
	rlog.Info("Update finished successfully")
	return nil
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
			rlog.Info("Try to update from modbus to REST", driver.Mac)
			errF := s.updateHvac(driver.IP)
			rlog.Error("Update arcom", errF, driver.Mac)
			return err
		}
	}

	info, err := s.hvacGetVersion(driver.IP, token)
	if err != nil || info == nil {
		return err
	}
	rlog.Infof("For %v (%v) Get version %v and expect %v", driver.Mac, driver.IP, info.SoftwareVersion, s.conf.ClientAPI.APIVersion)
	if info.SoftwareVersion != s.conf.ClientAPI.APIVersion {
		time.Sleep(50 * time.Millisecond)
		errF := s.updateHvacNewAPI(driver.IP, token)
		rlog.Error("Update arcom", errF, driver.Mac)
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

func (s *Service) reloadHvac(new interface{}) error {
	driver, err := core.ToDevice(new)
	if err != nil || driver == nil {
		return err
	}
	rlog.Info("Reload HVAC config ", driver.Mac)
	hvac, ok := s.hvacs.Get(strings.ToUpper(driver.Mac))
	if ok {
		// check for IP changing
		d, _ := dhvac.ToHvac(hvac)
		rlog.Info("Change IP info for " + driver.Mac + " to " + driver.IP + " (was IP: " + d.IP + " )")
		d.IP = driver.IP
		s.hvacs.Set(strings.ToUpper(d.Mac), d)
		return nil
	}
	return s.newHvac(new)
}

func (s *Service) hvacGetStatus(IP string, token string) (*core.HvacLoop1, error) {
	url := "https://" + IP + "/api/runtime/hvac/loop1"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	req.Close = true
	transCfg := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		rlog.Error(err.Error())
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		rlog.Errorf("%v Received hvacGetStatus status code %v, body %v", IP, resp.StatusCode, string(body))
		return nil, NewError("Incorrect Status code")
	}

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
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		rlog.Error("Cannot get version: " + err.Error())
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		rlog.Errorf("%v Received hvacGetVersion status code %v, body %v", IP, resp.StatusCode, string(body))
		return nil, NewError("Incorrect Status code")
	}

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
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		rlog.Error("Cannot send request to: " + err.Error())
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		rlog.Errorf("%v Received hvacLogin status code %v, body %v", IP, resp.StatusCode, string(body))
		return "", NewError("Incorrect Status code")
	}

	auth := core.HvacAuth{}
	err = json.Unmarshal(body, &auth)
	if err != nil {
		rlog.Error("Cannot parse body: " + err.Error())
		return "", err
	}
	return auth.AccessToken, nil
}

func (s *Service) getHvacSetpoints(IP string, token string) (*core.HvacSetPointsValues, error) {
	url := "https://" + IP + "/api/setup/hvac/setpoint/loop1"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+token)
	req.Close = true
	transCfg := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		rlog.Error(err.Error())
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		rlog.Errorf("%v Received getHvacSetpoints status code %v, body %v", IP, resp.StatusCode, string(body))
		return nil, NewError("Incorrect Status code")
	}

	status := core.HvacSetPointsValues{}
	err = json.Unmarshal(body, &status)
	if err != nil {
		rlog.Error("Cannot parse body: " + err.Error())
		return nil, err
	}
	return &status, nil
}

func (s *Service) setHvacRuntime(conf dhvac.HvacConf, status dhvac.Hvac, IP string, token string) error {
	loopHvac := false
	maintenance, _ := s.getHvacMaintenanceMode(status.IP, token)
	if maintenance != nil {
		loopHvac = maintenance.Running
	}
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
	if conf.Hygrometry != nil {
		if param.AirRegister == nil {
			airRegister := core.HvacAirRegisterCtrl{}
			param.AirRegister = &airRegister
		}
		hygro := *conf.Hygrometry / 10
		param.AirRegister.SpaceHygroRel = &hygro
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

	if conf.TargetMode != nil {
		// 0 = OCCUPANCY_AUTO
		// 1 = OCCUPANCY_COMFORT
		// 2 = OCCUPANCY_STANDBY
		// 3 = OCCUPANCY_ECONOMY
		// 4 = OCCUPANCY_BUILDING_PROTECTION

		if param.Regulation == nil {
			airReg := core.HvacRegulationCtrl{}
			param.Regulation = &airReg
		}
		param.Regulation.OccManCmd = conf.TargetMode
	}

	if conf.HeatCool != nil {
		// 0 = HC_MODE_AUTO
		// 1 = HC_MODE_HEAT
		// 3 = HC_MODE_COOL
		// 6 = HC_MODE_OFF
		// 7 = HC_MODE_TEST
		// 8 = HC_MODE_EMERGENCY_HEAT

		if *conf.HeatCool != dhvac.HVAC_MODE_TEST {
			if loopHvac == true {
				if param.Regulation == nil {
					airReg := core.HvacRegulationCtrl{}
					param.Regulation = &airReg
				}
				param.Regulation.HeatCool = conf.HeatCool
			} else {
				_, err := s.setHvacMaintenanceBackMode(status.IP, token)
				if err != nil {
					rlog.Error("Cannot leave test mode", err.Error())
					return err
				}
				rlog.Info("HVAC leave test mode", status.Mac)
			}
		} else {
			rlog.Info("HVAC enter in test mode", status.Mac)
			err := s.setHvacMaintenanceMode(conf, status, status.IP, token)
			if err != nil {
				rlog.Error("Cannot switch in test mode", err)
				return err
			}
			err = s.setHvacMaintenanceParam(conf, status, status.IP, token)
			if err != nil {
				rlog.Error("Cannot prepare test mode", err)
				return err
			}
		}
		rlog.Infof("Switch HVAC " + status.Mac + ": in " + strconv.Itoa(*conf.HeatCool))
	}

	if (conf.HeatCool == nil || *conf.HeatCool == dhvac.HVAC_MODE_TEST) && loopHvac == false {
		err := s.setHvacMaintenanceParam(conf, status, status.IP, token)
		if err != nil {
			rlog.Error("Cannot send in test mode parameters ", err)
			return err
		}
	}

	if conf.Presence != nil {
		if status.OccManCmd1 == dhvac.OCCUPANCY_STANDBY ||
			status.OccManCmd1 == dhvac.OCCUPANCY_COMFORT ||
			((conf.TargetMode != nil) && ((*conf.TargetMode == dhvac.OCCUPANCY_STANDBY) || (*conf.TargetMode == dhvac.OCCUPANCY_COMFORT))) {
			//The presence is only takes into account in Standby mode
			if param.Regulation == nil {
				airReg := core.HvacRegulationCtrl{}
				param.Regulation = &airReg
			}
			pres := *conf.Presence
			if pres {
				mode := dhvac.OCCUPANCY_COMFORT
				param.Regulation.OccManCmd = &mode
			} else {
				mode := dhvac.OCCUPANCY_STANDBY
				param.Regulation.OccManCmd = &mode
			}
		}
	}

	if conf.ForcingAutoBack != nil {
		if *conf.ForcingAutoBack == 1 {
			s.setHvacMaintenanceBackMode(status.IP, token)
			rlog.Info("HVAC leave test mode", status.Mac)
		}
	}

	if (core.HvacLoopCtrl{}) == param {
		rlog.Infof("No new setHvacRuntime to set skip it %v: %v", status.Mac, param)
		return nil
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
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		rlog.Error(err.Error())
		return err
	}
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)

		rlog.Errorf("%v Received setHvacRuntime status code %v, body %v", status.Mac, resp.StatusCode, string(body))
		return NewError("Incorrect Status code")
	}

	return nil
}

func (s *Service) setHvacMaintenanceParam(conf dhvac.HvacConf, status dhvac.Hvac, IP string, token string) error {
	//Prepare Maintenance Outputs
	urlTask := "https://" + IP + "/api/maintenance/outputs"

	requestBodyTask := `{"output_Ya": 100, "output_Yb": 100`
	if conf.Forcing6waysValve != nil {
		requestBodyTask += `, "output_Y5": ` + strconv.Itoa(*conf.Forcing6waysValve)
	}
	if conf.ForcingDamper != nil {
		requestBodyTask += `, "output_Y6": ` + strconv.Itoa(*conf.ForcingDamper)
	}
	requestBodyTask += `}`
	rlog.Infof("Send HVAC test Mode parameters " + status.Mac + " : " + string(requestBodyTask))

	req, _ := http.NewRequest("POST", urlTask, strings.NewReader(requestBodyTask))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+token)
	req.Close = true
	transCfg := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		rlog.Error(err.Error())
		return err
	}
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)

		rlog.Errorf("%v Received setHvacMaintenanceParam status code %v, body %v", conf.Mac, resp.StatusCode, string(body))
		return NewError("Incorrect Status code")
	}
	return nil
}

func (s *Service) setHvacMaintenanceMode(conf dhvac.HvacConf, status dhvac.Hvac, IP string, token string) error {
	urlTask := "https://" + IP + "/api/maintenance/hvacTaskStatus"

	body := strings.NewReader(`{"running": false}`)

	rlog.Infof("Send new parameters to HVAC %v: %v", status.Mac, `{"running": false}`)
	req, _ := http.NewRequest("POST", urlTask, body)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+token)
	req.Close = true
	transCfg := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		rlog.Error(err.Error())
		return err
	}
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)

		rlog.Errorf("%v Received setHvacMaintenanceMode status code %v, body %v", conf.Mac, resp.StatusCode, string(body))
		return NewError("Incorrect Status code")
	}
	return nil
}

func (s *Service) setHvacMaintenanceBackMode(IP string, token string) (*core.HvacTask, error) {
	url := "https://" + IP + "/api/reboot"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+token)
	req.Close = true
	transCfg := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		rlog.Error(err.Error())
		return nil, err
	}
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)

		rlog.Errorf("%v Received setHvacMaintenanceBackMode status code %v, body %v", IP, resp.StatusCode, string(body))
		return nil, NewError("Incorrect Status code")
	}

	body, err := ioutil.ReadAll(resp.Body)

	status := core.HvacTask{}
	err = json.Unmarshal(body, &status)
	if err != nil {
		rlog.Error("Cannot parse body: " + err.Error())
		return nil, err
	}
	return &status, nil
}

func (s *Service) getHvacMaintenanceMode(IP string, token string) (*core.HvacTask, error) {
	url := "https://" + IP + "/api/maintenance/hvacTaskStatus"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+token)
	req.Close = true
	transCfg := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		rlog.Error(err.Error())
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		rlog.Errorf("%v Received getHvacMaintenanceMode status code %v, body %v", IP, resp.StatusCode, string(body))
		return nil, NewError("Incorrect Status code")
	}

	status := core.HvacTask{}
	err = json.Unmarshal(body, &status)
	if err != nil {
		rlog.Error("Cannot parse body: " + err.Error())
		return nil, err
	}
	return &status, nil
}

func (s *Service) setHvacSetupAirRegister(setup dhvac.HvacSetup, IP string, token string) error {
	url := "https://" + IP + "/api/setup/hvac/airRegister"

	hygroMode := 1
	config := core.HvacSetupAirQualityCtrl{
		HygroMode: &hygroMode,
	}
	if setup.OaDamperMode != nil {
		config.OADamperMode = setup.OaDamperMode
	}
	if setup.CO2Mode != nil {
		config.CO2Mode = setup.CO2Mode
	}
	if setup.CO2Max != nil {
		config.CO2Max = setup.CO2Max
	}

	requestBody, err := json.Marshal(config)
	if err != nil {
		return err
	}

	if (core.HvacSetupAirQualityCtrl{}) == config {
		rlog.Infof("No new setHvacSetupAirRegister to set skip it %v: %v", setup.Mac, config)
		return nil
	}
	rlog.Infof("Send HVAC AirRegister parameters " + setup.Mac + " : " + string(requestBody))

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+token)
	req.Close = true
	transCfg := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		rlog.Error(err.Error())
		return err
	}
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)

		rlog.Errorf("%v Received setHvacSetupAirRegister status code %v, body %v", setup.Mac, resp.StatusCode, string(body))
		return NewError("Incorrect Status code")
	}

	return nil
}

func (s *Service) setHvacSetupRegulation(setup dhvac.HvacSetup, IP string, token string) error {
	url := "https://" + IP + "/api/setup/hvac/regulation"

	if setup.TemperatureOffsetStep == nil {
		return nil
	}

	offset := float32(*setup.TemperatureOffsetStep) / 10.0

	config := core.HvacSetupRegulationCtrl{
		TemperOffsetStep:  &offset,
		TemperatureSelect: setup.TemperatureSelection,
		RegulType:         setup.RegulationType,
		LoopsUsed:         setup.LoopUsed,
	}

	if (core.HvacSetupRegulationCtrl{}) == config {
		rlog.Infof("No new HvacSetupRegulationCtrl to set skip it %v: %v", setup.Mac, config)
		return nil
	}
	requestBody, err := json.Marshal(config)
	if err != nil {
		return err
	}

	rlog.Infof("Send HVAC parameters " + setup.Mac + " : " + string(requestBody))

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+token)
	req.Close = true
	transCfg := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		rlog.Error(err.Error())
		return err
	}
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)

		rlog.Errorf("%v Received setHvacSetupRegulation status code %v, body %v", setup.Mac, resp.StatusCode, string(body))
		return NewError("Incorrect Status code")
	}

	return nil
}

func (s *Service) setHvacSetupInputs(setup dhvac.HvacSetup, IP string, token string) error {
	url := "https://" + IP + "/api/setup/inputs"

	config := core.HvacInput{}

	if setup.InputE1 != nil {
		config.InputE1 = setup.InputE1
	}
	if setup.InputE2 != nil {
		config.InputE2 = setup.InputE2
	}
	if setup.InputE3 != nil {
		config.InputE3 = setup.InputE3
	}
	if setup.InputE4 != nil {
		config.InputE4 = setup.InputE4
	}
	if setup.InputE5 != nil {
		config.InputE5 = setup.InputE5
	}
	if setup.InputE6 != nil {
		config.InputE6 = setup.InputE6
	}
	if setup.InputC1 != nil {
		config.InputC1 = setup.InputC1
	}
	if setup.InputC2 != nil {
		config.InputC2 = setup.InputC2
	}

	if (core.HvacInput{}) == config {
		rlog.Infof("No new setHvacSetupInputs to set skip it %v: %v", setup.Mac, config)
		return nil
	}

	requestBody, err := json.Marshal(config)
	if err != nil {
		return err
	}

	rlog.Infof("Send HVAC parameters " + setup.Mac + " : " + string(requestBody))

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+token)
	req.Close = true
	transCfg := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		rlog.Error(err.Error())
		return err
	}
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)

		rlog.Errorf("%v Received setHvacSetupInputs status code %v, body %v", setup.Mac, resp.StatusCode, string(body))
		return NewError("Incorrect Status code")
	}

	return nil
}

func (s *Service) setHvacSetupOutputs(setup dhvac.HvacSetup, IP string, token string) error {
	url := "https://" + IP + "/api/setup/outputs"

	config := core.HvacOutput{}

	if setup.OutputY5 != nil {
		config.OutputY5 = setup.OutputY5
	}
	if setup.OutputY6 != nil {
		config.OutputY6 = setup.OutputY6
	}
	if setup.OutputY7 != nil {
		config.OutputY7 = setup.OutputY7
	}
	if setup.OutputY8 != nil {
		config.OutputY8 = setup.OutputY8
	}
	if setup.OutputYa != nil {
		config.OutputYa = setup.OutputYa
	}
	if setup.OutputYb != nil {
		config.OutputYb = setup.OutputYb
	}

	if (core.HvacOutput{}) == config {
		rlog.Infof("No new HvacOutput to set skip it %v: %v", setup.Mac, config)
		return nil
	}

	requestBody, err := json.Marshal(config)
	if err != nil {
		return err
	}

	rlog.Infof("Send HVAC parameters " + setup.Mac + " : " + string(requestBody))

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+token)
	req.Close = true
	transCfg := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		rlog.Error(err.Error())
		return err
	}
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)

		rlog.Errorf("%v Received setHvacSetupInputs status code %v, body %v", setup.Mac, resp.StatusCode, string(body))
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
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		rlog.Error(err.Error())
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		rlog.Errorf("%v Received getHvacSetupRegulation status code %v, body %v", IP, resp.StatusCode, string(body))
		return nil, NewError("Incorrect Status code")
	}

	status := core.HvacSetupRegulation{}
	err = json.Unmarshal(body, &status)
	if err != nil {
		rlog.Error("Cannot parse body: " + err.Error())
		return nil, err
	}
	return &status, nil
}

func (s *Service) getHvacSetupInputs(IP string, token string) (*core.HvacInputValues, error) {
	url := "https://" + IP + "/api/setup/inputs"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+token)
	req.Close = true
	transCfg := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		rlog.Error(err.Error())
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		rlog.Errorf("%v Received getHvacSetupInputs status code %v, body %v", IP, resp.StatusCode, string(body))
		return nil, NewError("Incorrect Status code")
	}

	status := core.HvacInputValues{}
	err = json.Unmarshal(body, &status)
	if err != nil {
		rlog.Error("Cannot parse body: " + err.Error())
		return nil, err
	}
	return &status, nil
}

func (s *Service) getHvacMaintenanceOutputs(IP string, token string) (*core.HvacOutputValues, error) {
	url := "https://" + IP + "/api/maintenance/outputs"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+token)
	req.Close = true
	transCfg := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		rlog.Error(err.Error())
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		rlog.Errorf("%v Received getHvacMaintenanceOutputs status code %v, body %v", IP, resp.StatusCode, string(body))
		return nil, NewError("Incorrect Status code")
	}

	status := core.HvacOutputValues{}
	err = json.Unmarshal(body, &status)
	if err != nil {
		rlog.Error("Cannot parse body: " + err.Error())
		return nil, err
	}
	return &status, nil
}

func (s *Service) getHvacSetupOutputs(IP string, token string) (*core.HvacOutputValues, error) {
	url := "https://" + IP + "/api/setup/outputs"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+token)
	req.Close = true
	transCfg := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		rlog.Error(err.Error())
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		rlog.Errorf("%v Received getHvacSetupOutputs status code %v, body %v", IP, resp.StatusCode, string(body))
		return nil, NewError("Incorrect Status code")
	}

	status := core.HvacOutputValues{}
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
		SetpointOccCool:    &OccCool,
		SetpointOccHeat:    &OccHeat,
		SetpointUnoccHeat:  &UnoccHeat,
		SetpointUnoccCool:  &UnoccCool,
		SetpointStanbyCool: &CoolStandby,
		SetpointStanbyHeat: &HeatStandby,
	}

	if (core.HvacSetPoints{}) == config {
		rlog.Infof("No new Setpoint in HvacInit to set skip it %v: %v", setup.Mac, config)
		return nil
	}

	requestBody, err := json.Marshal(config)
	if err != nil {
		return err
	}

	rlog.Infof("Send HVAC parameters " + setup.Mac + " : " + string(requestBody))

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	req.Close = true
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+token)
	transCfg := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		resp.Body.Close()
	}
	if err != nil {
		rlog.Error(err.Error())
		return err
	}
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)

		rlog.Errorf("%v Received hvacInit status code %v, body %v", setup.Mac, resp.StatusCode, string(body))
		return NewError("Incorrect Status code")
	}
	return nil
}

func (s *Service) hvacSetAFConfig(setup dhvac.HvacConf, IP string, token string) error {
	url := "https://" + IP + "/api/setup/hvac/setpoint/loop1"

	config := core.HvacSetPoints{}
	if setup.SetpointCoolOccupied != nil {
		value := float32(*setup.SetpointCoolOccupied) / 10
		config.SetpointOccCool = &value
	}
	if setup.SetpointHeatOccupied != nil {
		value := float32(*setup.SetpointHeatOccupied) / 10
		config.SetpointOccHeat = &value
	}
	if setup.SetpointHeatInoccupied != nil {
		value := float32(*setup.SetpointHeatInoccupied) / 10
		config.SetpointUnoccHeat = &value
	}
	if setup.SetpointCoolInoccupied != nil {
		value := float32(*setup.SetpointCoolInoccupied) / 10
		config.SetpointUnoccCool = &value
	}
	if setup.SetpointCoolStandby != nil {
		value := float32(*setup.SetpointCoolStandby) / 10
		config.SetpointStanbyCool = &value
	}
	if setup.SetpointHeatStandby != nil {
		value := float32(*setup.SetpointHeatStandby) / 10
		config.SetpointStanbyHeat = &value
	}

	if (core.HvacSetPoints{}) == config {
		rlog.Infof("No new Setpoint to set skip it %v : %v", setup.Mac, config)
		return nil
	}

	requestBody, err := json.Marshal(config)
	if err != nil {
		return err
	}
	rlog.Infof("Send HVAC parameters " + setup.Mac + " : " + string(requestBody))

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	req.Close = true
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("authorization", "Bearer "+token)
	transCfg := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if resp != nil {
		resp.Body.Close()
	}

	if err != nil {
		rlog.Error(err.Error())
		return err
	}

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		rlog.Errorf("%v Received hvacSetAFConfig status code %v, body %v", setup.Mac, resp.StatusCode, string(body))

		return NewError("Incorrect Status code: " + strconv.Itoa((resp.StatusCode)) + " : " + string(body))
	}
	return nil
}
