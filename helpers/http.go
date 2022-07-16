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
