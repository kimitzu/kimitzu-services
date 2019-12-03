# Kimitzu Services
* Location Service
  * Distance within radius of (lat, lng)
* Voyager Crawler Service
* Search Service

## Documentation
API documentation can be found in [Postman API Docs](https://documenter.getpostman.com/view/7522385/SVtN5CZU?version=latest).

## Todos
* Rewrite the entire voyager part of the application, or at least refactor.
* Potentially get rid of `/models` since the services doesn't need to know the entire structure of the listings/peers.

# Installation
```bash
go get github.com/kimitzu-foundation/kimitzu-services
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

