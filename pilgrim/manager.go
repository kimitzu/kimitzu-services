package pilgrim

import (
    "fmt"
    "time"

    "github.com/nokusukun/particles/roggy"

    "github.com/kimitzu/kimitzu-services/servicestore"
)

/*
  TODO:
    * Skip already ingested peers
    * Invervalized ingesting of bootstrap peers
    * Updating already ingested peers
*/

const (
    ListingType = "type/listing"
    PeerType    = "type/peer"
)

var log = roggy.Printer("pilgrim")

type ManagerConfig struct {
    Workers     int
    QueueLength int
}

type Manager struct {
    JobPeerQueue    chan *Job
    JobQueueListing chan *Job
    PeerWorkers     []*Worker
    ListingWorkers  []*Worker
    OriginID        string
    Store           *servicestore.Store
}

func InitializeManager(store *servicestore.Store, config ManagerConfig) *Manager {
    log.Infof("Pilgrim Manager Initialize")
    m := new(Manager)

    if store == nil {
        panic(fmt.Errorf("passed store is nil"))
    }
    m.Store = store

    // Retrieve ID
    origin, err := GetSelfPeerID()
    if err != nil {
        log.Errorf("Failed to retrieve Origin Peer ID: %v", err)
        roggy.Wait()
        panic(err)
    }
    m.OriginID = origin

    // Initialize Queue Buffers
    if config.QueueLength <= 0 {
        m.JobQueueListing = make(chan *Job, 1000)
        m.JobPeerQueue = make(chan *Job, 1000)
        log.Info("Queue buffer size: 1000")
    } else {
        m.JobQueueListing = make(chan *Job, config.QueueLength)
        m.JobPeerQueue = make(chan *Job, config.QueueLength)
        log.Info("Queue buffer size: ", config.QueueLength)
    }

    // Spawn workers
    if config.Workers <= 0 {
        m.SpawnWorker(5)
    } else {
        m.SpawnWorker(config.Workers)
    }

    return m
}

// internal ingest function for processing based on type
func (m *Manager) ingest(target, _type string, queue chan *Job) chan interface{} {
    log.Verbosef("Sent to '%v' queue: %v", _type, target)
    c := make(chan interface{})

    // Send job to the specified queue
    queue <- &Job{
        Target:   target,
        Type:     _type,
        Response: c,
        Manager:  m,
    }
    return c
}

func (m *Manager) IngestPeer(peerID string) chan interface{} {
    retC := make(chan interface{}, 1)
    go func() {
        data := <-m.ingest(peerID, PeerType, m.JobPeerQueue)
        retC <- data
        if data == nil {
            log.Verbosef("%v job returned nil", peerID)
            return
        }

        _, err := m.Store.Peers.Insert(peerID, data)
        if err != nil {
            log.Errorf("Ingest peer '%v' failed: %v", peerID, err)
        }

        log.Debugf("Retrieved data from '%v': %v", peerID, data)
    }()

    return retC
}

func (m *Manager) IngestListing(pidSlug string) chan interface{} {
    retC := make(chan interface{}, 1)
    go func() {
        data := <-m.ingest(pidSlug, ListingType, m.JobQueueListing)
        retC <- data

        if data == nil {
            return
        }

        _, err := m.Store.Listings.Insert(pidSlug, data)
        if err != nil {
            log.Errorf("Ingest peer '%v' failed: %v", pidSlug, err)
        }

        log.Debugf("Retrieved data from '%v': %v", pidSlug, data)
        retC <- data
    }()

    return retC
}

func (m *Manager) StartPilgrimage() {
    log.Info("Starting Pilgrimage")
    for _, w := range m.ListingWorkers {
        w.Start()
    }

    for _, w := range m.PeerWorkers {
        w.Start()
    }

    m.IngestPeer(m.OriginID)

    go func() {
        for {
            pids, err := getNodePeers()
            if err != nil {
                panic(fmt.Errorf("Failed to retrieve peers: %v", err))
            }
            log.Debugf("Sending bootstrapped peers: %v", pids)
            for _, pid := range pids {
                m.IngestPeer(pid)
            }
            time.Sleep(time.Minute * 10)
        }
    }()
}

func (m *Manager) SpawnWorker(count int) {
    log.Infof("Spawning %v workers", count)
    for i := 0; i <= count; i++ {
        w := &Worker{
            Manager:    m,
            Type:       PeerType,
            JobQueue:   m.JobPeerQueue,
            Identifier: fmt.Sprintf("p.work.peer.%v", i),
            stop:       make(chan interface{}),
        }
        m.PeerWorkers = append(m.PeerWorkers, w)
    }

    for i := 0; i <= count; i++ {
        w := &Worker{
            Manager:    m,
            Type:       ListingType,
            JobQueue:   m.JobQueueListing,
            Identifier: fmt.Sprintf("p.work.list.%v", i),
            stop:       make(chan interface{}),
        }
        m.ListingWorkers = append(m.ListingWorkers, w)
    }
}
