# s3proxy
Simple HTTP server which proxies requests to Amazon Simple Storage Service. It allows legacy HTTP fetchers to download privately shared files from Amazon S3 buckets.

## Usage
Use the following format in order to download a file from S3 bucket:

```bash
$ curl http://localhost:8080/s3-region-name/s3-bucket-name/path/to/your/file.txt
```

## Deployment
Specify AWS credentials or EC2 Role Provider URL as environment variables and specify a port number as the only parameter:

```bash
$ AWS_ACCESS_KEY_ID=XXXX AWS_SECRET_ACCESS_KEY=XXXX s3proxy 8080
```

or

```bash
$ s3proxy 8080
```

which would have the same effect as 

```bash
$ AWS_METADATA_URL=http://169.254.169.254:80/latest s3proxy 8080
```

## Deployment on DC/OS
1. Edit `s3proxy.marathon.json`:
  1. Specify credentials with `env.AWS_ACCESS_KEY_ID` and `env.AWS_SECRET_ACCESS_KEY` or EC2 Role Provider URL with `env.AWS_METADATA_URL` or do not specify `env` at all which would have the same effect as if you set `env.AWS_METADATA_URL` to `http://169.254.169.254:80/latest`.
  2. Specify path to some small file which resides on S3 bucket with `healthChecks.path` in order to set health-check or remove `healthChecks` field if you don't want to have health checks.
  3. If you'd like to use your own build of s3proxy, then change s3proxy executable URI with `uris[0]`. Default URI points to the file built for 64-bit Linux.
2. Deploy s3proxy using Marathon UI or DC/OS CLI:
  
  ```bash
  $ dcos marathon app add s3proxy.marathon.json
  ```

## Build
1. Install Go: https://golang.org/doc/install
2. Run 
  
  ```bash
  $ GOOS=linux GOARCH=amd64 go get github.com/adyatlov/s3proxy
  ```
  
  Omit setting `GOOS` and `GOARCH` if you'd like to build s3proxy for the host platform.
