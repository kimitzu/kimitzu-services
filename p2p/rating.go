package p2p

import (
    "fmt"
    "os"

    "github.com/boltdb/bolt"

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

func InitializeRatingManager(path string) (man *RatingManager, err error) {
    man = new(RatingManager)
    db, err := bolt.Open(path, os.ModePerm, nil)
    if err != nil {
        return
    }

    // Initialize bucket for ratings
    err = db.Update(func(tx *bolt.Tx) (err error) {
        _, err = tx.CreateBucketIfNotExists([]byte("ratings"))
        return
    })
    if err != nil {
        return
    }

    man.db = db
    return
}

type RatingManager struct {
    db *bolt.DB
}

func makeId(source, dest string) []byte {
    return []byte(fmt.Sprint(source, dest))
}

func (rm *RatingManager) InsertRating(rating *Rating) (err error) {
    return rm.db.Update(func(tx *bolt.Tx) error {
        b, err := tx.CreateBucketIfNotExists([]byte("ratings"))
        if err != nil {
            return err
        }

        bRat, err := json.Marshal(rating)
        if err != nil {
            return err
        }

        return b.Put(makeId(rating.Source, rating.Destination), bRat)
    })
}

func (rm *RatingManager) Close() error {
    return rm.db.Close()
}

func (rm *RatingManager) IngestCompletionRating(contract *models.Contract) (err error) {
    rating, err := VendorRatingFromContract(contract)
    if err != nil {
        return
    }

    err = rm.InsertRating(rating)
    if err != nil {
        return
    }

    return
}

func (rm *RatingManager) IngestFulfillmentRating(contract *models.Contract) (err error) {
    rating, err := BuyerRatingFromContract(contract)
    if err != nil {
        return
    }

    err = rm.InsertRating(rating)
    if err != nil {
        return
    }

    return
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
