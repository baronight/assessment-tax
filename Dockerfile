# build stage
FROM golang:1.22.2-alpine3.19 AS builder
WORKDIR /app
COPY . ./
RUN go mod download
RUN CGO_ENABLED=0 go test -v
RUN go build -v -o /bin/app

# deploy stage
FROM alpine:3.19
COPY --from=builder /bin/app /bin
EXPOSE 8080
CMD [ "/bin/app" ]
