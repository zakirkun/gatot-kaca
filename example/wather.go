package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// WeatherTool implements the Tool interface and fetches weather data for a given city.
// It retrieves a short weather forecast from wttr.in.
type WeatherTool struct{}

// Name returns the name of the weather tool.
func (w WeatherTool) Name() string {
	return "weather"
}

// Description returns a brief description of the weather tool.
func (w WeatherTool) Description() string {
	return "Fetches current weather information for a given city using wttr.in"
}

// Execute makes an HTTP GET request to wttr.in to get a concise weather report.
// The input should be a city name.
func (w WeatherTool) Execute(ctx context.Context, input string) (string, error) {
	city := input
	if city == "" {
		return "", fmt.Errorf("city name must be provided")
	}
	// Use the wttr.in API (plain text format).
	// remove \n
	city = strings.ReplaceAll(city, "\n", "")
	url := fmt.Sprintf("http://wttr.in/%s?format=3", city)

	// Create an HTTP request with a timeout set in the context.
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
