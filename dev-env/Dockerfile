FROM golang:1.16-alpine

WORKDIR /app
COPY src/go.mod ./

RUN go mod download

COPY ../src/storage.go ./
COPY ../cert/* ./cert/
RUN go build -o /storage

EXPOSE 1321
CMD [ "/storage" ]




