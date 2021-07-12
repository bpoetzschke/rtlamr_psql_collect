FROM golang:1.16-alpine

ADD ./ /rtlamr_psql_collect

WORKDIR /rtlamr_psql_collect
RUN go build

FROM alpine:latest
COPY --from=0 /rtlamr_psql_collect/rtlamr_psql_collect /app/rtlamr_psql_collect