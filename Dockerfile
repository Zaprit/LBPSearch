FROM golang:latest AS lbpsearchbuild
LABEL authors="zaprit"

RUN mkdir /build
COPY . /build
WORKDIR /build

RUN go build -o /lbpsearch ./cmd/lbpsearch
RUN chmod 755 lbpsearch

FROM rust:latest AS archivedlbuild
WORKDIR /
COPY ./deps/archive_dl /lbp_archive_dl
WORKDIR /lbp_archive_dl
RUN cargo build --release

FROM debian AS final
COPY --from=lbpsearchbuild /lbpsearch /lbpsearch
COPY --from=archivedlbuild /lbp_archive_dl/target/release/archive_dl /archive_dl

ENV LBPSEARCH_ARCHIVE_DL_COMMAND_PATH="/archive_dl"
ENV LBPSEARCH_CACHE_PATH="/data"
WORKDIR /data
ENTRYPOINT ["/lbpsearch"]