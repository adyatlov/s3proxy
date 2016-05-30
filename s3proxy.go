package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"log"
	"net/url"
	"os"
)

func parseUrl(u *url.URL) (region, bucket, path string, err error) {
	return "us-west-2", "andrey-so-36323287", "/pi.conf", nil
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

func main() {
	region, bucket, path, err := parseUrl(nil)
	config := getAWSConfig(region)
	sess := session.New(config)
	client := s3.New(sess)
	req := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	}
	resp, err := client.GetObject(req)
	if err != nil {
		log.Fatalln("Cannot GetObject:", err)
	}
	_, err = io.Copy(os.Stdout, resp.Body)
}
