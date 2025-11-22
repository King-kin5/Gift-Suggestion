package service

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func searchProductImages(productName string, svc *GiftService) ([]string, error) {

	apiKey := svc.GoogleSearchAPIKey
	cx := svc.GoogleSearchCX

	query := url.QueryEscape(productName + "product image")
	apiURL := fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&q=%s&searchType=image&num=3",
		apiKey, cx, query)
	client := &http.Client{}
	req, err := http.NewRequest("GET", apiURL, nil)
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

func SuggestGiftHandler(w http.ResponseWriter, r *http.Request, svc *GiftService) {
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

	suggestions, err := generateGiftSuggestions(req, svc)
	if err != nil {
		log.Printf("Error generating suggestions: %v", err)
		http.Error(w, "Failed to generate suggestions", http.StatusInternalServerError)
		return
	}

	// Render suggestions as HTML
	w.Header().Set("Content-Type", "text/html")
	tmpl := template.Must(template.New("suggestions").Parse(`
        {{range .Suggestions}}
        <div class="bg-white/5 backdrop-blur-md p-6 rounded-xl shadow-lg hover:shadow-gold-500/20 transition-all duration-300 border border-white/10 group">
            <div class="relative overflow-hidden rounded-lg mb-4 aspect-square">
                {{if .Images}}
                <img src="{{index .Images 0}}" class="w-full h-full object-cover transform group-hover:scale-110 transition-transform duration-500">
                {{else}}
                <div class="w-full h-full bg-emerald-800 flex items-center justify-center text-emerald-400">No Image</div>
                {{end}}
                <div class="absolute top-2 right-2 bg-black/70 text-gold-400 px-2 py-1 rounded text-xs backdrop-blur-md border border-gold-500/30">
                    {{.Category}}
                </div>
            </div>
            
            <h3 class="font-serif font-bold text-xl mb-2 text-white group-hover:text-gold-400 transition-colors">{{.Name}}</h3>
            <p class="text-gray-300 text-sm mb-3 line-clamp-2">{{.Description}}</p>
            
            <div class="bg-emerald-900/50 p-3 rounded-lg mb-3 border border-emerald-800">
                <p class="text-xs text-gold-500 font-semibold uppercase tracking-wider mb-1">Why it's a match</p>
                <p class="text-sm text-gray-200 italic">"{{.Reasoning}}"</p>
            </div>

            <div class="flex justify-between items-center pt-2 border-t border-white/10">
                <span class="text-2xl font-bold text-gold-400">${{printf "%.2f" .Price}}</span>
            </div>
        </div>
        {{end}}
    `))
	tmpl.Execute(w, suggestions)
}
