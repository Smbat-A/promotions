FROM golang:latest

WORKDIR /app

RUN apt-get update && apt-get install -y git

RUN git clone https://github.com/Smbat-A/promotions.git

# Set the Current Working Directory inside the container
WORKDIR /app/promotions

# Build the Go app
RUN go mod download && go build -o app/promotions/main

EXPOSE 8080

CMD ["app/promotions/main"]
