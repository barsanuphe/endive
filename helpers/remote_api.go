package helpers

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"time"
)

// GetXMLData retrieves XML responses from online APIs.
func GetXMLData(uri string, i interface{}) {
	data := getRequest(uri)
	xmlUnmarshal(data, i)
}

func getRequest(uri string) (body []byte) {
	// 10s timeout
	timeout := time.Duration(10 * time.Second)
	client := http.Client{Timeout: timeout}
	res, err := client.Get(uri)
	if err != nil {
		Error(err.Error())
		return
	}

	body, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		Error(err.Error())
	}
	return
}

func xmlUnmarshal(b []byte, i interface{}) {
	err := xml.Unmarshal(b, i)
	if err != nil {
		Error(err.Error())
	}
}
