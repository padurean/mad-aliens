# build the go binary
FROM golang:1.18 as builder
WORKDIR /app/mad-aliens
COPY go.mod ./
COPY go.sum ./
COPY world.txt ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /build/invasion cmd/invasion/main.go

# build final alpine image
FROM alpine:latest

# arguments
ARG APP_USER=app
ARG PROJECT_ROOT=/app/
ARG WORLD=world.txt

# install required libs
# RUN apk update && apk --no-cache --update add ca-certificates

# create app dir and user
RUN mkdir -p ${PROJECT_ROOT} && \
    addgroup -g 1000 ${APP_USER} && \
    adduser -u 1000 -D ${APP_USER} -G ${APP_USER}

# set local directory
WORKDIR ${PROJECT_ROOT}

# copy final go binary from the builder stage
COPY --from=builder /build/invasion ${PROJECT_ROOT}invasion
COPY --from=builder /app/mad-aliens/${WORLD} ${PROJECT_ROOT}${WORLD}

# change permissions on our project directory so that our app user has access
RUN chown -R ${APP_USER}:${APP_USER} ${PROJECT_ROOT}

# change to our non root user for security purposes
USER ${APP_USER}

# finally run the binary
ENTRYPOINT ["./invasion"]
