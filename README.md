# Djali Services
* Location Service
  * Distance within radius of (lat, lng)
* Voyager Crawler Service
* Search Service

## Documentation
API documentation can be found in [Postman API Docs](https://documenter.getpostman.com/view/7522385/SVtN5CZU?version=latest).

# Installation
```bash
go get github.com/djali-foundation/djali-services
```

# Run
```bash
go run services
```

To enabled logging (loglevels 1 to 5)
```bash
go run services.go -log <logLevel>
```


# Testing
```bash
go test ./... -v
```