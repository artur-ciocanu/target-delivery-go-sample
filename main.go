package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/artur-ciocanu/target-delivery-go-sample/visitor"
	"github.com/google/uuid"
)

const TEMPLATE = `
<!doctype html>
<html>
<head>
	<meta charset="UTF-8">
	<link rel="icon" href="data:,">
  <title>ECID and Analytics integration Sample</title>
  <script src="static/VisitorAPI.js"></script>
  <script>
		var visitor = Visitor.getInstance("${organizationId}", {serverState: ${visitorState}});
  </script>
</head>
<body>
	<p>${content}</p>
  <script src="static/AppMeasurement.js"></script>
  <script>var s_code=s.t();if(s_code)document.write(s_code);</script>
</body>
</html>
`
const IMS_ORG_ID = "011B56B451AE49A90A490D4D@AdobeOrg"
const DELIVERY_URL = "http://bullseye.tt.omtrdc.net/rest/v1/delivery?client=bullseye&"
const DELIVERY_API_PAYLOAD = `
{
	"context": {
		"channel": "web",
		"address": {
				"url": "http://10.58.114.112/bullseye/Unofolio.html?type=abt&rs=analytics&gm=analytics"
		}
	},
	"experienceCloud": {
		"analytics": {
			"supplementalDataId": "${sdid}",
			"trackingServer": "adobetargeteng.d1.sc.omtrdc.net",
			"trackingServerSecure": "adobetargeteng.d1.sc.omtrdc.net"
		}
	},
	"execute": {
		"pageLoad": {}
	}
}
`

func toJson(state map[string]*visitor.SupplementalDataIdState) ([]byte, error) {
	data, err := json.Marshal(state)

	if err != nil {
		log.Println("JSON marhalling error", err)

		return nil, err
	}

	return data, nil
}

func getDeliveryPayload(sdid string) string {
	return strings.Replace(DELIVERY_API_PAYLOAD, "${sdid}", sdid, -1)
}

func fetchTargetData(sessionId string, payload string) (string, error) {
	data := bytes.NewBuffer([]byte(DELIVERY_API_PAYLOAD))
	url := DELIVERY_URL + "sessionId=" + sessionId
	resp, err := http.Post(url, "application/json", data)

	if err != nil {
		log.Println("HTTP error", err)

		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println("Response body read error", err)

		return "", err
	}

	return string(body), nil
}

func getOrCreateSessionId() string {
	id := uuid.New()

	return id.String()
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	sessionId := getOrCreateSessionId()
	// Here we use global mbox if we have multiple mboxes
	// we can concatenate using ":" and pass as a single string
	sdid := visitor.GetSupplementalDataId("target-global-mbox")
	payload := getDeliveryPayload(sdid)
	content, err := fetchTargetData(sessionId, payload)

	log.Println("Target payload", payload)

	if err != nil {
		log.Println("Target request failed", err)

		w.Header().Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.121 Safari/537.36")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, err.Error())
		return
	}

	log.Println("Target response", content)

	state := visitor.GetState(IMS_ORG_ID)
	visitorState, err := toJson(state)

	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, err.Error())
		return
	}

	response := strings.Replace(TEMPLATE, "${organizationId}", IMS_ORG_ID, -1)
	response = strings.Replace(response, "${visitorState}", string(visitorState), -1)
	response = strings.Replace(response, "${content}", content, -1)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, response)
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", handleRequest)

	http.ListenAndServe(":3000", nil)
}
