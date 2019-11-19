package p2p

type Rating struct {
    Source      string      `json:"src"`
    Destination string      `json:"dst"`
    OrderId     string      `json:"orderid"`
    Signature   []byte      `json:"sig"`
    Content     interface{} `json:"rating"`
}

type RatingRequest struct {
    Identity string `json:"ident"`
}
