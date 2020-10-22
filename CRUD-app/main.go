package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

const GETALL string = "GETALL"
const GETONE string = "GETONE"
const POST string = "POST"
const PUT string = "PUT"
const DELETE string = "DELETE"

type clichePair struct {
	Id      int
	Cliche  string
	Counter string
}

type crudRequest struct {
	verb    string
	cp      *clichePair
	id      int
	cliche  string
	counter string
	confirm chan string
}

var clichesList = []*clichePair{}
var masterId = 1
var crudRequests chan *crudRequest

func ClichesAll(res http.ResponseWriter, req *http.Request) {
	cr := &crudRequest{verb: GETALL, confirm: make(chan string)}
	completeRequest(cr, res, "read all")
}

func ClichesOne(res http.ResponseWriter, req *http.Request) {
	id := getIdFromRequest(req)
	cr := &crudRequest{verb: GETONE, id: id, confirm: make(chan string)}
	completeRequest(cr, res, "read one")
}

func ClichesCreate(res http.ResponseWriter, req *http.Request) {
	cliche, counter := getDataFromRequest(req)
	cp := new(clichePair)
	cp.Cliche = cliche
	cp.Counter = counter
	cr := &crudRequest{verb: POST, cp: cp, confirm: make(chan string)}
	completeRequest(cr, res, "create")
}

func ClichesEdit(res http.ResponseWriter, req *http.Request) {
	id := getIdFromRequest(req)
	cliche, counter := getDataFromRequest(req)
	cr := &crudRequest{verb: PUT, id: id, cliche: cliche, counter: counter,
		confirm: make(chan string)}
	completeRequest(cr, res, "edit")
}

func ClichesDelete(res http.ResponseWriter, req *http.Request) {
	id := getIdFromRequest(req)
	cr := &crudRequest{verb: DELETE, id: id, confirm: make(chan string)}
	completeRequest(cr, res, "delete")
}

func completeRequest(cr *crudRequest, res http.ResponseWriter, logMsg string) {
	crudRequests <- cr
	msg := <-cr.confirm
	res.Write([]byte(msg))
	logIt(logMsg)
}

func main() {
	populateClichesList()

	crudRequests = make(chan *crudRequest, 8)
	go func() {
		for {
			select {
			case req := <-crudRequests:
				if req.verb == GETALL {
					req.confirm <- readAll()
				} else if req.verb == GETONE {
					req.confirm <- readOne(req.id)
				} else if req.verb == POST {
					req.confirm <- addPair(req.cp)
				} else if req.verb == PUT {
					req.confirm <- editPair(req.id, req.cliche, req.counter)
				} else if req.verb == DELETE {
					req.confirm <- deletePair(req.id)
				}
			}
		}
	}()

	startServer()
}

func startServer() {
	router := mux.NewRouter()

	router.HandleFunc("/", ClichesAll).Methods("GET")
	router.HandleFunc("/cliches", ClichesAll).Methods("GET")
	router.HandleFunc("/cliches/{id:[0-9]+}", ClichesOne).Methods("GET")

	router.HandleFunc("/cliches", ClichesCreate).Methods("POST")
	router.HandleFunc("/cliches/{id:[0-9]+}", ClichesEdit).Methods("PUT")
	router.HandleFunc("/cliches/{id:[0-9]+}", ClichesDelete).Methods("DELETE")

	http.Handle("/", router)

	port := ":8888"
	fmt.Println("\nListening on port " + port)
	http.ListenAndServe(port, router)
}

func readAll() string {
	msg := "\n"
	for _, cliche := range clichesList {
		next := strconv.Itoa(cliche.Id) + ": " + cliche.Cliche + "  " + cliche.Counter + "\n"
		msg += next
	}
	return msg
}

func readOne(id int) string {
	msg := "\n" + "Bad Id: " + strconv.Itoa(id) + "\n"

	index := findCliche(id)
	if index >= 0 {
		cliche := clichesList[index]
		msg = "\n" + strconv.Itoa(id) + ": " + cliche.Cliche + "  " + cliche.Counter + "\n"
	}
	return msg
}

func addPair(cp *clichePair) string {
	cp.Id = masterId
	masterId++
	clichesList = append(clichesList, cp)
	return "\nCreated: " + cp.Cliche + " " + cp.Counter + "\n"
}

func editPair(id int, cliche string, counter string) string {
	msg := "\n" + "Bad Id: " + strconv.Itoa(id) + "\n"
	index := findCliche(id)
	if index >= 0 {
		clichesList[index].Cliche = cliche
		clichesList[index].Counter = counter
		msg = "\nCliche edited: " + cliche + " " + counter + "\n"
	}
	return msg
}

func deletePair(id int) string {
	idStr := strconv.Itoa(id)
	msg := "\n" + "Bad Id: " + idStr + "\n"
	index := findCliche(id)
	if index >= 0 {
		clichesList = append(clichesList[:index], clichesList[index+1:]...)
		msg = "\nCliche " + idStr + " deleted\n"
	}
	return msg
}

func findCliche(id int) int {
	for i := 0; i < len(clichesList); i++ {
		if id == clichesList[i].Id {
			return i
		}
	}
	return -1
}

func getIdFromRequest(req *http.Request) int {
	vars := mux.Vars(req)
	id, _ := strconv.Atoi(vars["id"])
	return id
}

func getDataFromRequest(req *http.Request) (string, string) {
	req.ParseForm()
	form := req.Form
	cliche := form["cliche"][0]
	counter := form["counter"][0]
	return cliche, counter
}

func logIt(msg string) {
	fmt.Println(msg)
}

func populateClichesList() {
	var cliches = []string{
		"Out of sight, out of mind.",
		"A penny saved is a penny earned.",
		"He who hesitates is lost.",
	}
	var counterCliches = []string{
		"Absence makes the heart grow fonder.",
		"Penny-wise and dollar-foolish.",
		"Look before you leap.",
	}

	for i := 0; i < len(cliches); i++ {
		cp := new(clichePair)
		cp.Id = masterId
		masterId++
		cp.Cliche = cliches[i]
		cp.Counter = counterCliches[i]
		clichesList = append(clichesList, cp)
	}
}
