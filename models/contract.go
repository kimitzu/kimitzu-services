// This file was generated from JSON Schema using quicktype, do not modify it directly.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    contract, err := UnmarshalContract(bytes)
//    bytes, err = contract.Marshal()

package models

import "encoding/json"

func UnmarshalContract(data []byte) (Contract, error) {
    var r Contract
    err := json.Unmarshal(data, &r)
    return r, err
}

func (r *Contract) Marshal() ([]byte, error) {
    return json.Marshal(r)
}

type Contract struct {
    Contract                   ContractClass               `json:"contract"`
    State                      string                      `json:"state"`
    Read                       bool                        `json:"read"`
    Funded                     bool                        `json:"funded"`
    UnreadChatMessages         int64                       `json:"unreadChatMessages"`
    PaymentAddressTransactions []PaymentAddressTransaction `json:"paymentAddressTransactions"`
}

type ContractClass struct {
    VendorListings          []VendorListing          `json:"vendorListings"`
    BuyerOrder              BuyerOrder               `json:"buyerOrder"`
    VendorOrderConfirmation VendorOrderConfirmation  `json:"vendorOrderConfirmation"`
    VendorOrderFulfillment  []VendorOrderFulfillment `json:"vendorOrderFulfillment"`
    BuyerOrderCompletion    BuyerOrderCompletion     `json:"buyerOrderCompletion"`
    Signatures              []Signature              `json:"signatures"`
}

type BuyerOrder struct {
    RefundAddress        string        `json:"refundAddress"`
    RefundFee            int64         `json:"refundFee"`
    Shipping             Shipping      `json:"shipping"`
    BuyerID              RID           `json:"buyerID"`
    Timestamp            string        `json:"timestamp"`
    Items                []ItemElement `json:"items"`
    Payment              Payment       `json:"payment"`
    RatingKeys           []string      `json:"ratingKeys"`
    AlternateContactInfo string        `json:"alternateContactInfo"`
    Version              int64         `json:"version"`
}

type RID struct {
    PeerID     string  `json:"peerID"`
    Handle     string  `json:"handle"`
    Pubkeys    Pubkeys `json:"pubkeys"`
    BitcoinSig string  `json:"bitcoinSig"`
}

type Pubkeys struct {
    Identity string `json:"identity"`
    Bitcoin  string `json:"bitcoin"`
}

type ItemElement struct {
    ListingHash    string         `json:"listingHash"`
    Quantity       int64          `json:"quantity"`
    Quantity64     int64          `json:"quantity64"`
    ShippingOption ShippingOption `json:"shippingOption"`
    Memo           string         `json:"memo"`
    PaymentAddress string         `json:"paymentAddress"`
}

type Payment struct {
    Method       string `json:"method"`
    Moderator    string `json:"moderator"`
    Amount       int64  `json:"amount"`
    Chaincode    string `json:"chaincode"`
    Address      string `json:"address"`
    RedeemScript string `json:"redeemScript"`
    Coin         string `json:"coin"`
}

type Shipping struct {
    ShipTo       string `json:"shipTo"`
    Address      string `json:"address"`
    City         string `json:"city"`
    State        string `json:"state"`
    PostalCode   string `json:"postalCode"`
    Country      string `json:"country"`
    AddressNotes string `json:"addressNotes"`
}

type BuyerOrderCompletion struct {
    OrderID   string   `json:"orderId"`
    Timestamp string   `json:"timestamp"`
    Ratings   []Rating `json:"ratings"`
}

type Rating struct {
    RatingData RatingData `json:"ratingData"`
    Signature  string     `json:"signature"`
}

type RatingData struct {
    RatingKey       string          `json:"ratingKey"`
    VendorID        RID             `json:"vendorID"`
    VendorSig       RatingSignature `json:"vendorSig"`
    BuyerID         RID             `json:"buyerID"`
    BuyerName       string          `json:"buyerName"`
    BuyerSig        string          `json:"buyerSig"`
    Timestamp       string          `json:"timestamp"`
    Overall         int64           `json:"overall"`
    Quality         int64           `json:"quality"`
    Description     int64           `json:"description"`
    DeliverySpeed   int64           `json:"deliverySpeed"`
    CustomerService int64           `json:"customerService"`
    Review          string          `json:"review"`
}

type RatingSignature struct {
    Metadata  RatingSignatureMetadata `json:"metadata"`
    Signature string                  `json:"signature"`
}

type RatingSignatureMetadata struct {
    ListingSlug  string `json:"listingSlug"`
    RatingKey    string `json:"ratingKey"`
    ListingTitle string `json:"listingTitle"`
    Thumbnail    Image  `json:"thumbnail"`
}

type Signature struct {
    Section        string `json:"section"`
    SignatureBytes string `json:"signatureBytes"`
}

type VendorListing struct {
    Slug               string                `json:"slug"`
    VendorID           RID                   `json:"vendorID"`
    Metadata           VendorListingMetadata `json:"metadata"`
    Item               VendorListingItem     `json:"item"`
    ShippingOptions    []interface{}         `json:"shippingOptions"`
    Coupons            []Coupon              `json:"coupons"`
    Moderators         []interface{}         `json:"moderators"`
    TermsAndConditions string                `json:"termsAndConditions"`
    RefundPolicy       string                `json:"refundPolicy"`
    Location           Location              `json:"location"`
}

type Coupon struct {
    Title           string `json:"title"`
    Hash            string `json:"hash"`
    PercentDiscount int64  `json:"percentDiscount"`
}

type VendorListingItem struct {
    Title          string        `json:"title"`
    Description    string        `json:"description"`
    ProcessingTime string        `json:"processingTime"`
    Price          int64         `json:"price"`
    Nsfw           bool          `json:"nsfw"`
    Tags           []string      `json:"tags"`
    Images         []Image       `json:"images"`
    Categories     []string      `json:"categories"`
    Grams          int64         `json:"grams"`
    Condition      string        `json:"condition"`
    Options        []interface{} `json:"options"`
    Skus           []Skus        `json:"skus"`
}

type Location struct {
    Latitude   string `json:"latitude"`
    Longitude  string `json:"longitude"`
    PlusCode   string `json:"plusCode"`
    AddressOne string `json:"addressOne"`
    AddressTwo string `json:"addressTwo"`
    City       string `json:"city"`
    State      string `json:"state"`
    Country    string `json:"country"`
    ZipCode    string `json:"zipCode"`
}

type VendorListingMetadata struct {
    Version               int64    `json:"version"`
    ContractType          string   `json:"contractType"`
    Format                string   `json:"format"`
    Expiry                string   `json:"expiry"`
    AcceptedCurrencies    []string `json:"acceptedCurrencies"`
    PricingCurrency       string   `json:"pricingCurrency"`
    Language              string   `json:"language"`
    EscrowTimeoutHours    int64    `json:"escrowTimeoutHours"`
    CoinType              string   `json:"coinType"`
    CoinDivisibility      int64    `json:"coinDivisibility"`
    PriceModifier         int64    `json:"priceModifier"`
    ServiceRateMethod     string   `json:"serviceRateMethod"`
    ServiceClassification string   `json:"serviceClassification"`
}

type VendorOrderConfirmation struct {
    OrderID         string `json:"orderID"`
    Timestamp       string `json:"timestamp"`
    PaymentAddress  string `json:"paymentAddress"`
    RequestedAmount int64  `json:"requestedAmount"`
}

type VendorOrderFulfillment struct {
    OrderID         string          `json:"orderId"`
    Slug            string          `json:"slug"`
    Timestamp       string          `json:"timestamp"`
    RatingSignature RatingSignature `json:"ratingSignature"`
    Note            string          `json:"note"`
    BuyerRating     BuyerRating     `json:"buyerRating"`
}

type BuyerRating struct {
    Comment   string  `json:"comment"`
    Slug      string  `json:"slug"`
    OrderID   string  `json:"orderId"`
    Fields    []Field `json:"fields"`
    Timestamp string  `json:"timestamp"`
    SourceID  string  `json:"sourceId"`
    TargetID  string  `json:"targetId"`
}

type Field struct {
    Type   string `json:"type"`
    Score  int64  `json:"score"`
    Max    int64  `json:"max"`
    Weight int64  `json:"weight"`
}

type PaymentAddressTransaction struct {
    Txid          string `json:"txid"`
    Value         int64  `json:"value"`
    Confirmations int64  `json:"confirmations"`
    Height        int64  `json:"height"`
    Timestamp     string `json:"timestamp"`
}
