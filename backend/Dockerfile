# ----
# backend/Dockerfile: Docker image to run the go backend.
# Author: 190014935
# ----

# ----
# Go docker image
# ---
FROM golang:latest
RUN mkdir -p /backend
WORKDIR /backend
COPY . /backend
RUN go build -o main .
CMD ["./main"]
EXPOSE 3333
