# Use the official Go image as the base image
FROM golang:latest

# Set arguments for access s3 bucket to mount using s3fs
ARG BUCKET_NAME
ARG ACCESS_KEY_ID
ARG SECRET_ACCESS_KEY
ARG S3_ENDPOINT

# Add environment variables based on arguments
ENV OPERATOR_HOME ${OPERATOR_HOME}
ENV OPERATOR_USER ${OPERATOR_USER}
ENV OPERATOR_UID ${OPERATOR_UID}
ENV BUCKET_NAME ${BUCKET_NAME}
ENV S3_ENDPOINT ${S3_ENDPOINT}
ENV ACCESS_KEY_ID ${ACCESS_KEY_ID}
ENV SECRET_ACCESS_KEY ${SECRET_ACCESS_KEY}

# Install dependency libraries

RUN apt-get update && \
    apt-get install -y \
    iputils-ping \
    telnet \
    cifs-utils \
    smbclient \
    s3fs -y

# setup s3fs configs


# Install cifs-utils to provide CIFS support
WORKDIR /app

# Copy only the necessary files for the Go module

USER root

# Copy the entire project into the container
COPY . .

RUN mkdir /mnt/smb
RUN mkdir /mnt/s3_bucket

RUN echo "${ACCESS_KEY_ID}:${SECRET_ACCESS_KEY}" > /etc/passwd-s3fs
RUN chmod 600 /etc/passwd-s3fs

# map drive
# Create a script to perform the S3 bucket mount and execute Go code
RUN echo "#!/bin/bash\n\
    set -e\n\
    s3fs ${BUCKET_NAME} /mnt/s3_bucket -o passwd_file=/etc/passwd-s3fs -o allow_other -o nonempty -o use_path_request_style -o url=${S3_ENDPOINT} \n\
        if ! mountpoint -q /mnt/s3_bucket; then\n\
        echo 'S3 bucket not mounted'\n\
        exit 1\n\
    fi\n\
    ./app \"$@\"" > mount-and-run.sh

# Make the script executable
RUN chmod +x mount-and-run.sh
# Build the Go application
RUN go build -o app ./cmd

# Run test
RUN go test ./cmd

#USER 1001
# Define the entry point for the container
ENTRYPOINT ["/app/mount-and-run.sh"]