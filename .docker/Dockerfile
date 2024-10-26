FROM golang:1.23.2-alpine AS build

RUN addgroup -g 501 app \
    && adduser --disabled-password \
        --gecos "" \
        --uid 501 \
        --ingroup app \
        --no-create-home \
        app \
    && grep app /etc/passwd > /etc/passwd_app

WORKDIR /app
RUN go env -w GOMODCACHE=/root/.cache/go-build

COPY . .

RUN go mod download \ 
    && go build -v -o bin/app ./cmd/main.go

FROM alpine AS build-info

ARG GIT_REV
ARG GIT_VERSION
ARG GIT_URL
ARG BUILD_DATE

WORKDIR /app

RUN echo "GIT_COMMIT_HASH=$GIT_REV"  > buildinfo && \
    echo "GIT_VERSION=$GIT_VERSION" >> buildinfo && \
    echo "GIT_URL=$GIT_URL"         >> buildinfo && \
    echo "BUILD_DATE=$BUILD_DATE"   >> buildinfo

FROM scratch

WORKDIR /go

COPY --from=build /etc/passwd_app /etc/passwd
USER app

COPY --from=build /app/bin/app /go/bin/app
COPY --from=build-info /app/buildinfo /go/buildinfo

CMD ["/go/bin/app"]

