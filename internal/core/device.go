package core

import "encoding/json"

//Device network object
type Device struct {
	IP  string `json:"ip"`
	Mac string `json:"mac"`
}

//ToDevice convert map interface to Device object
func ToDevice(val interface{}) (*Device, error) {
	var driver Device
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &driver)
	return &driver, err
}
