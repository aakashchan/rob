package ccacenue

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var (
	testSubDomain    = "test"
	prodSubDomain    = "secure"
	subDomain        = testSubDomain
	EncryptionKeyUrl = fmt.Sprintf("https://%s.ccavenue.com/transaction/getRSAKey", subDomain)
)

var (
	merchantId = "145970"
	accessCode string
	workingKey string
)

var (
	testAccessCode = "AVJI01EI22AO92IJOA"
	testWorkingKey = "4BCF230B635E0040976D8119CB21FCB8"
)

func GetEncryptionKey(orderId string) (string, error) {
	data := url.Values{}
	data.Set("access_code", accessCode)
	data.Set("order_id", orderId)
	req, err := http.NewRequest("POST", EncryptionKeyUrl, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if res.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("Wrong return code: %d", res.StatusCode))
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	b := string(body)
	if strings.Contains(b, "ERROR") {
		return "", errors.New(fmt.Sprintf("Return body contains the word ERROR. Body: %s", b))
	}
	return b, nil
}

func init() {
	if subDomain == testSubDomain {
		accessCode = testAccessCode
		workingKey = testWorkingKey
	} else {
		// TODO: Get the keys from .ccavenue_creds
	}
	/*
		s, err := GetEncryptionKey("1")
		if err != nil {
			panic(err)
		}
		fmt.Println(s)
	*/
}
