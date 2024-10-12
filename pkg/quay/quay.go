package quay

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetQuayRobotToken(fedToken string, robotAccount string, url string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://"+url+"/oauth2/federation/robot/token", nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(robotAccount, fedToken)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.Status != "200 OK" {
		return "", fmt.Errorf(resp.Status)
	}

	// fmt.Println("Response Status:", resp.Status)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}

	json.Unmarshal([]byte(body), &result)
	return result["token"].(string), nil
}
