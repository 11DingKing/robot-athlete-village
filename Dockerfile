FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 go build -trimpath -o /out/village ./cmd/server

FROM debian:bookworm-slim
RUN useradd --create-home --uid 10001 village && mkdir -p /data && chown village:village /data
WORKDIR /app
COPY --from=build /out/village /app/village
COPY migrations /app/migrations
USER village
ENV APP_ADDR=:8080 DATABASE_URL=file:/data/village.db
EXPOSE 8080
HEALTHCHECK --interval=5s --timeout=3s CMD /app/village healthcheck
ENTRYPOINT ["/app/village"]
