package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/usmayoung/animoto_interview/troll_captcha/models"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
)

var tpl *template.Template

//captchaCache is a cache with the key as the unique TrollCaptcha id
//(calculated as the hash of the original text) used for contant time lookup (on average)
// of TrollCaptcha in memory
var captchaCache map[string]*models.TrollCaptcha

//captchaIndex is used to get random access, contant time lookup necessary for generating
//random TrollCaptcha to the client
var captchaIndex []*models.TrollCaptcha

//init runs once and before main()
func init() {
	//create the cache on the heap
	captchaCache = make(map[string]*models.TrollCaptcha)
	//parse template views
	tpl = template.Must(template.ParseGlob("templates/*"))
	//build TrollCaptcha from local text files
	readTextFiles()
}

func main() {
	//create router
	r := mux.NewRouter()

	//assign handlers for routes
	r.HandleFunc("/", index)
	r.HandleFunc("/troll_captchas/{id}", trollCaptcha)
	r.Handle("/favicon.ico", http.NotFoundHandler())

	//start server on port :8080
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}

}

//index is the splash page that renders index.gohtml, the form for submitting
//the captcha, a random captcha is slected each time the splash page is rendered
func index(w http.ResponseWriter, req *http.Request) {
	trollCaptcha := captchaIndex[rand.Intn(len(captchaIndex))]
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Println(trollCaptcha.Words)
	tpl.ExecuteTemplate(w, "index.gohtml", trollCaptcha)
}

//trollCaptcha is the handler for the submission of a trollcaptcha
//the url is parse "troll_captchas/:id, where the id is the unique identifier
//for the trollcaptcha (calculated as the hash of the original text)
func trollCaptcha(w http.ResponseWriter, req *http.Request) {

	path := mux.Vars(req)
	//get the unique trollcaptcha id from url
	captchaId := path["id"]
	var serverCaptcha *models.TrollCaptcha

	//Check if id is found in memory, and is valid, if not, you may have a troll
	if v, ok := captchaCache[captchaId]; ok {
		serverCaptcha = v
	} else {
		tpl.ExecuteTemplate(w, "failure.gohtml", "You're a troll")
	}

	//parse client form
	err := req.ParseForm()
	if err != nil {
		log.Println(err)
	}
	decoder := schema.NewDecoder()

	//decode client form and map to client captcha
	clientCaptcha := &models.TrollCaptcha{}
	err = decoder.Decode(clientCaptcha, req.PostForm)

	if err != nil {
		log.Println(err)
	}

	//Validate the server (valid captcha) with client captcha, and render appropriate view
	message, valid := serverCaptcha.ValidateClientCaptcha(clientCaptcha)
	if valid {
		tpl.ExecuteTemplate(w, "success.gohtml", message)
	} else {
		tpl.ExecuteTemplate(w, "failure.gohtml", message)
	}

}

//Process local text files, create TrollCaptcha for each file
//add each TrollCaptcha to cache
func readTextFiles() {
	files, err := ioutil.ReadDir("./texts")
	captchaIndex = make([]*models.TrollCaptcha, len(files))

	if err != nil {
		log.Fatal(err)
	}
	for i, f := range files {
		data, err := ioutil.ReadFile("./texts/" + f.Name())
		check(err)

		file_string := string(data)
		trollCap := models.NewTrollCaptcha(file_string, i)
		captchaCache[trollCap.Id] = trollCap
		captchaIndex[i] = trollCap

	}

}

//Helper method to check for errors, and responds with panic if error
func check(e error) {
	if e != nil {
		panic(e)
	}
}
