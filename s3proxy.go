package main

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func parseUrl(u *url.URL) (region, bucket, path string, err error) {
	args := strings.SplitN(strings.TrimPrefix(u.EscapedPath(), "/"), "/", 3)
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

func serve(w http.ResponseWriter, r *http.Request) {
	log.Println("Downloading", r.URL.EscapedPath())
	region, bucket, path, err := parseUrl(r.URL)
	if err != nil {
		log.Println("Cannot parse URL:", err)
		http.NotFound(w, r)
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
		http.NotFound(w, r)
		return
	}
	nBytes, err := io.Copy(w, s3resp.Body)
	s3resp.Body.Close()
	log.Printf("%v: %v bytes are copied.\n", r.URL.EscapedPath(), nBytes)
}

func main() {
	http.HandleFunc("/", serve)
	if len(os.Args) == 1 {
		log.Fatalln("Port should be specified as a first argument.")
	}
	port := os.Args[1]
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
