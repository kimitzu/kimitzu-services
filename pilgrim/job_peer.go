package pilgrim

import (
    "fmt"
    "time"

    "github.com/levigross/grequests"
)

func (j *Job) IngestPeer() (interface{}, error) {
    log := log.Sub("JobPeer")
    peerID := j.Target

    // Get Profile Data
    // Terminate early if there's no profile data
    profile, err := getProfileData(peerID)
    if err != nil {
        log.Errorf("Failed to get profile data for '%v': %v", peerID, err)
        log.Verbosef("Trying to retrieve nearest peer for '%v'", peerID)

        // Get Closest Peers
        nearest, err := getNearestPeers(peerID)
        if err != nil {
            log.Errorf("Failed to get nearest peers for '%v': %v", peerID, err)
        }

        // Ingest closest peers
        for _, pid := range nearest {
            j.Manager.IngestPeer(pid)
        }

        return nil, err
    }

    // Get Listings
    slugs, err := getPeerListingSlugs(peerID)

    for _, slug := range slugs {
        j.Manager.IngestListing(PIDSlugMarshal(peerID, slug))
    }

    // Get Closest Peers
    nearest, err := getNearestPeers(peerID)
    if err != nil {
        log.Errorf("Failed to get nearest peers for '%v': %v", peerID, err)
    }

    // Ingest closest peers
    for _, pid := range nearest {
        j.Manager.IngestPeer(pid)
    }

    return profile, nil
}

func getProfileData(peerID string) (profile interface{}, err error) {
    d, err := grequests.Get(
        fmt.Sprintf("http://localhost:8100/ob/profile/%v", peerID),
        &grequests.RequestOptions{RequestTimeout: 120 * time.Second})
    if err != nil {
        return
    }

    temp := map[string]interface{}{}
    err = d.JSON(&temp)
    if err != nil {
        return
    }

    _, exists := temp["peerID"]
    if !exists {
        err = fmt.Errorf("profile not found: %v", temp)
    }

    profile = temp
    return
}

func getPeerListingSlugs(peerID string) (slugs []string, err error) {
    // http://localhost:8100/ob/listings/QmRFnERGuQwvA5WCF8gL6acUG9mCiugB5CeQEmJ2rd7RfG
    d, err := grequests.Get(
        fmt.Sprintf("http://localhost:8100/ob/listings/%v", peerID),
        &grequests.RequestOptions{RequestTimeout: 120 * time.Second})
    if err != nil {
        return
    }

    temp := new([]map[string]interface{})
    err = d.JSON(&temp)
    if err != nil {
        return
    }

    for _, item := range *temp {
        slug, exist := item["slug"].(string)
        if !exist {
            log.Errorf("Failed to ingest item from source '%v', does not have slug: %v", peerID, item)
            continue
        }

        // Checks if the contractType is service, since kimitzu only caters to SERVICE type contracts
        contractType, exist := item["contractType"].(string)
        if !exist {
            log.Errorf("Failed to ingest item from source '%v', does not have contractType: %v", peerID, item)
            continue
        }
        if contractType == "SERVICE" {
            slugs = append(slugs, slug)
        }
    }

    return
}

func getNearestPeers(peerID string) (peers []string, err error) {
    d, err := grequests.Get(
        fmt.Sprintf("http://localhost:8100/ob/closestpeers/%v", peerID),
        &grequests.RequestOptions{RequestTimeout: 120 * time.Second})
    if err != nil {
        return
    }

    err = d.JSON(&peers)
    return
}
