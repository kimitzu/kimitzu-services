package pilgrim

import (
    "encoding/json"
    "fmt"
    "regexp"
    "time"

    "github.com/levigross/grequests"
)

type dictI = map[string]interface{}

// TODO: Finish This
func (j *Job) IngestListing() (interface{}, error) {
    log := log.Sub("JobListing")
    pid, slug, err := PIDSlugUnmarshal(j.Target)

    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal PIDSlug '%v': %v", j.Target, err)
    }

    ld, err := getListingData(pid, slug)
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve listing data '%v': %v", j.Target, err)
    }

    listing, exists := ld["listing"].(dictI)
    if !exists {
        b, _ := json.Marshal(ld)
        return nil, fmt.Errorf("listng '%v' has malformed listing structure: %v",
            j.Target, string(b))
    }

    lType, exists := listing["metadata"].(dictI)["contractType"].(string)
    if !exists {
        b, _ := json.Marshal(ld)
        return nil, fmt.Errorf("listng '%v' has malformed listing structure: %v",
            j.Target, string(b))
    }

    if lType != "SERVICE" {
        return nil, fmt.Errorf("disposing '%v': type not 'SERVICE' instead '%v'", slug, lType)
    }

    images := listing["item"].(dictI)["images"]
    imgbytes, err := json.Marshal(images)
    if err != nil {
        log.Errorf("Failed to marshal images data: %v", err)
    }

    resHashes := scrapeResourcesFromJson(imgbytes)
    if len(resHashes) != 0 {
        go func() {
            for _, resHash := range resHashes {
                err := downloadImage(j.Manager.Store.StorePath, string(resHash))
                if err != nil {
                    log.Errorf("Job '%v' failed to download resource '%v': '%v'",
                        j.Target, string(resHash), err)
                } else {
                    log.Verbosef("Job '%v' saved binary resource: %v",
                        j.Target, string(resHash))
                }
            }
        }()
    }

    listing["hash"] = j.Target

    return listing, nil
}

func getListingData(pid, slug string) (listing map[string]interface{}, err error) {
    d, err := grequests.Get(
        fmt.Sprintf("http://localhost:8100/ob/listing/%v/%v", pid, slug),
        &grequests.RequestOptions{RequestTimeout: 120 * time.Second})

    if err != nil {
        return nil, err
    }

    err = d.JSON(&listing)
    return
}

func scrapeResourcesFromJson(b []byte) [][]byte {
    filter := regexp.MustCompile(`([A-Za-z0-9]){46,49}`)
    data := filter.FindAll(b, -1)
    log.Debugf("Scraping Resources from \n'%v'\n found '%v'.", string(b), len(data))
    return data
}
