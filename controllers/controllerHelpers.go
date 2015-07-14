package controllers

import (
	"fmt"
	"net/http"

	"github.com/TeamPlayTF/Server/config"
	"github.com/TeamPlayTF/Server/helpers"
	"github.com/bitly/go-simplejson"
	"github.com/gorilla/sessions"
)

func sendJSON(w http.ResponseWriter, json *simplejson.Json) {
	w.Header().Add("Content-Type", "application/json")
	val, _ := json.String()
	fmt.Fprintf(w, val)
}

func buildSuccessJSON(data *simplejson.Json) *simplejson.Json {
	j := simplejson.New()
	j.Set("success", true)
	j.Set("data", data)

	return j
}

func buildFailureJSON(message string, code int) *simplejson.Json {
	e := helpers.NewTPError(message, code)
	return e.ErrorJSON()
}

func buildFakeSocketRequest(request *simplejson.Json) *http.Request {
	cookiesObj := request.Get("cookies")

	if cookiesObj == nil {
		return &http.Request{}
	}

	cookies, err := cookiesObj.Map()
	if err != nil {
		return &http.Request{}
	}

	str := ""

	first := true
	for k, v := range cookies {
		vStr, ok := v.(string)
		if !ok {
			continue
		}

		if !first {
			str += ";"
		}
		str += k + "=" + vStr
		first = false
	}

	if str == "" {
		return &http.Request{}
	}

	headers := http.Header{}
	headers.Add("Cookie", str)

	return &http.Request{Header: headers}
}

func isLoggedIn(r *http.Request) bool {
	session, _ := config.CookieStore.Get(r, config.Constants.SessionName)

	val, ok := session.Values["steamid"]
	return ok && val != ""
}

func getDefaultSession(r *http.Request) *sessions.Session {
	session, _ := config.CookieStore.Get(r, config.Constants.SessionName)
	return session
}

func redirectHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, config.Constants.Domain, 303)
}
