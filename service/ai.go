package service

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func generateGiftSuggestions(req GiftRequest) (GiftResponse, error) {
	// First, try Gemini API
	suggestions, err := tryGeminiSuggestions(req)
	if err != nil {
		log.Printf("Gemini API error: %v. Falling back to default suggestions.", err)
		suggestions = generateFallbackSuggestions(req)
	}
	for i := range suggestions {
		images, err := searchProductImages(suggestions[i].Name)
		if err != nil {
			log.Printf("Failed to find images for %s: %v", suggestions[i].Name, err)
			continue
		}
		if len(images)>2{
			images=images[:2]
		}
		suggestions[i].Images = images
	}

	return GiftResponse{Suggestions: suggestions}, nil
}

func tryGeminiSuggestions(req GiftRequest) ([]Gift, error) {
	// Set up Gemini API client 
	ctx := context.Background() 
	client, err := genai.NewClient(ctx, option.WithAPIKey("AIzaSyBEX-kVe-mS2sasfPKKnRvrH2xGq0z9-6E")) 
	if err != nil { 
		return nil, err 
	} 
	defer client.Close()

	// Construct prompt 
	prompt := fmt.Sprintf(`Generate gift suggestions based on these criteria:
  - Age: %d years old
  - Interests: %v
  - Budget: $%.2f 

  For each suggestion, provide in this exact format:
  1. Gift Name: [name]
   Description: [description]
   Price: $[price]
   Category: [category]

  Provide exactly 10 unique gift suggestions that match the criteria.`,  
	req.Age, strings.Join(req.Interests, ", "), req.Budget)

	// Generate suggestions using Gemini 
	model := client.GenerativeModel("gemini-pro") 
	resp, err := model.GenerateContent(ctx, genai.Text(prompt)) 
	if err != nil { 
		return nil, err 
	}

	// Check if we have candidates
	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates returned from Gemini")
	}

	// Extract text from the first candidate
	var responseText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			responseText = string(text)
			break
		}
	}

	

	// Parse Gemini response and convert to Gift suggestions 
	suggestions := parseGiftSuggestions(responseText)

	// If we got no suggestions, fallback to default
	if len(suggestions) == 0 {
		suggestions = generateFallbackSuggestions(req)
	}

	return suggestions, nil
}

func generateFallbackSuggestions(req GiftRequest) []Gift {
	// Predefined gift suggestions with some variability based on input
	giftCategories := []string{"Electronics", "Books", "Gadgets", "Hobby", "Personal Care", "Experience", "Fashion", "Home & Kitchen"}
	interests := req.Interests

	suggestions := []Gift{}
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 10; i++ {
		// Base price calculation
		basePrice := req.Budget / 2
		priceVariation := basePrice * 0.5
		price := basePrice + rand.Float64() * priceVariation

		// Try to match interests if possible
		category := giftCategories[rand.Intn(len(giftCategories))]
		if len(interests) > 0 {
			category = strings.Title(strings.ToLower(interests[rand.Intn(len(interests))]))
		}

		suggestion := Gift{
			Name:        fmt.Sprintf("Gift %d for %d-year-old", i+1, req.Age),
			Description: fmt.Sprintf("A thoughtful gift in the %s category", category),
			Price:       price,
			Category:    category,
		}
		suggestions = append(suggestions, suggestion)
	}

	return suggestions
}
func parseGiftSuggestions(responseText string) []Gift {
	suggestions := []Gift{}
	
	// Use a more flexible regex to match gift blocks
	giftRegex := regexp.MustCompile(`(?:^|\n)(\d+\.)?(?:\s*)?Gift\s*Name:\s*(.+)\n*(?:\s*)?Description:\s*(.+)\n*(?:\s*)?Price:\s*\$?(\d+(?:\.\d{1,2})?)\n*(?:\s*)?Category:\s*(.+)`)
	
	matches := giftRegex.FindAllStringSubmatch(responseText, -1)
	
	for _, match := range matches {
		// Ensure we have all the expected submatches
		if len(match) >= 6 {
			price, err := strconv.ParseFloat(match[4], 64)
			if err != nil {
				log.Printf("Error parsing price: %v", err)
				continue
			}
			
			suggestion := Gift{
				Name:        strings.TrimSpace(match[2]),
				Description: strings.TrimSpace(match[3]),
				Price:       price,
				Category:    strings.TrimSpace(match[5]),
			}
			suggestions = append(suggestions, suggestion)
		}
	}

	// If no suggestions found, log the full response for debugging
	if len(suggestions) == 0 {
		log.Printf("No suggestions parsed. Full response: %s", responseText)
	}

	return suggestions
}
