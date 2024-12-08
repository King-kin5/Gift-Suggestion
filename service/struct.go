package service

type GiftRequest struct{
    Age int `json:"age"`
    Interests []string `json:"interests"`
	Budget    float64  `json:"budget"`
}
type GiftResponse struct {
	Suggestions []Gift `json:"suggestions"`
}

type Gift struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Images      []string `json:"images"`
	
	
}
