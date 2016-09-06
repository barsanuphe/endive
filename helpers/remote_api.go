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

	err = xmlUnmarshal(data, i)
	if err != nil {
		Error("Could not parse GoodReads response")
	}
	return
}

func getRequest(uri string) (body []byte, err error) {
	// 10s timeout
	timeout := time.Duration(10 * time.Second)
	client := http.Client{Timeout: timeout}
	res, err := client.Get(uri)
	if err != nil {
		Error(err.Error())
		return body, err
	}

	body, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		Error(err.Error())
	}
	return
}

func xmlUnmarshal(b []byte, i interface{}) (err error) {
	err = xml.Unmarshal(b, i)
	if err != nil {
		Error(err.Error())
	}
	return
}
