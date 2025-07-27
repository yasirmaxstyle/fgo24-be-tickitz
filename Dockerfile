FROM golang:alpine as build

WORKDIR /buildapp

COPY . .

RUN go build -o noir main.go

FROM alpine:3.22

WORKDIR /app

COPY --from=build /buildapp/noir /app/noir

ENTRYPOINT [ "/app/noir" ]