# Websocket chat

This is a Go implementation of a chat using Google cloud's datastore and pubsub.

The chat only supports direct messages and doesn't send notifications for new subscribed users.

A VERY basic client interface is available for testing. To test it build and run the project as explained below, 
then open the interface in two different browsers, signup with two different accounts and then reload the page.

The interface has a small layont bug on Firefox because of the _flexbox_ layout.

## Running the project
To run the project you need the last stable docker engine

### Build
This step is needed to compile _wschat_
```bash
docker-compose build
```

### Running
```bash
docker-compose up -d wschat
open localhost:8080
```

## Developing
### Updating the local environment
In order to update the dependencies after a branch switch or update, run the following task
```bash
docker-compose run --rm glide install
```

### Testing
```bash
docker-compose run --rm test
```
#### Testing a specific package
```bash
docker-compose run --rm -e PKG=./packageName test
```

#### Verbose test output
```bash
docker-compose run --rm test -v
```

### Coverage tests
```bash
docker-compose run --rm cover && open /tmp/cover-results/coverage.html
```

### Fmt (linting)
```bash
docker-compose run --rm fmt
```

### Run a go command
```bash
docker-compose run --rm go <arguments>
``` 

### Glide (package manager)
```bash
docker-compose run --rm glide <arguments>
``` 

### Godoc
```bash
docker-compose up -d godoc
open http://localhost:6060
``` 
