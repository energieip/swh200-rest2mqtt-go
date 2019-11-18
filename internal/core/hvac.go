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
	HeatOutput          int     `json:"heatOutput"`
	CoolOutput          int     `json:"coolOutput"`
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

type HvacSetPointsValues struct {
	SetpointOccCool    float32 `json:"setpointOccCool"`
	SetpointOccHeat    float32 `json:"setpointOccHeat"`
	SetpointUnoccCool  float32 `json:"setpointUnoccCool"`
	SetpointUnoccHeat  float32 `json:"setpointUnoccHeat"`
	SetpointStanbyCool float32 `json:"setpointStanbyCool"`
	SetpointStanbyHeat float32 `json:"setpointStanbyHeat"`
}
type HvacSetPoints struct {
	SetpointOccCool    *float32 `json:"setpointOccCool,omitempty"`
	SetpointOccHeat    *float32 `json:"setpointOccHeat,omitempty"`
	SetpointUnoccCool  *float32 `json:"setpointUnoccCool,omitempty"`
	SetpointUnoccHeat  *float32 `json:"setpointUnoccHeat,omitempty"`
	SetpointStanbyCool *float32 `json:"setpointStanbyCool,omitempty"`
	SetpointStanbyHeat *float32 `json:"setpointStanbyHeat,omitempty"`
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

type HvacSetupAirQualityCtrl struct {
	OADamperMode  *int     `json:"oaDamperMode,omitempty"`
	OADamperMin   *int     `json:"oaDamperMin,omitempty"`
	OADamperMax   *int     `json:"oaDamperMax,omitempty"`
	CO2Mode       *int     `json:"co2Mode,omitempty"`
	CO2Setpoint   *int     `json:"co2Setpoint,omitempty"`
	CO2Bp         *int     `json:"co2Bp,omitempty"`
	CO2Max        *int     `json:"co2Max,omitempty"`
	HygroMode     *int     `json:"hygroMode,omitempty"`
	HygroSetpoint *float32 `json:"hygroSetpoint,omitempty"`
	HygroBp       *float32 `json:"hygroBp,omitempty"`
}

type HvacInput struct {
	InputE1 *int `json:"inputE1,omitempty"`
	InputE2 *int `json:"inputE2,omitempty"`
	InputE3 *int `json:"inputE3,omitempty"`
	InputE4 *int `json:"inputE4,omitempty"`
	InputE5 *int `json:"inputE5,omitempty"`
	InputE6 *int `json:"inputE6,omitempty"`
	InputC1 *int `json:"inputC1,omitempty"`
	InputC2 *int `json:"inputC2,omitempty"`
}

type HvacInputValues struct {
	InputE1 int `json:"inputE1"`
	InputE2 int `json:"inputE2"`
	InputE3 int `json:"inputE3"`
	InputE4 int `json:"inputE4"`
	InputE5 int `json:"inputE5"`
	InputE6 int `json:"inputE6"`
	InputC1 int `json:"inputC1"`
	InputC2 int `json:"inputC2"`
}

type HvacOutput struct {
	OutputY5 *int `json:"outputY5,omitempty"`
	OutputY6 *int `json:"outputY6,omitempty"`
	OutputY7 *int `json:"outputY7,omitempty"`
	OutputY8 *int `json:"outputY8,omitempty"`
	OutputYa *int `json:"outputYa,omitempty"`
	OutputYb *int `json:"outputYb,omitempty"`
}

type HvacOutputValues struct {
	OutputY5 int `json:"outputY5"`
	OutputY6 int `json:"outputY6"`
	OutputY7 int `json:"outputY7"`
	OutputY8 int `json:"outputY8"`
	OutputYa int `json:"outputYa"`
	OutputYb int `json:"outputYb"`
}

type HvacUpdateParams struct {
	TftpServerIP string `json:"tftpServerIP"`
	StartUpdate  bool   `json:"startUpdate"`
}

type HvacTask struct {
	Running bool `json:"running"`
}
