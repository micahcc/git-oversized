package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Config struct {
	S3Bucket   string
	S3Path     string
	AWSProfile string
}

func Init(gitRoot string, bucket string, s3Prefix string, credentialProfile string) {
	configFile := path.Join(gitRoot, "phat.json")
	var err error
	var configData []byte

	// check that creds are readable
	creds := credentials.NewSharedCredentials("", credentialProfile)
	if _, err = creds.Get(); err != nil {
		fmt.Println("While reading credentials, got: %v", err)
		os.Exit(-1)
	}

	// test that bucket is writeable

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
