FROM golang:1.15.5-alpine

#COPY go.mod go.sum ./
#RUN go mod download
COPY main.go ./

ENTRYPOINT ["go", "run", "main.go"]
