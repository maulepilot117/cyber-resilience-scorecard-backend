FROM golang:1.24.4

# Install wkhtmltopdf
RUN apt-get update && apt-get install -y wkhtmltopdf

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

EXPOSE 3000
CMD ["./main"]