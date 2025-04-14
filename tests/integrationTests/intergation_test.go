package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

const baseURL = "http://app:8080"

func getToken(t *testing.T, role string) string {
	url := baseURL + "/dummyLogin"
	payload := map[string]string{"role": role}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal data for dummyLogin: %v", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to perform dummyLogin request for role %s: %v", role, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("dummyLogin for role %s returned status %d", role, resp.StatusCode)
	}

	var token string
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		t.Fatalf("Failed to decode token for role %s: %v", role, err)
	}
	t.Logf("Token received for role %s: %s", role, token)
	return token
}

func createPVZ(t *testing.T, token string) map[string]interface{} {
	url := baseURL + "/pvz"
	payload := map[string]string{"city": "Москва"}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal data for PVZ creation: %v", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create request for PVZ creation: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute PVZ creation request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("PVZ creation failed: status %d, response: %s", resp.StatusCode, string(data))
	}

	var pvz map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&pvz); err != nil {
		t.Fatalf("Failed to decode PVZ creation response: %v", err)
	}
	t.Logf("PVZ created: %v", pvz)
	return pvz
}

func createReception(t *testing.T, token, pvzId string) map[string]interface{} {
	url := baseURL + "/receptions"
	payload := map[string]string{"pvzId": pvzId}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal data for reception creation: %v", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create request for reception creation: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute reception creation request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("Reception creation failed: status %d, response: %s", resp.StatusCode, string(data))
	}

	var reception map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&reception); err != nil {
		t.Fatalf("Failed to decode reception creation response: %v", err)
	}
	t.Logf("Reception created: %v", reception)
	return reception
}

func addProduct(t *testing.T, token, pvzId, productType string) {
	url := baseURL + "/products"
	payload := map[string]string{
		"pvzId": pvzId,
		"type":  productType,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal data for product addition: %v", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create request for product addition: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute product addition request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("Product addition failed: status %d, response: %s", resp.StatusCode, string(data))
	}
}

func closeReception(t *testing.T, token, pvzId string) map[string]interface{} {
	url := fmt.Sprintf("%s/pvz/%s/close_last_reception", baseURL, pvzId)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		t.Fatalf("Failed to create request for reception closure: %v", err)
	}
	req.Header.Set("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute reception closure request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("Reception closure failed: status %d, response: %s", resp.StatusCode, string(data))
	}

	var closedReception map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&closedReception); err != nil {
		t.Fatalf("Failed to decode reception closure response: %v", err)
	}
	t.Logf("Reception closed: %v", closedReception)
	return closedReception
}

func TestIntegration(t *testing.T) {
	modToken := getToken(t, "moderator")
	pvz := createPVZ(t, modToken)

	pvzId, ok := pvz["id"].(string)
	if !ok || pvzId == "" {
		t.Fatal("Invalid PVZ ID")
	}

	empToken := getToken(t, "employee")
	_ = createReception(t, empToken, pvzId)

	for i := 1; i <= 50; i++ {
		addProduct(t, empToken, pvzId, "электроника")
	}
	t.Log("50 products added successfully")

	closedReception := closeReception(t, empToken, pvzId)
	if status, exists := closedReception["status"].(string); !exists || status != "close" {
		t.Fatalf("Reception was not properly closed, received status: %v", closedReception["status"])
	}

	t.Log("Integration test completed successfully")
}
