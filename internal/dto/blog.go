package dto

type BlogUpdateDTO struct {
	Id       string    `validate:"required"`
	Title    *string   `json:"title"`
	Category *string   `json:"category"`
	Content  *string   `json:"content"`
	Tags     *[]string `json:"tags"`
}

type BlogCreateDto struct {
	Title    string   `json:"title" validate:"required"`
	Category string   `json:"category" validate:"required"`
	Content  string   `json:"content" validate:"required"`
	Tags     []string `json:"tags" validate:"required"`
}

type BlogDeleteDto struct {
	Id string `validate:"required"`
}

type BlogGetDto struct {
	Id string `validate:"required"`
}
