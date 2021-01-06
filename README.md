# This project is discontinued.

# Kimitzu Services

This repository contains the API layer used by the [Kimitzu Client](https://github.com/kimitzu/kimitzu-client).

# Features

* Voyager Crawler Service
  * Indexing
  * Caching
* Search Service
  * Keyword Search
  * Advanced Filtering
    * By Listing
    * By Profile
  * Location Service
    * Distance within radius of (lat, lng)
    * Zip Code
* S/Kademlia P2P Service for Decentralized Ratings
  * P2P rendezvous via S/Kademlia DHT network

# Documentation
API documentation can be found in [Postman API Docs](https://documenter.getpostman.com/view/7522385/SVtN5CZU?version=latest).

(Documentation in progress.)

# Prerequisites
- Go Version 1.12 or higher
- Packr2: `go get -u github.com/gobuffalo/packr/v2/packr2`

# Building
```bash
go get -u -v github.com/kimitzu-foundation/kimitzu-services
packr2
go build services.go
```

`Packr2` is used to package some external dependencies (location maps) for a single and compact build file.

# Run
## For Windows
```bash
./services.exe
```

To enable logging (loglevels 1 to 5)
```bash
./services.exe --log <logLevel>
```

## For Unix Systems
```bash
./services
```

To enable logging (loglevels 1 to 5)
```bash
./services --log <logLevel>
```

# License

[MPL-2.0](LICENSE).
