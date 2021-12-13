package models

type User struct {
	Address     string `json:"address"`
	CreatedAt   string `json:"createdAt"`
	Email       string `json:"email"`
	FirstName   string `json:"firstName"`
	Id          string `json:"Id"`
	LastName    string `json:"lastName"`
	PhoneNumber string `json:"phone"`
	UpdatedAt   string `json:"updatedAt"`
}
