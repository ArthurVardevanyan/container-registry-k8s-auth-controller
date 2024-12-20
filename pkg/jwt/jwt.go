package jwt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func getClaim(tokenString string) (map[string]interface{}, error) {
	// Split the token into its three parts
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("Invalid token format")
	}

	// Decode the payload (the second part of the token)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("Error decoding payload: %v", err)
	}

	// Extract the exp claim
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("Error un-marshalling claims: %v", err)
	} else {
		return claims, nil
	}
}

func TokenExpiration(tokenString string) (string, error) {

	claims, err := getClaim(tokenString)
	if err != nil {
		return "", fmt.Errorf("Error getting claims: %v", err)
	}

	exp, ok := claims["exp"].(float64)
	if ok {
		expirationTime := time.Unix(int64(exp), 0).UTC()
		return expirationTime.String(), nil
	} else {
		return "", fmt.Errorf("exp claim not found or wrong type")

	}
}

func Issuer(tokenString string) (string, error) {

	claims, err := getClaim(tokenString)
	if err != nil {
		return "", fmt.Errorf("Error getting claims: %v", err)
	}

	issuer, ok := claims["iss"].(string)
	if ok {
		return issuer, nil
	} else {
		return "", fmt.Errorf("exp claim not found or wrong type")

	}
}

func Subject(tokenString string) (string, error) {

	claims, err := getClaim(tokenString)
	if err != nil {
		return "", fmt.Errorf("Error getting claims: %v", err)
	}

	subject, ok := claims["sub"].(string)
	if ok {
		return subject, nil
	} else {
		return "", fmt.Errorf("exp claim not found or wrong type")

	}
}
