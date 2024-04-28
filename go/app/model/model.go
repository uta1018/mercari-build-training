package model

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	ImageName string `json:"image_name"`
}

type Items struct {
	Items []Item `json:"items"`
}
