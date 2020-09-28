package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	"github.com/pkg/errors"
)

type Response struct {
	Code        string
	Body        []byte
	ContentType string
}

const Dir = "../../html"

// Problem: create http server
// = socket interface
// = i0 interface
// = handler

// require headers, status codes and body of the meesage
func main() {

	//
	//
	//
	ln, err := setUpListener()

	fatalerror(err)

	for {
		conn, err := acceptConnectionFromListen(ln)
		fatalerror(err)

		requestRawBuffer, err := readFromConnection(conn)
		fatalerror(err)

		requestObj, err := requestParse(requestRawBuffer)
		fatalerror(err)

		bytResponse, err := readFileFromDisk(requestObj.Path)

		response := &Response{
			Code:        "200 OK",
			Body:        bytResponse,
			ContentType: "text/html",
		}

		if err != nil {
			//todo proper code handler on the error
			response.Code = "500 OK"
			response.Body = []byte("")
		}

		fatalerror(requestResponse(conn, response))
	}
}

func fatalerror(err error) {
	if err != nil {
		panic(err)
	}
}

// read file from disks and return bytes

func readFileFromDisk(filename string) ([]byte, error) {

	var bty []byte
	var err error
	if filename[len(filename):] == "/" {
		filename = "/index.html"
	}

	pathAndFile := Dir + filename

	if bty, err = ioutil.ReadFile(pathAndFile); err != nil {

		return nil, errors.Wrap(err, "readFileFromDisk() - ")

	}

	return bty, nil

}

// socket inteface for accepting client queries
// curl https://localhost/path

func setUpListener() (net.Listener, error) {

	ln, err := net.Listen("tcp", ":8888")

	if err != nil {
		return nil, errors.Wrap(err, "setupListener() - ")
	}

	return ln, nil

}

// parse the socket input

func acceptConnectionFromListen(ln net.Listener) (net.Conn, error) {

	conn, err := ln.Accept()

	if err != nil {
		return nil, errors.Wrap(err, "acceptConnectionFromListener")
	}

	return conn, nil

}

// read from connection
func readFromConnection(conn net.Conn) ([][]byte, error) {

	r := bufio.NewReader(conn)
	var requestObj [][]byte
	for {

		line, _, err := r.ReadLine()
		fmt.Printf("Getting line: %s: [%s]\n", string(line), err)

		if len(line) == 0 || (err != nil && err.Error() == "EOF") { // todo
			return requestObj, nil
		}

		if err != nil {
			return nil, errors.Wrap(err, "readFromConnection() - ")
		}

		requestObj = append(requestObj, line)

	}

	return nil, errors.New("Impossible error")

}

type ParseRequest struct {
	Directive string
	Path      string
	Protocol  string
}

// parse the request
func requestParse(bty [][]byte) (*ParseRequest, error) {
	fmt.Printf("RequestParse() - %v\n", bty)
	for _, elm := range bty {

		headerLine := string(elm)
		strArr := strings.Split(headerLine, " ")
		responseObj := &ParseRequest{
			strArr[0], strArr[1], strArr[2],
		}

		if responseObj.Directive != "GET" {
			return nil, errors.New("INVALID_REQUEST")
		}

		if responseObj.Directive == "GET" {
			return responseObj, nil
		}
	}
	return nil, errors.New("INVALID RESPONSE ERROR")

}

// response assumes a valid response
func requestResponse(conn net.Conn, res *Response) error {
	defer conn.Close()
	fmt.Printf("About to respond\n")
	// protocol response
	responseString := fmt.Sprintf("HTTP/1.1 %s\n\r", res.Code)
	_, err := conn.Write([]byte(responseString))
	if err != nil {
		return err
	}

	// headers
	_, err = conn.Write([]byte("Content-Type: " + res.ContentType + "\n\r"))
	if err != nil {
		return err
	}

	contentLenStr := fmt.Sprintf("Content-Length: %d \n\r", len(res.Body))
	// headers
	_, err = conn.Write([]byte(contentLenStr))
	if err != nil {
		return err
	}

	// seperator
	_, err = conn.Write([]byte("\n\r"))
	if err != nil {
		return err
	}

	// body
	_, err = conn.Write(res.Body) // todo for content length we will need the bytes returnd
	if err != nil {
		return err
	}

	return nil

}
