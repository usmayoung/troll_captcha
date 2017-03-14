package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"./models"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"encoding/json"
	"flag"
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
	tpl = template.Must(template.ParseGlob("./templates/*"))
	//build TrollCaptcha from local text files
}

var numberCaptcha int
func main() {
	//processing flags from command line, if no flags use local files as captcha text,
	//if captcha_numb is set, use that number to retrieve the correnct texts from api
	//listed below
	numbPtr := flag.Int("captcha_numb", 0, "integer flag for determining number of captcha to fetch")
	flag.Parse()
	numberCaptcha = *numbPtr

	if numberCaptcha == 0 {
		readTextFiles()
	} else {
		getChuckNorrisText()
	}

	//create router
	r := mux.NewRouter()

	//assign handlers for routes
	r.HandleFunc("/", index)
	r.HandleFunc("/troll_captchas/{id}", trollCaptcha)
	r.Handle("/favicon.ico", http.NotFoundHandler())

	//start server on port :8000
	err := http.ListenAndServe(":8000", r)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}

}

//index is the splash page that renders index.gohtml, the form for submitting
//the captcha, a random captcha is slected each time the splash page is rendered
func index(w http.ResponseWriter, req *http.Request) {
	trollCaptcha := captchaIndex[rand.Intn(len(captchaIndex))]
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
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

	fmt.Println("Server has processed all local files")

}

//getChuckNorrisText gets the number of texts specified in -captcha_numb flag
//all requests to api use goroutines and run in parallel to get jokes
func getChuckNorrisText() {
	n := numberCaptcha
	c := make(chan models.Message)
	done := make(chan bool)
	captchaIndex = make([]*models.TrollCaptcha, n)

	for i := 0; i < n; i++ {
		go func () {
			resp, _ :=http.Get("http://api.icndb.com/jokes/random/")
			var m models.Message
			if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
				log.Println(err)
			}

			c <- m
			done <- true
		}()

	}


	go func() {
		for i := 0; i < n; i++ {
			<-done
		}
		close(c)
	}()

	count := 0;
	for m := range c {
		trollCap := models.NewTrollCaptcha(m.Value.Joke, m.Id)
		captchaCache[trollCap.Id] = trollCap
		captchaIndex[count] = trollCap
		count++
	}
	fmt.Println("Server has processed all ", numberCaptcha, "Chuck Norris jokes")
}

//Helper method to check for errors, and responds with panic if error
func check(e error) {
	if e != nil {
		panic(e)
	}
}
