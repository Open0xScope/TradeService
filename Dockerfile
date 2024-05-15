# Start from golang v1.11 base image
FROM golang:1.22-alpine as builder

RUN echo "http://dl-3.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories &&\
    apk add --no-cache ca-certificates bash gcc g++ linux-headers

WORKDIR /work

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

# Force the go compiler to use modules
# ENV GO111MODULE=on

# ENV GOPROXY=https://goproxy.io

#This is the ‘magic’ step that will download all the dependencies that are specified in
# the go.mod and go.sum file.
# Because of how the layer caching system works in Docker, the  go mod download
# command will _ only_ be re-run when the go.mod or go.sum file change
# (or when we add another docker instruction this line)
RUN go mod download

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o gosdk ./cmd

######## Start a new stage from scratch #######
FROM alpine:latest

RUN apk --no-cache add bash  ca-certificates tree

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /work/gosdk ./

# Document that the service listens on port 24000.
EXPOSE 8000

CMD ["./gosdk"]
