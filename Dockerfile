FROM golang:1.16-buster

ADD ./ /rtlamr_psql_collect

WORKDIR /rtlamr_psql_collect
RUN go build
RUN go get github.com/bemasher/rtlamr

FROM debian:buster-20210621

ARG DEBUG
ARG DB_HOST
ARG DB_PORT
ARG DB_USER
ARG DB_PASSWORD
ARG DB_DATABASE
ARG RTLAMR_FILTERID

COPY --from=0 /rtlamr_psql_collect/rtlamr_psql_collect /app/rtlamr_psql_collect
COPY --from=0 /go/bin/rtlamr /usr/local/bin/rtlamr
RUN apt update && apt install -y rtl-sdr && apt autoremove --purge
CMD /app/rtlamr_psql_collect