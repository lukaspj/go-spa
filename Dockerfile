FROM golang:1.14 AS build

COPY backend /src
WORKDIR /src

RUN CGO_ENABLED=0 go build server.go

FROM scratch AS final
COPY --from=build /src/server /server

ENTRYPOINT ["/server"]