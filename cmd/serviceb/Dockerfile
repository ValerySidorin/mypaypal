FROM golang:1.20-alpine AS build

WORKDIR /app

COPY ../../go.mod ../../go.sum ./

RUN go mod download

COPY . .

RUN go build -o /serviceb ./cmd/serviceb

FROM gcr.io/distroless/static-debian11

WORKDIR /

COPY --from=build /serviceb /serviceb

ENTRYPOINT [ "/serviceb"]