package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"time"
	"io"
	"log"
)

func main() {

	// check current parent folder
	dir, _ := os.Getwd()
	fmt.Println("Current dir: " + dir);

	fmt.Println("-- Preparing and moving api.json and responses files ...")

	if os.Args == nil || len(os.Args) == 1 {
		fmt.Println("ERROR: You must provide path to api.json")
		os.Exit(1)
	}
	apiPath := os.Args[1]

	// check for relative path
	fmt.Printf("args: %v", os.Args)

	relativePath := ""
	if len(os.Args) == 3 {
		relativePath = os.Args[2]
	}

	var docTemplate = template.Must(template.New("doc").ParseFiles(relativePath + "theme/index.html"))

	// clean input dir
	os.RemoveAll(relativePath + "input")
	os.Mkdir(relativePath + "input", os.ModePerm)

	// copy api.json
	CopyFile(apiPath + "/api.json", relativePath + "input/api.json")

	// copy responses if exist
	err := CopyDir(apiPath + "/responses", relativePath + "input/responses")
	if err != nil {
		log.Fatal(err)
	} else {
		log.Print("Files copied.")
	}

	fmt.Println("-- start api docs generation...")
	fmt.Println("-- read input: api.json")

	file, e := ioutil.ReadFile(relativePath + "input/api.json")
	if e != nil {
		fmt.Printf("ERROR: File error: %v\n", e)
		os.Exit(1)
	}
	var api API
	json.Unmarshal(file, &api)

	fmt.Println("-- load resources")
	for i, endpoint := range api.Endpoints {
		response, e := ioutil.ReadFile(relativePath + "input/responses/" + endpoint.Response)
		if e != nil {
			fmt.Printf("ERROR: File error: %v\n", e)
			os.Exit(1)
		}
		api.Endpoints[i].Response = string(response)
		api.Endpoints[i].Example = createCurlExample(api.Base.Url, &endpoint)

		if len(endpoint.Params) > 0 {
			api.Endpoints[i].HasParams = true
		} else {
			api.Endpoints[i].HasParams = false
		}

		if len(endpoint.UrlParams) > 0 {
			api.Endpoints[i].HasUrlParams = true
		} else {
			api.Endpoints[i].HasUrlParams = false
		}
	}

	fmt.Println("-- create doc")
	doc := &Doc{
		Title:       api.Title,
		GeneratedAt: time.Now().Format(time.RFC1123),
		Api:         api,
	}

	f, _ := os.Create(relativePath + "theme/doc.html")
	docTemplate.ExecuteTemplate(f, "doc", doc)

	fmt.Println("-- finish generation")
}

func createCurlExample(baseUrl string, endpoint *Endpoint) string {
	example := "" // todo
	return example
}


func CopyFile(source string, dest string) (err error) {
	sf, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sf.Close()
	df, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer df.Close()
	_, err = io.Copy(df, sf)
	if err == nil {
		si, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, si.Mode())
		}

	}

	return
}

// Recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
func CopyDir(source string, dest string) (err error) {

	// get properties of source dir
	fi, err := os.Stat(source)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return &CustomError{"Source is not a directory"}
	}

	// ensure dest dir does not already exist

	_, err = os.Open(dest)
	if !os.IsNotExist(err) {
		return &CustomError{"Destination already exists"}
	}

	// create dest dir

	err = os.MkdirAll(dest, fi.Mode())
	if err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(source)

	for _, entry := range entries {

		sfp := source + "/" + entry.Name()
		dfp := dest + "/" + entry.Name()
		if entry.IsDir() {
			err = CopyDir(sfp, dfp)
			if err != nil {
				log.Println(err)
			}
		} else {
			// perform copy
			err = CopyFile(sfp, dfp)
			if err != nil {
				log.Println(err)
			}
		}

	}
	return
}

// A struct for returning custom error messages
type CustomError struct {
	What string
}

// Returns the error message defined in What as a string
func (e *CustomError) Error() string {
	return e.What
}

type Doc struct {
	Title       string
	GeneratedAt string
	Api         API
}

type API struct {
	Title 	  string 	 `json:"title"`
	Base      Base       `json:"base"`
	Endpoints []Endpoint `json:"endpoints"`
}

type Base struct {
	Url     string   `json:"url"`
	Headers []Header `json:"headers"`
}

type Header struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Endpoint struct {
	Endpoint     string       `json:"endpoint"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	Method       string       `json:"method"`
	UrlParams    []Param      `json:"url-params"`
	Params       []Param      `json:"params"`
	Response     string       `json:"response"`
	ResultCodes  []ResultCode `json:"codes"`
	Example      string
	HasUrlParams bool
	HasParams    bool
}

type Param struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Value       string `json:"value"`
	Must        bool   `json:"must"`
	Default     string `json:"default"`
	Options     string `json:"options"`
}

type ResultCode struct {
	Code        int    `json:"code"`
	Type        string `json:"type"`
	Description string `json:"description"`
}
