package models

import "github.com/youngjae-lim/golang-fullstack-bnb-website/internal/forms"

// TemplateData holds data sent from the handlers to templates
type TemplateData struct {
	StringMap map[string]string
	IntMap    map[string]int
	FloatMap  map[string]float32
	Data      map[string]interface{}
	CSRFToken string
	Flash     string
	Warning   string
	Error     string
	Form      *forms.Form
}
