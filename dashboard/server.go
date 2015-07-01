package main

import (
	"encoding/json"
	"fmt"
	"github.com/haxpax/gosms"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

//reposne structure to /sms
type SMSResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	UUID string    `json:"uuid"`
}

//response structure to /smsdata/
type SMSDataResponse struct {
	Status   int            `json:"status"`
	Message  string         `json:"message"`
	Summary  []int          `json:"summary"`
	DayCount map[string]int `json:"daycount"`
	Messages []gosms.SMS    `json:"messages"`
}

type SMSMessageResponse struct {
	Status   int            `json:"status"`
	Message  string         `json:"message"`
	Messages []gosms.SMS    `json:"messages"`
}

SMSMessageResponse

// Cache templates
var templates = template.Must(template.ParseFiles("./templates/index.html"))

/* dashboard handlers */

// dashboard
func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("--- indexHandler")
	// templates.ExecuteTemplate(w, "index.html", nil)
	// Use during development to avoid having to restart server
	// after every change in HTML
	t, _ := template.ParseFiles("./templates/index.html")
	t.Execute(w, nil)
}

// handle all static files based on specified path
// for now its /assets
func handleStatic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	static := vars["path"]
	http.ServeFile(w, r, filepath.Join("./assets", static))
}

/* end dashboard handlers */

/* API handlers */

// push sms, allowed methods: POST
func sendSMSHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("--- sendSMSHandler")
	w.Header().Set("Content-type", "application/json")

	//TODO: validation
	r.ParseForm()
	mobile := r.FormValue("mobile")
	message := r.FormValue("message")
	uuid := uuid.NewV1()
	sms := &gosms.SMS{UUID: uuid.String(), Mobile: mobile, Body: message, Retries: 0}
	gosms.EnqueueMessage(sms, true)

	smsresp := SMSResponse{Status: 200, Message: "ok", UUID: uuid}
	var toWrite []byte
	toWrite, err := json.Marshal(smsresp)
	if err != nil {
		log.Println(err)
		//lets just depend on the server to raise 500
	}
	w.Write(toWrite)
}

// dumps JSON data, used by log view. Methods allowed: GET
func getLogsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("--- getLogsHandler")
	messages, _ := gosms.GetMessages("")
	summary, _ := gosms.GetStatusSummary()
	dayCount, _ := gosms.GetLast7DaysMessageCount()
	logs := SMSDataResponse{
		Status:   200,
		Message:  "ok",
		Summary:  summary,
		DayCount: dayCount,
		Messages: messages,
	}
	var toWrite []byte
	toWrite, err := json.Marshal(logs)
	if err != nil {
		log.Println(err)
		//lets just depend on the server to raise 500
	}
	w.Header().Set("Content-type", "application/json")
	w.Write(toWrite)
}

// dumps JSON data, used by log view. Methods allowed: GET
func getMessageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("--- getMessageHandler")
	r.ParseForm()
	uuid := r.FormValue("uuid")
	query := fmt.Sprintf("WHERE uuid=%v", uuid)

	messages, _ := gosms.GetMessages(query)
	logs := SMSMessageResponse{
		Status:   200,
		Message:  "ok",
		Messages: messages,
	}
	var toWrite []byte
	toWrite, err := json.Marshal(logs)
	if err != nil {
		log.Println(err)
		//lets just depend on the server to raise 500
	}
	w.Header().Set("Content-type", "application/json")
	w.Write(toWrite)
}

/* end API handlers */

func InitServer(host string, port string) error {
	log.Println("--- InitServer ", host, port)

	r := mux.NewRouter()
	r.StrictSlash(true)

	r.HandleFunc("/", indexHandler)

	// handle static files
	r.HandleFunc(`/assets/{path:[a-zA-Z0-9=\-\/\.\_]+}`, handleStatic)

	// all API handlers
	api := r.PathPrefix("/api").Subrouter()
	api.Methods("GET").Path("/logs/").HandlerFunc(getLogsHandler)
	api.Methods("POST").Path("/sms/").HandlerFunc(sendSMSHandler)
	api.Methods("GET").Path("/query/").HandlerFunc(getMessageHandler)

	http.Handle("/", r)

	bind := fmt.Sprintf("%s:%s", host, port)
	log.Println("listening on: ", bind)
	return http.ListenAndServe(bind, nil)

}
