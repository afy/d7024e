FROM golang:1.23-alpine

COPY . /app

WORKDIR /app

RUN chmod +x ./run_test.sh

ENTRYPOINT ["./run_test.sh"]
