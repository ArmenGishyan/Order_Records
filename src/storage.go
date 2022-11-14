package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// same utility functions
func makeString(prev []string, current []string) []string {
	if len(current) == 0 {
		return prev
	}

	prev[len(prev)-1] = prev[len(prev)-1] + current[0]

	for i := 1; i < len(current); i++ {
		prev = append(prev, current[i])
	}
	return prev
}

func fileNameWithoutExtSliceNotation(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

type Data struct {
	Id              string  `json:"id"`
	Price           float64 `json:"price"`
	Expiration_date string  `json:"expiration_date"`
}

type RecordsStorage struct {
	newStoreagePath  string
	file             *os.File
	isBusy           bool
	currecntFileName string
	recordSize       int
	oldFileName      string
	dataRoot         string
	mu               sync.RWMutex
}

func (rs *RecordsStorage) Close() error {
	rs.mu.Lock()
	os.Remove(rs.dataRoot + rs.currecntFileName)
	rs.mu.Unlock()
	return nil
}

func (rs *RecordsStorage) appendRecord(data string) {
	// should never happen
	if rs.file == nil {
		log.Fatal("file pointer cannot be null")
	}

	len, err := rs.file.Write([]byte(data))
	if err != nil {
		log.Fatal(err)
	}

	if len < rs.recordSize {
		rs.file.Write([]byte("\n"))
		for i := 0; i < rs.recordSize-1-len; i++ {
			// adding dummy symbols to align each line
			rs.file.Write([]byte("#"))
		}
	}
}

func (rs *RecordsStorage) getRecord(index int, root string) (string, error) {
	if root != fileNameWithoutExtSliceNotation(rs.currecntFileName) {
		return "", errors.New("Wrong path ")
	}

	// read lock
	rs.mu.RLock()

	f, err := os.Open(rs.dataRoot + rs.currecntFileName)
	_, err = f.Seek(int64((index-1)*rs.recordSize), 1)
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(f)
	line, _ := reader.ReadString('\n')

	// read unlock
	rs.mu.RUnlock()

	return line, nil
}

func (rs *RecordsStorage) makeReady(fileName string) error {
	rs.mu.RLock()
	if rs.isBusy {
		rs.mu.RUnlock()
		return errors.New("Server is busy. try leter")
	}

	rs.mu.RUnlock()

	rs.oldFileName = rs.currecntFileName
	rs.currecntFileName = fileName
	rs.file, _ = os.OpenFile(rs.newStoreagePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	rs.mu.Lock()
	rs.isBusy = true
	rs.mu.Unlock()

	return nil
}

func (rs *RecordsStorage) finish() {
	defer rs.file.Close()
	rs.mu.Lock()

	e := os.Rename(rs.newStoreagePath, rs.dataRoot+rs.currecntFileName)
	if len(rs.oldFileName) > 0 {
		fmt.Println("removing old file ", rs.oldFileName)
		e = os.Remove(rs.oldFileName)
		if e != nil {
			log.Fatal(e)
		}
	}

	rs.isBusy = false
	rs.mu.Unlock()
}

type DataCollector struct {
	storage          *RecordsStorage
	prevLine         []string
	currecntFileName string
	oldFileName      string
}

func (dc *DataCollector) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Header.Get("connection_reason") == "checking_ready_status" {
		// storage must be initialized
		err := dc.storage.makeReady(req.Header.Get("file_name"))
		if err == nil {
			io.WriteString(w, "ready")
		} else {
			io.WriteString(w, "Server isn't ready to fetch data")
		}
		return
	}

	if req.Header.Get("connection_reason") == "done" {
		// TODO storage must be initialized
		dc.storage.finish()
	}

	if req.Header.Get("connection_reason") != "sending_csv_records" {
		w.Write([]byte("wrong connection reason"))
		return
	}

	reader := csv.NewReader(req.Body)
	var firstIteration bool = true
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		if firstIteration {
			firstIteration = false
			if len(dc.prevLine) > 0 {
				record = makeString(dc.prevLine, record)
			}
		}

		dc.prevLine = record
		if len(record) == 3 {
			var iot Data

			iot.Id = record[0]

			floatVar, _ := strconv.ParseFloat(record[1], 64)
			iot.Price = floatVar

			res1 := strings.Split(record[2], " ")

			if len(res1) > 2 {
				iot.Expiration_date = res1[0] + " " + res1[1]

				json.Marshal(iot)
				b, err := json.Marshal(iot)
				if err != nil {
					fmt.Printf("Error: %s", err)
					return
				}

				dc.storage.appendRecord(string(b))
			}
		}
	}
}

type RecordsSender struct {
	recordsStorage *RecordsStorage
}

func (rs *RecordsSender) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if rs.recordsStorage == nil {
		log.Fatal("Data Storage is empty")
	}

	var path = html.EscapeString(r.URL.Path)
	endPointPaths := strings.Split(path, "/")

	if len(endPointPaths) == 3 {
		intVar, err := strconv.ParseInt(endPointPaths[2], 10, 0)

		foundRecord, err := rs.recordsStorage.getRecord(int(intVar), endPointPaths[1])
		if err != nil {
			w.Write([]byte("Record not found, wrong path"))
			log.Println(err)
			return
		}

		if len(foundRecord) > 0 {
			w.Write([]byte(foundRecord))
		} else {
			w.Write([]byte("Record not found"))
		}
	} else {
		w.Write([]byte("Wrong path or usage error"))
	}
}

func main() {
	storage := RecordsStorage{
		newStoreagePath: "data/.temperoery_storage.txt",
		file:            nil,
		isBusy:          false,
		recordSize:      110,
		dataRoot:        "data/",
	}

	rCollector := &DataCollector{storage: &storage}
	rSender := &RecordsSender{recordsStorage: &storage}

	var port string
	flag.StringVar(&port, "port", "", "Usage")
	flag.Parse()

	if port == "" {
		port = "1321"
	}

	http.Handle("/", rSender)
	http.Handle("/data", rCollector)

	err := http.ListenAndServeTLS(":"+port, "cert/localhost.test.crt", "cert/localhost.test.key", nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
