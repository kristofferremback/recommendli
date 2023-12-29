FROM golang:1.21-alpine3.18 as build
WORKDIR /src
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main -v main.go

FROM gcr.io/distroless/static:nonroot
COPY --from=build /src/main /main
ENTRYPOINT ["/main"]
