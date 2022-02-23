FROM golang:1.16-bullseye as builder

ENV CGO_ENABLED=0

WORKDIR /opt
COPY . .

RUN go build .

FROM hashicorp/terraform:1.1.6

ENV WORKING_DIR="/opt"

COPY --from=builder /opt/terracd /bin/

ENTRYPOINT [""]
CMD ["/bin/terracd"]