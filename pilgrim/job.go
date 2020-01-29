package pilgrim

import (
    "fmt"
)

const ListingEndpoint = ""
const PeerEndpoint = ""

type Job struct {
    Target   string
    Type     string
    Response chan interface{}
    Manager  *Manager
}

func (j *Job) Execute() (result interface{}, err error) {
    log := log.Sub("Job")
    switch j.Type {
    case PeerType:
        result, err = j.IngestPeer()
    case ListingType:
        result, err = j.IngestListing()
    default:
        log.Errorf("Unknown type '%v' in target '%v'", j.Type, j.Target)
    }

    if err != nil {
        return result, fmt.Errorf("failed to run job for target '%v': %v", j.Target, err)
    }

    return result, nil
}
