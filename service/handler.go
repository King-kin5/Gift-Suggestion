package service

import (
	"log"
	"net/http"
	"strings"
	"html/template"
	"strconv"
)

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
        <div class="bg-white p-4 rounded-lg shadow">
            <h3 class="font-bold text-xl">{{.Name}}</h3>
            <p>{{.Description}}</p>
            <p class="text-green-600">Price: ${{printf "%.2f" .Price}}</p>
            <p class="text-gray-500">Category: {{.Category}}</p>
        </div>
        {{end}}
    `))
    tmpl.Execute(w, suggestions)
}

