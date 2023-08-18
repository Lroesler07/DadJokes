FROM golang as builder
#RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /dadjokes
COPY go.mod ./
RUN go mod download
RUN go mod vendor
RUN go mod tidy
COPY . ./
RUN CGO_ENABLED=0 go build -o dadjokes
# post build stage
FROM alpine
WORKDIR /root
COPY --from=builder /dadjokes .
EXPOSE 8080
CMD ["./dadjokes"]