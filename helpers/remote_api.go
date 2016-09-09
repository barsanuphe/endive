package helpers

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"time"
)

// GetXMLData retrieves XML responses from online APIs.
func GetXMLData(uri string, i interface{}) (err error) {
	currentPass := 0
	const maxTries = 5
	var data []byte
	for currentPass < maxTries {
		data, err = getRequest(uri)
		if err != nil {
			currentPass++
			// wait a little
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}
	// test if the last pass was successful
	if err != nil {
		return
	}
	return xml.Unmarshal(data, i)
}

func getRequest(uri string) (body []byte, err error) {
	// 10s timeout
	timeout := time.Duration(10 * time.Second)
	client := http.Client{Timeout: timeout}
	res, err := client.Get(uri)
	if err != nil {
		return body, err
	}

	body, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
	}
	return
}
