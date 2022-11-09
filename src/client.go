package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func openFile(filePath string) {
	var bufferSize int = 1024 * 1024

	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("cannot able to read the file", err)
		return
	}

	r := bufio.NewReader(f)
	fmt.Printf("start !!!")

	for {
		buf := make([]byte, bufferSize) //the chunk size
		n, err := r.Read(buf)           //loading chunk into buffer
		buf = buf[:n]
		if errors.Is(err, io.EOF) { // prefered way by GoLang doc
			fmt.Println("Reading file finished...")
			break
		}

		resp, err := http.Post("http://localhost:3333/data", "bytes/string", bytes.NewReader(buf))
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
		// parsing body
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			fmt.Printf("Body : %s", body)
			return
		}
	}

	fmt.Printf("Done !!!")
	// UPDATE: close after checking error
	defer f.Close() //Do not forget to close the file
}

func main() {
	openFile("promotions.csv")
}
