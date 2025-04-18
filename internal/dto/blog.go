package dto

type BlogUpdateDTO struct {
	Id       string
	Title    *string   `json:"title"`
	Category *string   `json:"category"`
	Content  *string   `json:"content"`
	Tags     *[]string `json:"tags"`
}

type BlogCreateDto struct {
	Title    string   `json:"title"`
	Category string   `json:"category"`
	Content  string   `json:"content"`
	Tags     []string `json:"tags"`
}
