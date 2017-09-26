package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/philippmachs/machsdb/db"
	"net/http"
	"time"
)

type AppError struct {
	error string
}

func (a AppError) Error() string {
	return a.error
}

type Client struct {
	httpclient *http.Client
	url        string
	login      string
	password   string
}

func NewClient(url string, timeout int, login string, password string) *Client {
	c := &Client{
		&http.Client{Timeout: time.Duration(timeout) * time.Second},
		url,
		login,
		password,
	}
	return c
}

func (c *Client) Set(key string, value interface{}, expire time.Duration) (result *db.Value, err error) {
	body, err := json.Marshal(value)
	if err != nil {
		return
	}
	url := fmt.Sprintf("%v/%v?ttl=%v", c.url, key, expire)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	c.setHeaders(req)

	resp, err := c.httpclient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	if err = extractError(resp, decoder); err != nil {
		return
	}

	result = &db.Value{}
	err = decoder.Decode(result)
	return
}

func (c *Client) Get(key string) (result *db.Value, err error) {

	url := fmt.Sprintf("%v/%v", c.url, key)
	req, err := http.NewRequest("GET", url, nil)
	c.setHeaders(req)

	resp, err := c.httpclient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	if err = extractError(resp, decoder); err != nil {
		return
	}

	result = &db.Value{}
	err = decoder.Decode(result)
	return

}
func (c *Client) Keys() (keys []string, err error) {
	req, err := http.NewRequest("GET", c.url, nil)
	c.setHeaders(req)

	resp, err := c.httpclient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	if err = extractError(resp, decoder); err != nil {
		return
	}

	keys = []string{}
	err = decoder.Decode(&keys)
	return
}

func (c *Client) GetAtIndex(key string, index interface{}) (result interface{}, err error) {
	url := fmt.Sprintf("%v/%v/%v", c.url, key, index)
	req, err := http.NewRequest("GET", url, nil)
	c.setHeaders(req)

	resp, err := c.httpclient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	if err = extractError(resp, decoder); err != nil {
		return
	}

	err = decoder.Decode(&result)
	return
}

func (c *Client) Remove(key string) (err error) {
	url := fmt.Sprintf("%v/%v", c.url, key)
	req, err := http.NewRequest("DELETE", url, nil)
	c.setHeaders(req)

	resp, err := c.httpclient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	if resp.StatusCode == http.StatusOK {
		return nil
	}
	if resp.StatusCode == http.StatusBadRequest {
		return decodeAppError(decoder)
	}
	return errors.New(http.StatusText(resp.StatusCode))
}

func (c *Client) setHeaders(r *http.Request) {
	r.Header.Set("Content-Type", "application/json")

	if c.login != "" && c.password != "" {
		r.SetBasicAuth(c.login, c.password)
	}
}

func extractError(resp *http.Response, decoder *json.Decoder) error {
	if resp.StatusCode == http.StatusBadRequest {
		return decodeAppError(decoder)
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(http.StatusText(resp.StatusCode))
	}
	return nil
}

func decodeAppError(decoder *json.Decoder) error {
	appError := AppError{}
	err := decoder.Decode(&appError)
	if err != nil {
		return errors.New(http.StatusText(http.StatusBadRequest))
	}
	return appError
}
