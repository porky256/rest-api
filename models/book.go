package models

type Book struct {
	ID     int     `json:"id"`
	Name   string  `json:"name" binding:"min=1,max=100"`
	Price  float64 `json:"price" binding:"min=0"`
	Genre  int     `json:"genre" binding:"min=1,max=3"`
	Amount int     `json:"amount" binding:"min=0"`
}
