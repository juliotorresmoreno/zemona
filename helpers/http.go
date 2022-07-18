package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

func getContentType(r *http.Request) string {
	for key := range r.Header {
		if strings.ToLower(key) == "content-type" {
			return r.Header.Get(key)
		}
	}
	return ""
}

//GetPostParams Get the parameters sent by the post method in an http request
func GetPostParams(r *http.Request) url.Values {
	contentType := getContentType(r)
	switch {
	case contentType == "application/json":
		params := map[string]interface{}{}
		result := url.Values{}
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&params)
		if err != nil {
			fmt.Println(err)
		}
		for k, v := range params {
			if reflect.ValueOf(v).Kind().String() == "string" {
				result.Set(k, v.(string))
			}
		}
		return result
	case contentType == "application/x-www-form-urlencoded":
		r.ParseForm()
		return r.Form
	case strings.Contains(contentType, "multipart/form-data"):
		r.ParseMultipartForm(int64(10 * 1000))
		return r.Form
	}
	return url.Values{}
}

// Cors permite el acceso desde otro servidor
func Cors(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}

	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Cache-Control")
	w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,OPTIONS")
}

func GetCookie(r *http.Request, name string) string {
	var c string = r.Header.Get("Cookie")
	var s []string = strings.Split(c, ";")
	var t []string
	for i := 0; i < len(s); i++ {
		t = strings.Split(strings.Trim(s[i], " "), "=")
		if t[0] == name {
			return t[1]
		}
	}
	return ""
}

// GetToken retorna el token
func GetToken(r *http.Request) string {
	var _token = r.URL.Query().Get("token")
	if _token == "" {
		_token = GetCookie(r, "token")
	}
	if _token == "" {
		_token = r.Header.Get("authorization")
	}
	return _token
}
