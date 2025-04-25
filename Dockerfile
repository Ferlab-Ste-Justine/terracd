FROM golang:1.23-bullseye as builder

ENV CGO_ENABLED=0

WORKDIR /opt
COPY . .

RUN go build .

FROM hashicorp/terraform:1.4.6

ENV WORKING_DIR="/opt"

COPY --from=builder /opt/terracd /bin/

ENTRYPOINT [""]
CMD ["/bin/terracd"]