FROM golang:1.18-bullseye as builder

ENV CGO_ENABLED=0

WORKDIR /opt
COPY . .

RUN go build .

FROM hashicorp/terraform:1.2.8

ENV WORKING_DIR="/opt"

COPY --from=builder /opt/terracd /bin/

ENTRYPOINT [""]
CMD ["/bin/terracd"]