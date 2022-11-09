package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

var count int = 0

func getData(w http.ResponseWriter, req *http.Request) {
	reader := csv.NewReader(req.Body)
	var a int = 0

	file, err := os.Create("data/same_dummy_name.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("File created successfully")

	for {
		// read one row from csv
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		//if err != nil {
		//	return nil, err
		//}

		// add record to result set
		len, err := file.WriteString(strings.Join(record, ""))
		file.WriteString("\n")

		if err != nil {
			fmt.Println("File created successfully", len)

			log.Fatalf("failed writing to file: %s", err)
			break
		}
		a++
	}

	count = count + a

	first := strconv.FormatInt(int64(count), 10)
	second := strconv.FormatInt(int64(count+a), 10)
	e := os.Rename("data/same_dummy_name.txt", "data/"+first+"_"+second+".txt")
	if e != nil {
		log.Fatal(e)
	}

	defer file.Close()

	//fmt.Printf("records count is", a)

	//err := os.WriteFile("", d1, 0644)

	//return results, nil
	/*
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}
	*/

	//fmt.Printf("%s", string(b))
	/*
	   fmt.Printf("got / request\n")
	   var path = html.EscapeString(r.URL.Path)
	   endPointPaths := strings.Split(path, "/")

	   	if len(endPointPaths) == 3 {
	   		if endPointPaths[1] == "promotions" {
	   			io.WriteString(w, "found "+endPointPaths[2])
	   		} else {
	   			io.WriteString(w, "not found ")
	   		}
	   	}
	*/

}

func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got / request\n")
	var path = html.EscapeString(r.URL.Path)
	endPointPaths := strings.Split(path, "/")

	if len(endPointPaths) == 3 {
		if endPointPaths[1] == "promotions" {
			io.WriteString(w, "found "+endPointPaths[2])
		} else {
			io.WriteString(w, "not found ")
		}
	}
}

func getHello(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		w.Write([]byte("Received a GET request\n"))
	case "POST":
		w.Write([]byte("Received a POST request\n"))
	default:
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
	}
	fmt.Printf("got /hello request\n")
	io.WriteString(w, "Hello, HTTP!\n")
}

func main() {
	//openFile("promotions.csv")

	http.HandleFunc("/", getRoot)
	http.HandleFunc("/hello", getHello)
	http.HandleFunc("/data", getData)

	err := http.ListenAndServe(":3333", nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}

}
