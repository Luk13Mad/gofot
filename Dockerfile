FROM golang:1.19.2

RUN mkdir -p /app

ADD . /app

WORKDIR /app

RUN go build -o main ./main.go

EXPOSE 8080
CMD [ "/app/main" ]