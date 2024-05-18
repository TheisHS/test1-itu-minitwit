# Downloads base-image from dockerhub which include necessary dependencies for GoLang.
# https://hub.docker.com/_/golang
FROM golang:1.21

# Environment variable to enable C code in Go application
# ENV CGO_ENABLED=1
ENV SESSION_KEY="bb9cfb7ab2a6e36d683b0b209f96bb33"

# WORKDIR sets the directory (or creates it) in the docker image
WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change

# COPY will copy files/directories and add them to the filesystem of the container
COPY src/go.mod src/go.sum ./
# RUN executes commands and builds a new layer on top of the current image.
# So, download dependencies and check the dependencies of the main module.
RUN go mod download && go mod verify

# Copies the whole minitwit directory to the container
COPY src ./
COPY schema.sql .env* ./

# Compiles code into executable
# -v prints names of packages as they are compiled
# -o /usr/local/bin/app specifies output file of the executable. So the binary is called app and placed in bin/app within the image
# ./... tells go build to build all Go packages starting from the current directory
RUN go build -v -o /usr/local/bin/app ./...

# CMD sets the instruction to be executed when running the container from the image
CMD ["app"]

# To build and run the container from the image defined in the Dockerfile, run:
# $ docker compose up --build
# The --build flag builds new image before starting container, to make sure we run the newest version.
# port 5000:5000 is defined in the compose.yaml and the app should be running at http://localhost:5000/
# OBS. for windows: make sure docker desktop is running, otherwise the docker daemon is not running and you cannot connect.