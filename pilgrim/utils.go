package pilgrim

import (
    "fmt"
    "io"
    "os"
    "path"
    "path/filepath"
    "strings"
    "time"

    "github.com/levigross/grequests"
)

func getNodePeers() (peers []string, err error) {
    d, err := grequests.Get("http://localhost:8100/ob/peers",
        &grequests.RequestOptions{RequestTimeout: 60 * time.Second})
    if err != nil {
        return
    }

    err = d.JSON(&peers)
    return
}
func GetSelfPeerID() (id string, err error) {
    dat, err := grequests.Get("http://localhost:8100/ob/config/",
        &grequests.RequestOptions{RequestTimeout: 30 * time.Second})
    if err != nil {
        return
    }

    temp := map[string]interface{}{}
    err = dat.JSON(&temp)
    if err != nil {
        return
    }

    id, exists := temp["peerID"].(string)
    if !exists {
        err = fmt.Errorf("config data does not exist: %v", temp)
    }

    return
}

func downloadImage(directory, fileName string) (err error) {
    resPath := path.Join(directory, "images", fileName)
    if doesFileExist(resPath) {
        err = fmt.Errorf("file already exsits: %v", resPath)
        return
    } else {
        ensureDir(resPath)
    }

    file, err := grequests.Get("http://localhost:8100/ipfs/"+fileName,
        &grequests.RequestOptions{RequestTimeout: 60 * time.Second})
    if err != nil {
        err = fmt.Errorf("failed to download resource: %v", err)
        return
    }

    outFile, err := os.Create(resPath)
    if err != nil {
        err = fmt.Errorf("failed to save resource: %v", err)
        return
    }
    defer outFile.Close()

    _, err = io.Copy(outFile, file)
    return
}

func ensureDir(fileName string) {
    dirName := filepath.Dir(fileName)
    if _, err := os.Stat(dirName); err != nil {
        err := os.MkdirAll(dirName, os.ModePerm)
        if err != nil {
            panic(err)
        }
    }
}

func doesFileExist(name string) bool {
    if _, err := os.Stat(name); err != nil {
        if os.IsNotExist(err) {
            return false
        }
    }
    return true
}

func PIDSlugUnmarshal(pidSlug string) (peerID string, slug string, err error) {
    elem := strings.Split(pidSlug, ":")
    if len(elem) != 2 {
        err = fmt.Errorf("'%v' is not a valid PIDSlug", pidSlug)
    }
    peerID = elem[0]
    slug = elem[1]
    return
}

func PIDSlugMarshal(peerID, slug string) string {
    return peerID + ":" + slug
}
