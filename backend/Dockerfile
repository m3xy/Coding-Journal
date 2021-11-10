# ----
# backend/Dockerfile: Docker image to run the go backend.
# Author: 190014935
# ----

# ----
# Go docker image
# ---
FROM golang:latest AS cs3099go
RUN mkdir -p gitlab.cs.st-andrews.ac.uk/cs3099group11/project-code/backend
COPY . gitlab.cs.st-andrews.ac.uk/cs3099group11/project-code/backend/
WORKDIR gitlab.cs.st-andrews.ac.uk/cs3099group11/project-code/backend/
RUN go build -o main gitlab.cs.st-andrews.ac.uk/cs3099group11/project-code/backend
CMD ["/go/gitlab.cs.st-andrews.ac.uk/cs3099group11/project-code/backend/main"]
EXPOSE 8080
EXPOSE 3307
