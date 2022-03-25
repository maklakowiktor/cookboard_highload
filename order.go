package main

import "encoding/json"

type Order struct {
	ID           string    `json:"id"`
	Hash         string    `json:"hash"`
	Type         string    `json:"type"`
	OrderName    string    `json:"orderName"`
	Action       string    `json:"action"`
	WaiterID     int       `json:"waiterId"`
	WaiterName   string    `json:"waiterName"`
	TableID      string    `json:"tableId"`
	TerminalID   string    `json:"terminalId"`
	Products     []Product `json:"products"`
	OrderComment string    `json:"orderComment"`
	MsgHash      string    `json:"msgHash"`
}

type Product struct {
	ID           string        `json:"id"`
	Count        int           `json:"count"`
	Name         string        `json:"name"`
	CookingTime  int           `json:"cookingTime"`
	Title        string        `json:"title"`
	TitleArray   []interface{} `json:"titleArray"`
	ProductID    int           `json:"productId"`
	Comment      string        `json:"comment"`
	Modification int           `json:"modification,omitempty"`
}

func UnmarshalOrder(data []byte) (Order, error) {
	var r Order
	err := json.Unmarshal(data, &r)
	return r, err
}

func UnmarshalProduct(data []byte) (Product, error) {
	var p Product
	err := json.Unmarshal(data, &p)
	return p, err
}

func (r *Order) Marshal() ([]byte, error) {
	return json.Marshal(r)
}
