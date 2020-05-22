package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/aws/aws-sdk-go/aws/credentials"
	_ "github.com/aws/aws-sdk-go/service/s3"
)

type Config struct {
	S3Bucket   string
	S3Path     string
	AWSProfile string
}

func isInit() {

}

func Init(gitRoot string, bucket string, s3Prefix string, credentialProfile string) {
    if

	configFile := path.Join(gitRoot, ".git", "config")

	var key string
	var err error
	var configData []byte

	if bucket == "" {
		fmt.Printf("Must provide a bucket to init\n")
		os.Exit(-1)
	}

	// check that creds are readable
	creds := credentials.NewSharedCredentials("", credentialProfile)
	if _, err = creds.Get(); err != nil {
		fmt.Printf("While reading credentials, got: %v\n", err)
		os.Exit(-1)
	}

	//	if s3Prefix == "" {
	//		key = "test"
	//	} else {
	//		key = path.Join(s3Prefix, "test")
	//	}
	//	testData := []byte("nope")
	//	if err = sendBytesToS3(credentialProfile, bucket, key, testData); err != nil {
	//		fmt.Printf("Failed to send to bucket: 's3://%v/%v, error: %v\n", bucket, key, err)
	//		os.Exit(-1)
	//	}

	// write out to config
	config := new(Config)
	config.S3Path = s3Prefix
	config.S3Bucket = bucket
	config.AWSProfile = credentialProfile

	if configData, err = json.Marshal(config); err != nil {
		fmt.Println("While marshalling config, got: %v", err)
		os.Exit(-1)
	}

	if err = ioutil.WriteFile(configFile, configData, 0600); err != nil {
		fmt.Println("Failed to write config '%v' to file: %v", configData, configFile)
		os.Exit(-1)
	}
}
