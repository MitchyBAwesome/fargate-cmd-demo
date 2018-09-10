# STEP 1 build executable binary

FROM golang:alpine As builder 
COPY . $GOPATH/src/github.com/mitchybawesome/http-server/
WORKDIR $GOPATH/src/github.com/mitchybawesome/http-server/

RUN apk add --no-cache git mercurial

#get dependancies
RUN go get -d -v

#build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/httpserver

# STEP 2 build a small image

# start from scratch
FROM scratch

# Copy our static executable from the builder
COPY --from=builder /go/bin/httpserver /go/bin/httpserver
ENTRYPOINT ["/go/bin/httpserver"]