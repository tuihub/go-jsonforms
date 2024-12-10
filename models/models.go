package models

type MenuItem struct {
	Link         string
	ExternalLink string
	Titel        string
	Current      bool
}

type Confirmation struct {
	ButtonText string
	Title      string
	Body       string
	Confirm    string
	Cancel     string
}
