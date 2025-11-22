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
)

func generateGiftSuggestions(req GiftRequest, svc *GiftService) (GiftResponse, error) {
	// First, try Gemini API
	suggestions, err := tryGeminiSuggestions(req, svc)
	if err != nil {
		log.Printf("Gemini API error: %v. Falling back to default suggestions.", err)
		suggestions = generateFallbackSuggestions(req)
	}
	for i := range suggestions {
		images, err := searchProductImages(suggestions[i].Name, svc)
		if err != nil {
			log.Printf("Failed to find images for %s: %v", suggestions[i].Name, err)
			continue
		}
		if len(images) > 2 {
			images = images[:2]
		}
		suggestions[i].Images = images
	}

	return GiftResponse{Suggestions: suggestions}, nil
}

func tryGeminiSuggestions(req GiftRequest, svc *GiftService) ([]Gift, error) {
	// Construct prompt
	prompt := fmt.Sprintf(`You are a helpful gift finder. Generate thoughtful gift suggestions based on these criteria:
  - Age: %d years old
  - Interests: %v
  - Budget: $%.2f 

  For each suggestion, provide in this exact format:
  1. Gift Name: [name]
   Description: [description]
   Price: $[price]
   Category: [category]
   Reasoning: [Why is this a good match?]
   Estimated Price: [price range or specific price]

  Provide exactly 5 unique, creative, and high-quality gift suggestions that match the criteria. Avoid generic items unless they are premium versions.`,
		req.Age, strings.Join(req.Interests, ", "), req.Budget)

	// Generate suggestions using Gemini
	model := svc.GenAIClient.GenerativeModel("gemini-2.5-flash")
	resp, err := model.GenerateContent(context.Background(), genai.Text(prompt))
	if err != nil {
		log.Printf("Gemini API Call Failed: %v", err)
		return nil, err
	}

	// Check if we have candidates
	if len(resp.Candidates) == 0 {
		log.Println("Gemini returned no candidates")
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

	log.Printf("Gemini Raw Response:\n%s", responseText)

	// Parse Gemini response and convert to Gift suggestions
	suggestions := parseGiftSuggestions(responseText)
	log.Printf("Parsed %d suggestions", len(suggestions))

	// If we got no suggestions, fallback to default
	if len(suggestions) == 0 {
		log.Println("Parsing failed, triggering fallback")
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

	for i := 0; i < 5; i++ {
		// Base price calculation
		basePrice := req.Budget / 2
		priceVariation := basePrice * 0.5
		price := basePrice + (rand.Float64()*2-1)*priceVariation

		// Try to match interests if possible
		category := giftCategories[rand.Intn(len(giftCategories))]
		if len(interests) > 0 {
			interestCategory := strings.Title(strings.ToLower(interests[rand.Intn(len(interests))]))
			if contains(giftCategories, interestCategory) {
				category = interestCategory
			}
		}

		suggestion := Gift{
			Name:           fmt.Sprintf("Premium Gift %d for %d-year-old", i+1, req.Age),
			Description:    fmt.Sprintf("A curated, thoughtful gift in the %s category", category),
			Price:          price,
			Category:       category,
			Reasoning:      "Matches the user's interests and budget perfectly.",
			EstimatedPrice: fmt.Sprintf("$%.2f", price),
		}
		suggestions = append(suggestions, suggestion)
	}

	return suggestions
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func parseGiftSuggestions(responseText string) []Gift {
	suggestions := []Gift{}

	// Use a more flexible regex to match gift blocks
	giftRegex := regexp.MustCompile(`(?s)Gift\s*Name:\s*(.+?)\n.*?Description:\s*(.+?)\n.*?Price:\s*\$?(\d+(?:\.\d{1,2})?)\n.*?Category:\s*(.+?)\n.*?Reasoning:\s*(.+?)\n.*?Estimated\s*Price:\s*(.+?)(?:\n|$)`)

	matches := giftRegex.FindAllStringSubmatch(responseText, -1)

	for _, match := range matches {
		// Ensure we have all the expected submatches
		if len(match) >= 7 {
			price, err := strconv.ParseFloat(match[3], 64)
			if err != nil {
				log.Printf("Error parsing price: %v", err)
				continue
			}
			suggestions = append(suggestions, Gift{
				Name:           strings.TrimSpace(match[1]),
				Description:    strings.TrimSpace(match[2]),
				Price:          price,
				Category:       strings.TrimSpace(match[4]),
				Reasoning:      strings.TrimSpace(match[5]),
				EstimatedPrice: strings.TrimSpace(match[6]),
			})
		}
	}

	return suggestions
}
