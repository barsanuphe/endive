package helpers

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
)

// GetXMLData retrieves XML responses from online APIs.
func GetXMLData(uri string, i interface{}) {
	data := getRequest(uri)
	xmlUnmarshal(data, i)
}

func getRequest(uri string) []byte {
	res, err := http.Get(uri)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return body
}

func xmlUnmarshal(b []byte, i interface{}) {
	err := xml.Unmarshal(b, i)
	if err != nil {
		log.Fatal(err)
	}
}
