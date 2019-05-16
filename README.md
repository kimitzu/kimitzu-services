# Djali Services
* Location Service
  * Distance within radius of (lat, lng)
* Voyager Crawler Service
* Search Service

# Installation
```bash
go get gitlab.com/kingsland-team-ph/djali/djali-services.git
```

# Run
```bash
go run bakedflood.go
```

To enabled logging (loglevels 1 to 5)
```bash
go run bakedflood.go -log <logLevel>
```

# Testing
```bash
go test ./... -v
```