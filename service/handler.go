package service

import (
	"log"
	"net/http"
	"strings"
	"html/template"
    "net/url"
    "encoding/json"
    "fmt"
	"io"
	"strconv"
)


func searchProductImages(productName string) ([]string,error) {

    apiKey := "AIzaSyAR1QmJjDCU-6J30hJBnYOQJwFRplPqEiE"
	cx := "a78134ac624874cd1"

    query:=url.QueryEscape(productName+"product image")
    apiURL := fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&q=%s&searchType=image&num=3", 
    apiKey, cx, query)
    client:=&http.Client{}
    req,err:=http.NewRequest("GET", apiURL, nil)
    if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
    	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Parse the JSON response
	var searchResult struct {
		Items []struct {
			Link string `json:"link"`
		} `json:"items"`
	}

	err = json.Unmarshal(body, &searchResult)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	// Extract image URLs
	imageUrls := []string{}
	for _, item := range searchResult.Items {
		imageUrls = append(imageUrls, item.Link)
	}

	return imageUrls, nil

    
}


















func SuggestGiftHandler(w http.ResponseWriter, r *http.Request) {
    // Parse form data
    if err := r.ParseForm(); err != nil {
        log.Printf("Failed to parse form: %v", err)
        http.Error(w, "Invalid form data", http.StatusBadRequest)
        return
    }

    // Convert form data to GiftRequest
    age, _ := strconv.Atoi(r.Form.Get("age"))
    budget, _ := strconv.ParseFloat(r.Form.Get("budget"), 64)
    interests := strings.Split(r.Form.Get("interests"), ",")

    req := GiftRequest{
        Age:       age,
        Interests: interests,
        Budget:    budget,
    }

    suggestions, err := generateGiftSuggestions(req)
    if err != nil {
        log.Printf("Error generating suggestions: %v", err)
        http.Error(w, "Failed to generate suggestions", http.StatusInternalServerError)
        return
    }

    // Render suggestions as HTML
    w.Header().Set("Content-Type", "text/html")
    tmpl := template.Must(template.New("suggestions").Parse(`
        {{range .Suggestions}}
        <div class="bg-white p-4 rounded-lg shadow mb-4">
            <h3 class="font-bold text-xl">{{.Name}}</h3>
            <p>{{.Description}}</p>
            <p class="text-green-600">Price: ${{printf "%.2f" .Price}}</p>
            <p class="text-gray-500">Category: {{.Category}}</p>
            
            {{if .Images}}
            <div class="grid grid-cols-2 gap-2 mt-2 w-full max-w-full">
                {{range .Images}}
                <div class="aspect-square overflow-hidden rounded">
                  <img src="{{.}}"  class="w-full h-full object-cover object-center hover:scale-110 transition-transform duration-300 ease-in-out">
                </div>
                {{end}}
            </div>
            {{end}}
        </div>
        {{end}}
    `))
    tmpl.Execute(w, suggestions)
}