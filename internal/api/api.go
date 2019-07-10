package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	pkg "github.com/energieip/common-components-go/pkg/service"
	"github.com/energieip/swh200-rest2mqtt-go/internal/core"
	"github.com/gorilla/mux"
)

func (api *API) getAPIs(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	versions := []string{"v1.0"}
	apiInfo := APIInfo{
		Versions: versions,
	}
	inrec, _ := json.MarshalIndent(apiInfo, "", "  ")
	w.Write(inrec)
}

//InitAPI start API connection
func InitAPI(conf pkg.ServiceConfig) *API {
	api := API{
		EventsToBackend: make(chan map[string]interface{}),
		certificate:     conf.InternalAPI.CertPath,
		keyfile:         conf.InternalAPI.KeyPath,
		apiIP:           conf.InternalAPI.IP,
		apiPassword:     conf.InternalAPI.Password,
		apiPort:         conf.InternalAPI.Port,
		browsingFolder:  conf.InternalAPI.BrowsingFolder,
	}
	go api.swagger()
	return &api
}

func (api *API) setDefaultHeader(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
}

func (api *API) getFunctions(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	functions := []string{"/versions"}
	apiInfo := APIFunctions{
		Functions: functions,
	}
	inrec, _ := json.MarshalIndent(apiInfo, "", "  ")
	w.Write(inrec)
}

func (api *API) getV1Functions(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	apiV1 := "/v1.0"
	functions := []string{apiV1 + "/device/new"}
	apiInfo := APIFunctions{
		Functions: functions,
	}
	inrec, _ := json.MarshalIndent(apiInfo, "", "  ")
	w.Write(inrec)
}

func (api *API) newDevice(w http.ResponseWriter, req *http.Request) {
	api.setDefaultHeader(w)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Error reading request body", http.StatusInternalServerError)
		return
	}

	dr := core.Device{}
	err = json.Unmarshal(body, &dr)
	if err != nil {
		api.sendError(w, APIErrorBodyParsing, "Could not parse input format "+err.Error(), http.StatusInternalServerError)
		return
	}
	event := make(map[string]interface{})
	event["newDevice"] = dr
	api.EventsToBackend <- event
	w.Write([]byte("{}"))
}

func (api *API) swagger() {
	router := mux.NewRouter()
	sh := http.StripPrefix("/swaggerui/", http.FileServer(http.Dir("/media/userdata/www/swaggerui/")))
	router.PathPrefix("/swaggerui/").Handler(sh)

	// API v1.0
	apiV1 := "/v1.0"
	router.HandleFunc(apiV1+"/functions", api.getV1Functions).Methods("GET")

	//status
	router.HandleFunc(apiV1+"/driver/new", api.newDevice).Methods("POST")

	//unversionned API
	router.HandleFunc("/versions", api.getAPIs).Methods("GET")
	router.HandleFunc("/functions", api.getFunctions).Methods("GET")

	if api.browsingFolder != "" {
		sh2 := http.StripPrefix("/", http.FileServer(http.Dir(api.browsingFolder)))
		router.PathPrefix("/").Handler(sh2)
	}

	log.Fatal(http.ListenAndServeTLS(api.apiIP+":"+api.apiPort, api.certificate, api.keyfile, router))
}
