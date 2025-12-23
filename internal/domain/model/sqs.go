package model

type Message struct {
	User     User
	Metadata Metadata
}

type User struct {
	Name    string `json:"name"`
	Age     int32  `json:"age"`
	City    string `json:"city"`
	Country string `json:"country"`
}

type Metadata struct {
	ReceiptHandle string
}
