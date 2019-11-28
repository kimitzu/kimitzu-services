package p2p

import (
    "fmt"

    "github.com/djali-foundation/djali-services/models"
)

type Rating struct {
    Source        string             `json:"src"`
    SourcePK      models.Pubkeys     `json:"srcpk"`
    Destination   string             `json:"dst"`
    DestinationPK models.Pubkeys     `json:"dstpk"`
    Signatures    []models.Signature `json:"sig"`
    Content       interface{}        `json:"rating"`
}

type RatingRequest struct {
    Identity string `json:"ident"`
}

func VendorRatingFromContract(contract *models.Contract) (rating *Rating, err error) {
    rating = new(Rating)

    if len(contract.Contract.VendorOrderFulfillment) == 0 {
        err = fmt.Errorf("no vendorOrderFulfillment")
        return
    }

    if len(contract.Contract.VendorListings) == 0 {
        err = fmt.Errorf("no vendor listings")
        return
    }

    slug := contract.Contract.VendorOrderFulfillment[0].Slug
    vendorID := contract.Contract.VendorListings[0].VendorID.PeerID

    rating.Destination = fmt.Sprintf("%v@%v", vendorID, slug)
    rating.Source = contract.Contract.BuyerOrder.BuyerID.PeerID

    rating.DestinationPK = contract.Contract.VendorListings[0].VendorID.Pubkeys
    rating.SourcePK = contract.Contract.BuyerOrder.BuyerID.Pubkeys

    rating.Content = contract.Contract.BuyerOrderCompletion
    rating.Signatures = contract.Contract.Signatures

    return
}

func BuyerRatingFromContract(contract *models.Contract) (rating *Rating, err error) {
    rating = new(Rating)

    if len(contract.Contract.VendorOrderFulfillment) == 0 {
        err = fmt.Errorf("no vendorOrderFulfillment")
        return
    }

    if len(contract.Contract.VendorListings) == 0 {
        err = fmt.Errorf("no vendor listings")
        return
    }
    vendorID := contract.Contract.VendorListings[0].VendorID.PeerID

    rating.Destination = contract.Contract.BuyerOrder.BuyerID.PeerID
    rating.Source = vendorID

    rating.DestinationPK = contract.Contract.BuyerOrder.BuyerID.Pubkeys
    rating.SourcePK = contract.Contract.VendorListings[0].VendorID.Pubkeys

    rating.Content = contract.Contract.VendorOrderFulfillment[0]
    rating.Signatures = contract.Contract.Signatures

    return
}
