package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Define the fields based on the JSON structure from the API response.
type Price struct {
	SEK_price float64 `json:"SEK_per_kWh"`
}

func fetchPrices(context *gin.Context) {

	var resp *http.Response
	var err error

	resp, err = http.Get("https://www.elprisetjustnu.se/api/v1/prices/2024/11-29_SE3.json")
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch prices"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	var prices []Price
	if err := json.Unmarshal(body, &prices); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse JSON"})
		return
	}

	context.JSON(http.StatusOK, prices)
}

func main() {
	router := gin.Default()
	router.GET("/prices", fetchPrices)
	router.Run("localhost:9090") // Running the server on port 9090
}
