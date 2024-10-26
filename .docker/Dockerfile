FROM golang:1.23.1-alpine AS build

RUN addgroup -g 501 guest \
    && adduser --disabled-password \
        --gecos "" \
        --home "$(pwd)" \
        --uid 501 \
        --ingroup guest \
        --no-create-home \
        guest \
    && cat /etc/passwd | grep guest > /etc/passwd_guest

WORKDIR /app
RUN go env -w GOMODCACHE=/root/.cache/go-build

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build go build -v -o bin/app ./cmd/main.go

FROM golang:1.23.1-alpine AS build-info

ARG GIT_REV
ARG GIT_VERSION
ARG GIT_URL
ARG BUILD_DATE

WORKDIR /app

RUN echo "GIT_COMMIT_HASH=$GIT_REV"  > buildinfo \
	&& echo "GIT_VERSION=$GIT_VERSION" >> buildinfo \ 
	&& echo "GIT_URL=$GIT_URL"         >> buildinfo \
	&& echo "BUILD_DATE=$BUILD_DATE"   >> buildinfo

FROM scratch

WORKDIR /go

COPY --from=build /etc/passwd_guest /etc/passwd
USER guest

COPY --from=build /app/bin/app /go/bin/app
COPY --from=build /app/.env /go/.env
COPY --from=build-info /app/buildinfo /go/buildinfo

CMD ["/go/bin/app"]

