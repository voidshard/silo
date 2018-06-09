package main

import (
	"net/http"
	"bytes"
	"testing"
	"io"
	"io/ioutil"
	"reflect"
)

const (
	configFile = "etc/silo.ini"
)

// parsed from cmd/silo/config.go
var cfg *Config
var users *fileConfig
var client *http.Client

func init() {
	config, err := parseConfig(configFile)
	if err != nil {
		panic(err)
	}

	tmp, err := readConfigFile(configFile)
	if err != nil {
		panic(err)
	}

	users = tmp
	client = NewClient()
	cfg = config
}

func DoRequest(method, url string, body io.Reader, user *entity) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body, )
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", BasicAuth(user.Id, user.Password))
	return client.Do(req)
}

func TestWrite(t *testing.T) {
	cases := []struct{
		Key string
		Method string
		Data []byte
		User *entity
		ExpectOnWrite int // expected status code (http)
		ExpectOnRead int
		ExpectEcho bool // we expect our data back
	}{
		{
			Key: "somekey",
			Method: http.MethodPost,
			Data: []byte("some data about ponies"),
			User: ReadOnlyRole(users),
			ExpectOnWrite: http.StatusForbidden, // we should be denied writing
			ExpectOnRead: http.StatusNotFound,
			ExpectEcho: false,
		},

		{
			Key: "somekey",
			Method: http.MethodPost,
			Data: []byte("some data about ponies"),
			User: ReadWriteRole(users),  // now we can write
			ExpectOnWrite: http.StatusOK,
			ExpectOnRead: http.StatusOK,
			ExpectEcho: true,
		},

		{
			Key: "somekey",
			Method: http.MethodPost,
			Data: []byte("some data about ponies"),
			User: ReadWriteRole(users),
			ExpectOnWrite: http.StatusForbidden,  // we've written this already and we lack RM, so should get denied
			ExpectOnRead: http.StatusOK,
			ExpectEcho: true,
		},

		{
			Key: "somekey",
			Method: http.MethodPost,
			Data: []byte("some data about ponies"),
			User: AllRole(users),
			ExpectOnWrite: http.StatusBadRequest,  // we've written this already, we have RM, so should get "use PUT"
			ExpectOnRead: http.StatusOK,
			ExpectEcho: true,
		},

		{
			Key: "somekey",
			Method: http.MethodPut,
			Data: []byte("some other data about ponies"),
			User: AllRole(users),
			ExpectOnWrite: http.StatusOK,
			ExpectOnRead: http.StatusOK,
			ExpectEcho: true,
		},
	}

	for i, tst := range cases {
		if tst.User == nil {
			t.Skip(i, "user not found")
		}

		// Write
		resp, err := DoRequest(tst.Method, Url(tst.Key, cfg.Server.HttpPort), bytes.NewBuffer(tst.Data), tst.User)
		if err != nil {
			t.Error(i, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != tst.ExpectOnWrite {
			t.Error(i, "expected status", tst.ExpectOnWrite, "got", resp.StatusCode)
		}

		// Read
		resp, err = DoRequest(http.MethodGet, Url(tst.Key, cfg.Server.HttpPort), nil, tst.User)
		if err != nil {
			t.Error(i, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != tst.ExpectOnRead {
			t.Error(i, "expected status", tst.ExpectOnRead, "got", resp.StatusCode)
		}

		if tst.ExpectEcho {
			result, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Error(i, err)
				continue
			}

			if !reflect.DeepEqual(result, tst.Data) {
				t.Error(i, "expected", tst.Data, "got", result)
			}
		}
	}
}