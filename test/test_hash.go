package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

const salt = "mysalt"

func hashPassword(password string) string {
	hash := md5.Sum([]byte(password + salt))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func main() {
	password := "admin123"
	hashed := hashPassword(password)
	fmt.Printf("Hashed password for '%s': %s\n", password, hashed)

	stored := "rzG+9lNc5LLNbJttRgegfA=="
	fmt.Printf("Stored hash: %s\n", stored)
	fmt.Printf("Match: %t\n", hashed == stored)

	// Test login
	fmt.Println("\nTesting login...")
	testLogin()
}

func testLogin() {
	url := "http://localhost:9090/login"
	payload := map[string]string{
		"email":    "admin@example.com",
		"password": "admin123",
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("Response: %v\n", result)
}
