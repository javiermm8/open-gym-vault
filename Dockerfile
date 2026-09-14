# --- build stage -----------------------------------------------------------
FROM golang:1.27-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /bin/opengymvault ./cmd/opengymvault && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /bin/migrate ./cmd/migrate

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /bin/opengymvault /opengymvault
COPY --from=build /bin/migrate /migrate

EXPOSE 8080
USER nonroot
ENTRYPOINT ["/opengymvault"]
