package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

var hostnamePort string = "https://localhost:1321"

func openFile(filePath string) {
	var bufferSize int = 1024 * 1024

	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("cannot able to read the file", err)
		return
	}

	r := bufio.NewReader(f)
	client := http.DefaultClient

	req, err := http.NewRequest(http.MethodGet, hostnamePort+"/data", nil)
	req.Header.Set("file_name", filePath)
	req.Header.Set("connection_reason", "checking_ready_status")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	if string(body) != "ready" {
		fmt.Printf("client: Server isn't ready to fetch data try leter", string(body))
		return
	}

	for {
		buf := make([]byte, bufferSize)
		n, err := r.Read(buf)
		buf = buf[:n]
		if errors.Is(err, io.EOF) {
			break
		}

		req, err := http.NewRequest(http.MethodPost, hostnamePort+"/data", bytes.NewReader(buf))
		if err != nil {
			fmt.Printf("client: could not create request: %s\n", err)
		}

		req.Header.Set("Content-Type", "bytes/string")
		req.Header.Set("file_name", filePath)
		req.Header.Set("connection_reason", "sending_csv_records")
		resp, err := client.Do(req)

		if n == 0 {
			if err != nil {
				fmt.Println(err)
				break
			}
			if err == io.EOF {
				fmt.Println("finished...")
				break
			}
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Body : %s", body)
			return
		}
	}

	req, err = http.NewRequest(http.MethodGet, hostnamePort+"/data", nil)
	req.Header.Set("file_name", filePath)
	req.Header.Set("connection_reason", "done")
	_, err = client.Do(req)
	if err != nil {
		fmt.Printf("Body : %s", body)
		return
	}

	fmt.Println("Data successful sent to server !!!")
	defer f.Close()
}

func main() {
	var name string
	flag.StringVar(&name, "file", "", "Usage")
	flag.Parse()

	openFile(name)
}
