package core

//HvacHello network object
type HvacHello struct {
	Mac             string `json:"mac"`
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
	AccessToken string `json:"accessToken"`
	ExpireIn    int    `json:"expireIn"`
	Admin       bool   `json:"admin"`
}

type HvacSysInfo struct {
	Mac               string `json:"macAddress"`
	ProductType       string `json:"productType"`
	FactoryVersion    string `json:"factoryVersion"`
	SoftwareVersion   string `json:"softwareVersion"`
	DatabaseVersion   string `json:"databaseVersion"`
	ParametersVersion string `json:"parametersVersion"`
}

type HvacLoop1 struct {
	Regulation  HvacRegulation  `json:"regulation"`
	Ventilation HvacVentilation `json:"ventilation"`
	AirRegister HvacAirRegister `json:"airRegister"`
}

type HvacRegulation struct {
	WindowHoldOff       int     `json:"windowHoldOff"`
	WindowHeartBeat     int     `json:"windowHeartBeat"`
	SpaceTemp           float32 `json:"spaceTemp"`
	OffsetTemp          int     `json:"offsetTemp"`
	OccManCmd           int     `json:"occManCmd"`
	HeatCool            int     `json:"heatCool"`
	EffectifSetPoint    float32 `json:"effectifSetPoint"`
	HeatOuput           int     `json:"heatOuput"`
	CoolOuput           int     `json:"coolOuput"`
	HeatOutputSecondary int     `json:"heatOutputSecondary"`
	DewSensor           int     `json:"dewSensor"`
	ChangeOver          int     `json:"changeOver"`
	DischAirTemp        float32 `json:"dischAirTemp"`
}

type HvacVentilation struct {
	FanSpeed         int `json:"fanSpeed"`
	FanSpeedCmdValue int `json:"fanSpeedCmdValue"`
	FanSpeedCmdMode  int `json:"fanSpeedCmdMode"`
}

type HvacAirRegister struct {
	SpaceCO2      int `json:"spaceCO2"`
	OADamper      int `json:"OADamper"`
	SpaceHygroRel int `json:"spaceHygroRel"`
}

type HvacLoopCtrl struct {
	Regulation  *HvacRegulationCtrl  `json:"regulation,omitempty"`
	Ventilation *HvacVentilationCtrl `json:"ventilation,omitempty"`
	AirRegister *HvacAirRegisterCtrl `json:"airRegister,omitempty"`
}

type HvacRegulationCtrl struct {
	WindowHoldOff       *int     `json:"windowHoldOff,omitempty"`
	WindowHeartBeat     *int     `json:"windowHeartBeat,omitempty"`
	SpaceTemp           *float32 `json:"spaceTemp,omitempty"`
	OffsetTemp          *int     `json:"offsetTemp,omitempty"`
	OccManCmd           *int     `json:"occManCmd,omitempty"`
	HeatCool            *int     `json:"heatCool,omitempty"`
	EffectifSetPoint    *float32 `json:"effectifSetPoint,omitempty"`
	HeatOuput           *int     `json:"heatOuput,omitempty"`
	CoolOuput           *int     `json:"coolOuput,omitempty"`
	HeatOutputSecondary *int     `json:"heatOutputSecondary,omitempty"`
	DewSensor           *int     `json:"dewSensor,omitempty"`
	ChangeOver          *int     `json:"changeOver,omitempty"`
	DischAirTemp        *float32 `json:"dischAirTemp,omitempty"`
}

type HvacVentilationCtrl struct {
	FanSpeed         *int `json:"fanSpeed,omitempty"`
	FanSpeedCmdValue *int `json:"fanSpeedCmdValue,omitempty"`
	FanSpeedCmdMode  *int `json:"fanSpeedCmdMode,omitempty"`
}

type HvacAirRegisterCtrl struct {
	SpaceCO2      *int `json:"spaceCO2,omitempty"`
	OADamper      *int `json:"OADamper,omitempty"`
	SpaceHygroRel *int `json:"spaceHygroRel,omitempty"`
}

type HvacSetPoints struct {
	SetpointOccCool    float32 `json:"setpointOccCool"`
	SetpointOccHeat    float32 `json:"setpointOccHeat"`
	SetpointUnoccCool  float32 `json:"setpointUnoccCool"`
	SetpointUnoccHeat  float32 `json:"setpointUnoccHeat"`
	SetpointStanbyCool float32 `json:"setpointStanbyCool"`
	SetpointStanbyHeat float32 `json:"setpointStanbyHeat"`
}

type HvacSetupRegulation struct {
	TemperatureSelect int     `json:"temperSelect"`
	OccResetOffset    int     `json:"occResetOffset"`
	TemperOffsetStep  float32 `json:"temperOffsetStep"`
	RegulType         int     `json:"regulType"`
	LoopsUsed         int     `json:"loopsUsed"`
	PropBandHeat      int     `json:"propBandHeat"`
	PropBandCool      int     `json:"propBandCool"`
	PropBandElec      int     `json:"propBandElec"`
	ResetTimeHeat     int     `json:"resetTimeHeat"`
	ResetTimeCool     int     `json:"resetTimeCool"`
	ResetTimeElec     int     `json:"resetTimeElec"`
}

type HvacSetupRegulationCtrl struct {
	TemperatureSelect *int     `json:"temperSelect,omitempty"`
	OccResetOffset    *int     `json:"occResetOffset,omitempty"`
	TemperOffsetStep  *float32 `json:"temperOffsetStep,omitempty"`
	RegulType         *int     `json:"regulType,omitempty"`
	LoopsUsed         *int     `json:"loopsUsed,omitempty"`
	PropBandHeat      *int     `json:"propBandHeat,omitempty"`
	PropBandCool      *int     `json:"propBandCool,omitempty"`
	PropBandElec      *int     `json:"propBandElec,omitempty"`
	ResetTimeHeat     *int     `json:"resetTimeHeat,omitempty"`
	ResetTimeCool     *int     `json:"resetTimeCool,omitempty"`
	ResetTimeElec     *int     `json:"resetTimeElec,omitempty"`
}
