package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var details = map[string]interface{}{}

func main() {
	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Call Center listening on port " + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	sessionId := r.FormValue("sessionId")

	details, err := getDetails(sessionId)
	if err != nil {
		log.Fatal(err)
	}

	if r.FormValue("isActive") == "0" {
		callEnded(w, r)
		return
	}

	digits := r.FormValue("dtmfDigits")
	fmt.Println("DTMF: " + digits)

	language, ok := details["language_selected"]
	if !ok || language.(bool) != true {
		switch digits {
		case "1":
			english(w, r)

		case "2":
			pidgin(w, r)

		default:
			base(w, r)
		}

		return
	}

	switch digits {
	case "1":
		support(w, r)

	case "2":
		sales(w, r)

	default:
		unknownOption(w, r)
	}

	return
}

func base(w http.ResponseWriter, r *http.Request) {
	text := "Welcome. Select your language. Press 1 for english, 2 for pidgin."

	response := `<?xml version="1.0" encoding="UTF-8"?>
		<Response>
		  <GetDigits  timeout="10">
		    <Say voice="man" playBeep="true">` + text + `</Say>
		  </GetDigits>
		</Response>`

	w.Write([]byte(response))

}

func english(w http.ResponseWriter, r *http.Request) {
	details["language_selected"] = true
	details["language"] = "english"
	updateDetails(r.FormValue("sessionId"), details)

	text := "You selected english, press 1 to speak to support, press 2 to speak to sales."

	response := `<?xml version="1.0" encoding="UTF-8"?>
		<Response>
		  <GetDigits  timeout="10">
		    <Say voice="man" playBeep="true">` + text + `</Say>
		  </GetDigits>
		</Response>`

	w.Write([]byte(response))
}

func pidgin(w http.ResponseWriter, r *http.Request) {
	details["language_selected"] = true
	details["language"] = "pidgin"
	updateDetails(r.FormValue("sessionId"), details)

	text := "You don select pdigin, press 1 to follow support talk, press 2 for sales."

	response := `<?xml version="1.0" encoding="UTF-8"?>
		<Response>
		  <GetDigits  timeout="10">
		    <Say voice="man" playBeep="true">` + text + `</Say>
		  </GetDigits>
		</Response>`

	w.Write([]byte(response))
}

func support(w http.ResponseWriter, r *http.Request) {
	language := details["language"]

	text := "Your call is being forwarded to a support agent. Note that this call may be recorded."
	phoneNumbers := os.Getenv("SUPPORT_PHONES_ENG")

	if language == "pidgin" {
		text = "We dey connect you to one of our support people. Know say we dey record this call."
		phoneNumbers = os.Getenv("SUPPORT_PHONES_PNG")
	}

	response := `<?xml version="1.0" encoding="UTF-8"?>
		<Response>
		  <Say voice="man" playBeep="true">` + text + `</Say>
		  <Dial phoneNumbers="` + phoneNumbers + `" record="true" sequential="false"/>
		</Response>`

	w.Write([]byte(response))
}

func sales(w http.ResponseWriter, r *http.Request) {
	language := details["language"]

	text := "Your call is being forwarded to a sales agent. Note that this call may be recorded."
	phoneNumbers := os.Getenv("SALES_PHONES_ENG")

	if language == "pidgin" {
		text = "We dey connect you to one of our sales people. Know say we dey record this call."
		phoneNumbers = os.Getenv("SALES_PHONES_PNG")
	}

	response := `<?xml version="1.0" encoding="UTF-8"?>
		<Response>
		  <Say voice="man" playBeep="true">` + text + `</Say>
		  <Dial phoneNumbers="` + phoneNumbers + `" record="true" sequential="false"/>
		</Response>`

	w.Write([]byte(response))
}

func unknownOption(w http.ResponseWriter, r *http.Request) {
	text := "You have selected an unknown option"
	if details["language"] == "pidgin" {
		text = "We no understand the option wey you select."
	}

	response := `<?xml version="1.0" encoding="UTF-8"?>
		    <Say voice="man" playBeep="true">` + text + `</Say>
		</Response>`

	w.Write([]byte(response))
}

func callEnded(w http.ResponseWriter, r *http.Request) {
	for key, _ := range r.Form {
		details[key] = r.FormValue(key)
	}

	updateDetails(r.FormValue("sessionId"), details)
	w.Write([]byte("call logged"))
	return
}

func getDetails(session string) (deets map[string]interface{}, err error) {
	path := "./data/" + session + ".json"

	_, err = os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			err = ioutil.WriteFile(path, []byte("{}"), 0666)
			if err != nil {
				return deets, err
			}
		} else {
			return deets, err
		}
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return deets, err
	}

	err = json.Unmarshal(content, &deets)
	return deets, err
}

func updateDetails(session string, newDetails map[string]interface{}) (err error) {
	path := "./data/" + session + ".json"

	for k, v := range newDetails {
		details[k] = v
	}

	deetBytes, err := json.MarshalIndent(details, "", "  ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(path, []byte(deetBytes), 0666); err != nil {
		return err
	}

	return
}
