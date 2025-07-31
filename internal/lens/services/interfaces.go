package services

// CardDatabase provides access to card information
type CardDatabase interface {
	GetCard(id string) (*CardData, bool)
	GetAll() map[string]*CardData
}

// CacheService provides caching functionality for lens computations
type CacheService interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Clear()
	GetStats() map[string]interface{}
}

// LensServices groups all service dependencies for lens functions
type LensServices struct {
	CardDB CardDatabase
	Cache  CacheService
}

// CardData represents card information from the database
type CardData struct {
	UniqueID    string `json:"Unique_ID"`
	Name        string `json:"Name"`
	Title       string `json:"Title"`
	Color       string `json:"Color"`
	Cost        int    `json:"Cost"`
	Inkable     bool   `json:"Inkable"`
	Type        string `json:"Type"`
	Lore        int    `json:"Lore"`
	Willpower   int    `json:"Willpower"`
	Strength    int    `json:"Strength"`
	Image       string `json:"Image"`
	Illustrator string `json:"Illustrator"`
	Language    string `json:"Language"`
	Set         string `json:"Set"`
}

