FROM golang:1.14 AS build

COPY . /src
WORKDIR /src

RUN CGO_ENABLED=0 go build

FROM scratch AS final
COPY --from=build /src/go-spa /go-spa

ENTRYPOINT ["/go-spa"]