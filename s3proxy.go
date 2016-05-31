package main

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/valyala/fasthttp"
	"io"
	"log"
	"os"
	"strings"
)

func parseFullPath(fullPath string) (region, bucket, path string, err error) {
	args := strings.SplitN(strings.TrimPrefix(fullPath, "/"), "/", 3)
	if len(args) < 3 {
		return "", "", "", errors.New("Malformed path")
	}
	region = args[0]
	bucket = args[1]
	path = args[2]
	if region == "" || bucket == "" || path == "" {
		return "", "", "",
			errors.New("Region, bucket and path should not be empty")
	}
	return
}

func getAWSConfig(region string) *aws.Config {
	conf := &aws.Config{}
	// Grab the metadata URL
	metadataURL := os.Getenv("AWS_METADATA_URL")
	if metadataURL == "" {
		metadataURL = "http://169.254.169.254:80/latest"
	}

	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{Filename: "", Profile: ""},
			&ec2rolecreds.EC2RoleProvider{
				Client: ec2metadata.New(session.New(&aws.Config{
					Endpoint: aws.String(metadataURL),
				})),
			},
		})

	conf.Credentials = creds
	if region != "" {
		conf.Region = aws.String(region)
	}

	return conf
}

func serve(ctx *fasthttp.RequestCtx) {
	fullPath := string(ctx.Path())
	log.Println("Downloading", fullPath)
	region, bucket, path, err := parseFullPath(fullPath)
	if err != nil {
		log.Println("Cannot parse URL:", err)
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}
	config := getAWSConfig(region)
	sess := session.New(config)
	client := s3.New(sess)
	req := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	}
	s3resp, err := client.GetObject(req)
	if err != nil {
		log.Println("Cannot GetObject:", err)
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}
	defer s3resp.Body.Close()

	if s3resp.ContentDisposition != nil {
		ctx.Response.Header.Set("Content-Disposition", *s3resp.ContentDisposition)
	}
	if s3resp.ContentEncoding != nil {
		ctx.Response.Header.Set("Content-Encoding", *s3resp.ContentEncoding)
	}
	if s3resp.ContentLanguage != nil {
		ctx.Response.Header.Set("Content-Language", *s3resp.ContentLanguage)
	}
	if s3resp.ContentLength != nil {
		ctx.Response.Header.SetContentLength(int(*s3resp.ContentLength))
	}
	if s3resp.ContentRange != nil {
		ctx.Response.Header.Set("Content-Range", *s3resp.ContentRange)
	}
	if s3resp.ContentType != nil {
		ctx.Response.Header.Set("Content-Type", *s3resp.ContentType)
	}

	nBytes, err := io.Copy(ctx.Response.BodyWriter(), s3resp.Body)
	if err != nil {
		log.Println("Error:", err)
		ctx.Response.Reset()
		ctx.SetStatusCode(fasthttp.StatusBadGateway)
		ctx.SetContentType("text/plain; charset=utf8")
		fmt.Fprintln(ctx, "Error:", err)
		return
	}
	log.Printf("%v: %v bytes are copied.\n", fullPath, nBytes)
}

func main() {
	if len(os.Args) == 1 {
		log.Fatalln("Port should be specified as a first argument.")
	}
	port := os.Args[1]
	log.Fatal(fasthttp.ListenAndServe(":"+port, serve))
}
