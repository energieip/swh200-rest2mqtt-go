package core

//HvacHello network object
type HvacHello struct {
	Mac             string `json:"mac"`
	FullMac         string `json:"fullMac"`
	IP              string `json:"ip"`
	Protocol        string `json:"protocol"`
	SwitchMac       string `json:"switchMac"`
	IsConfigured    bool   `json:"isConfigured"`
	DumpFrequency   int    `json:"dumpFrequency"`
	FriendlyName    string `json:"friendlyName"`
	SoftwareVersion string `json:"softwareVersion"`
	Error           int    `json:"error"`
}

type HvacLogin struct {
	UserKey string `json:"userKey"`
}

type HvacAuth struct {
	TokenType   string `json:"tokenType"`
	AccessToken string `json:"token"`
	ExpireIn    int    `json:"expireIn"`
	Admin       bool   `json:"admin"`
}

type HvacSysInfo struct {
	Mac                 string `json:"macAddress"`
	FactoryVersion      string `json:"factoryVersion"`
	MainAppfirmwVersion string `json:"mainAppfirmwVersion"`
	RestAPIVersion      string `json:"restApiVersion"`
	DatabaseVersion     string `json:"databaseVersion"`
}

type HvacLoop1 struct {
	Regulation  HvacRegulation  `json:"regulation"`
	Ventilation HvacVentilation `json:"ventilation"`
	AirRegister HvacAirRegister `json:"airRegister"`
}

type HvacRegulation struct {
	WindowHoldOff       int `json:"windowHoldOff"`
	WindowHeartBeat     int `json:"windowHeartBeat"`
	SpaceTemp           int `json:"spaceTemp"`
	OffsetTemp          int `json:"offsetTemp"`
	OccManCmd           int `json:"occManCmd"`
	HeatCool            int `json:"heatCool"`
	EffectifSetPoint    int `json:"effectifSetPoint"`
	HeatOuput           int `json:"heatOuput"`
	CoolOuput           int `json:"coolOuput"`
	HeatOutputSecondary int `json:"heatOutputSecondary"`
	DewSensor           int `json:"dewSensor"`
	ChangeOver          int `json:"changeOver"`
	DischAirTemp        int `json:"dischAirTemp"`
}

type HvacVentilation struct {
	FanSpeed         int `json:"fanSpeed"`
	FanSpeedCmdValue int `json:"fanSpeedCmdValue"`
	FanSpeedCmdMode  int `json:"fanSpeedCmdMode"`
}

type HvacAirRegister struct {
	SpaceCO2 int `json:"spaceCO2"`
	OADamper int `json:"OADamper"`
}

type HvacSetPoints struct {
	SetpointOccCool    float32 `json:"setpointOccCool"`
	SetpointOccHeat    float32 `json:"setpointOccHeat"`
	SetpointUnoccCool  float32 `json:"setpointUnoccCool"`
	SetpointUnoccHeat  float32 `json:"setpointUnoccHeat"`
	SetpointStanbyCool float32 `json:"setpointStanbyCool"`
	SetpointStanbyHeat float32 `json:"setpointStanbyHeat"`
}
