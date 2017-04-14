# Dockerfile for gRPC Go
FROM golang:1.7

# Set timezone
RUN echo "Asia/Shanghai" > /etc/timezone
RUN dpkg-reconfigure -f noninteractive tzdata

# Go dependencies
RUN go get github.com/satori/go.uuid
RUN go get github.com/go-redis/redis
RUN go get github.com/Sirupsen/logrus

# Build
ENV APP_SRC $GOPATH/src/git.youplus.cc/magic/Mataq
COPY . $APP_SRC
RUN cd $APP_SRC && go build -v -o $GOPATH/bin/mataqd

WORKDIR $APP_SRC
ENTRYPOINT ["mataqd"]
