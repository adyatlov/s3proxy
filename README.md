# s3proxy
Simple HTTP server which proxies requests to a Amazon Simple Storage Service. It allows legacy HTTP fetchers to download privately shared files from Amazon S3 buckets.

## Deploying
Specify AWS credentials or EC2 Role Provider URL as environment variables and specify a port number as the only parameter:

```bash
AWS_ACCESS_KEY_ID=XXXXXXXXXXXXXXXXXXXX AWS_SECRET_ACCESS_KEY=XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX s3proxy 8080
```

or

```bash
s3proxy 8080
```

which would have the same effect as 

```bash
AWS_METADATA_URL=http://169.254.169.254:80/latest s3proxy 8080
```

## Usage
Use the following format in order to download a file from S3 bucket.

```bash
curl http://localhost:8080/s3-region-name/s3-bucket-name/path/to/your/file.txt
```
## Deploying on DC/OS
Edit `s3proxy.marathon.json`:
Specify credentials with `env.AWS_ACCESS_KEY_ID` and `env.AWS_SECRET_ACCESS_KEY` or `AWS_METADATA_URL`.
Health check `healthChecks.path` 
Use Marathon application descriptor `s3proxy.marathon.json`

