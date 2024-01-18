# Migrate Data from File Share Server to S3

The goal of this project to to provide a fast, reliable and resilent way to migrate unstructural data from a file share server into an object storage like s3.

- Source := Windows File Share
- Destination := Object storage

## Overview

A file share on a Windows Server typically refers to a shared directory or folder on a server that can be accessed by multiple users or client devices over a network. This shared directory can be used to store and manage files, making them accessible to authorized users within the network. The protocols commonly used for file sharing on Windows servers include:

1. Server Message Block (SMB):

***Description***: SMB is a network file sharing protocol that allows applications and users to access files and devices on remote systems over a network. It is the primary protocol used by Windows for file and printer sharing.
***Versions***: Different versions of SMB exist, such as SMBv1, SMBv2, SMBv3, with each version offering improvements in terms of security, performance, and features.
2. Common Internet File System (CIFS):

***Description***: CIFS is the predecessor of SMB and is often used interchangeably with SMB. CIFS is an extension of the Server Message Block (SMB) protocol and provides additional features, including file and printer sharing, authentication, and authorization.
3. Network File System (NFS):

***Description***: While more commonly associated with Unix and Linux systems, Windows servers can also support NFS for file sharing. NFS is a distributed file system protocol that allows a user on a client computer to access files over a network much like local storage.


`s3fs` is a FUSE (Filesystem in Userspace) implementation that allows us to mount an Object Storage like S3 bucket as a local file system on a Linux-based system. It provides a convenient way to interact with our S3 storage using standard file system operations, making it appear as if the S3 bucket is a mounted drive or directory

### Key features of s3fs:

***Mounting S3 Buckets***:

    - s3fs enables you to mount an S3 bucket to a local directory on your Linux system, allowing you to access S3 objects as if they were regular files.
***File System Operations***:

    - Once mounted, you can use standard file system commands (e.g., ls, cp, mv, rm) to interact with S3 objects in the mounted directory.
***FUSE Integration***:

    - s3fs is built on FUSE, which allows non-privileged users to create their own file systems without requiring root access.

### Service Description 

The migration service is a containerized application developed in `Golang` to leverage its inherent parallel processing capabilities through the use of `goroutines`. This service functions as an intermediary to facilitate the migration of data from a source to a designated target destination.

## Set up

1. Mount Windows Shared server into our container using `CIFS v3.0`
From a local linux box, the below command is used to mount a windows file share server into a linux location

```
# mkdir /mnt/smb
# mount -t cifs -o "username={your_windows_username},domain={domain},vers=3.0" //{server_ip_or_hostname}/{shared_folder}/ {destination} -v
example

># mount -t cifs -o "username=jackapaul,domain=IDIR,vers=3.0" //filegish.co.uk/public /mnt/smb -v

```
But do achieve this within a docker container, we use the volume cifs mount option as below:
```
....

volumes:
  court-cifs-volume: 
    driver_opts: 
      type: cifs 
      o: "addr=${SERVER_NAME_OR_IP},username=${USERNAME},password=${PASSWORD},vers=3.0"  
      device: //{SERVER_NAME_OR_IP}/{SHARED_PATH}
```
And then use this volume in out container

```
services:
  virtual-court-container-1:
    image: virtual-court-data-migration:latest
    privileged: true
    profiles:
     - ${CHUNK_FOLDER_1_PROFILE:-"nomigrate"}
    build:
      context: .
      dockerfile: Dockerfile
      args:
        - ACCESS_KEY_ID=${ACCESS_KEY_ID:-0}
        - SECRET_ACCESS_KEY=${SECRET_ACCESS_KEY:-0}
        - BUCKET_NAME=${BUCKET_NAME:-""}
        - S3_ENDPOINT=${S3_ENDPOINT:-""}
    env_file:
      - .env
    environment:
      - SOURCE_PATH=${CHUNK_FOLDER_1:-""}
    volumes:
      - court-cifs-volume:${CIFS_PATH:-""}
```

### Object Store Map using S3FS

To map our object store into our container we make use of `s3fs`

```

echo "${ACCESS_KEY_ID}:${SECRET_ACCESS_KEY}" > /etc/passwd-s3fs
chmod 600 /etc/passwd-s3fs

s3fs ${BUCKET_NAME} {DESTINATION_PATH} -o passwd_file=/etc/passwd-s3fs -o allow_other -o nonempty -o use_path_request_style -o url=${S3_ENDPOINT}
      
```
### Environmental Variable Setup
- Create a .env file on the root of the project with the following

```
ACCESS_KEY_ID={ACCESS_KEY_ID}
SECRET_ACCESS_KEY={SECRET_ACCESS_KEY}
BUCKET_NAME={BUCKET_NAME}
S3_ENDPOINT={S3_ENDPOINT}
USERNAME={USERNAME}
PASSWORD={PASSWORD}
SMB_SERVER={SMB_SERVER}
SMB_SHARE={SMB_SHARE}
DESTINATION_PATH={DESTINATION_PATH} # s3fs container mapped path
CIFS_PATH={CIFS_PATH} # cifs container mapped path
CHUNK_FOLDER_1={CHUNK_FOLDER_1} # The folder you want to migrate e.g. `/Chunk_Jr/TESTING``
CHUNK_FOLDER_1_PROFILE=migrate # migrate folder or not, options are migrate, nomigrate
CHUNK_FOLDER_2={CHUNK_FOLDER_2}
CHUNK_FOLDER_2_PROFILE=nomigrate # do not migrate folder
CHUNK_FOLDER_3=
CHUNK_FOLDER_3_PROFILE=nomigrate 
CHUNK_FOLDER_4=
CHUNK_FOLDER_4_PROFILE=nomigrate 
CHUNK_FOLDER_5=
CHUNK_FOLDER_5_PROFILE=nomigrate 
CHUNK_FOLDER_6=
CHUNK_FOLDER_6_PROFILE=nomigrate 


```
### Running Migration Test

We created a functional test to sustaniably test our code function of migration and check the number of files migrated and content, both should PASS 

```
func TestMigrateFiles(t *testing.T) {
	// Set up a temporary source directory with test files
	sourceDir, err := ioutil.TempDir("", "source")
	tempDir := os.TempDir()
	fmt.Print(tempDir)
	if err != nil {
		t.Fatal("Error creating temporary source directory:", err)
	}
	defer os.RemoveAll(sourceDir)

    ....

	expectedFileCount := 3 // Adjust based on the number of test files created
	if fileCount != expectedFileCount {
		t.Errorf("Expected %d files in destination, got %d", expectedFileCount, fileCount)
	}

	// Test: Check the content of one of the migrated files
	firstDestFileContent, err := readContent(destinationDir, "file1.txt")
	if err != nil {
		t.Fatalf("Error reading content of destination file: %v", err)
	}

	expectedContent := []byte("Content of virtual court file 1")
	if string(firstDestFileContent) != string(expectedContent) {
		t.Errorf("Expected content '%s' in destination file, got '%s'", expectedContent, firstDestFileContent)
	}
}

```

The test run as part of the docker build but to run manually do -

```
go test ./cmd

{"level":"info","time":"2023-12-30T23:52:09-08:00","message":"Migration completed successfully."}
PASS
ok      virtual-records-data-transfer/cmd       0.173s
```