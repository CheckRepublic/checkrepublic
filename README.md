# checkrepublic
Implementation for the Hackatum Check24 Challenge 

## Build
To build the project, run the following command in the root directory of the project:
```sh
env GOOS=linux GOARCH=amd64 go build -tags=jsoniter .
```
This will use an optimized JSON library for Go, which is faster than the standard library.

## Run
To run the project, execute the following command in the root directory of the project:
```sh
DEBUG=true go run -tags=jsoniter .
```
This will start the server on port 3000. The server will listen for incoming requests on the following endpoints:
- `GET /api/offers`: Returns a list of offers filtered by the query parameters
- `GET /api/offers/all`: Debug endpoint to return all offers
- `POST /api/offers`: Adds a new offers to the list of offers in the database
- `DELETE /api/offers`: Deletes all offers from the database