FROM golang:1.26-alpine

WORKDIR /app
COPY . .

EXPOSE 8080

# TODO: Add build and startup instructions when the server is implemented.
