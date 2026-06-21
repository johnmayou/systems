FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 go build -o /bin/service ./cmd/counter

FROM scratch
COPY --from=build /bin/service /bin/service
ENTRYPOINT ["/bin/service"]
