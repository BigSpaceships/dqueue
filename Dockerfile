FROM docker.io/golang:1.25-alpine AS build

RUN mkdir /opt/dqueue 

WORKDIR /opt/dqueue

RUN apk add git 

COPY go.* .
RUN go mod download 
COPY auth auth
COPY websocket websocket
COPY queue queue
COPY *.go .
RUN go build -v -o dqueue 

FROM docker.io/alpine
WORKDIR /opt/dqueue
RUN apk add --no-cache tzdata 
ENV TZ=America/New_York
RUN cp /usr/share/zoneinfo/America/New_York /etc/localtime 
COPY static static
COPY --from=build /opt/dqueue/dqueue dqueue

ENTRYPOINT [ "./dqueue" ]
