package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
)

func GetEnvWithKey(key string) string {
	return os.Getenv(key)
}

func LoadEnv() {
	fmt.Print("Loading keys...\n")
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
		os.Exit(1)
	}
}

//Get All Buckets List

func GetAllBuckets(sess *session.Session) (*s3.ListBucketsOutput, error) {
	// snippet-start:[s3.go.list_buckets.imports.call]
	svc := s3.New(sess)

	BucketCount, err := svc.ListBuckets(&s3.ListBucketsInput{})
	// snippet-end:[s3.go.list_buckets.imports.call]
	if err != nil {
		return nil, err
	}

	return BucketCount, nil
}

//Byte Count to Human Readable Form

func ByteCountDecimal(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

//Main Function Begins

func main() {
	//Load Env function call to Load AWS Keys
	LoadEnv()

	//Check, the Regions as Arg

	if len(os.Args) != 2 {
		exitErrorf("Bucket Region required\nUsage: %s bucket_region",
			os.Args[0])
	}
	b_region := os.Args[1]

	// CustomBucketName := os.Args[2]

	// fmt.Printf(CustomBucketName)

	// Initialize a session, that the SDK will use to load
	// credentials from the shared credentials file ./.env.
	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(GetEnvWithKey(b_region)),
			Credentials: credentials.NewStaticCredentials(
				GetEnvWithKey("AWS_ACCESS_KEY_ID"),
				GetEnvWithKey("AWS_SECRET_ACCESS_KEY"),
				"", // a token will be created when the session it's used.
			),
		})
	if err != nil {
		panic(err)
	}

	// Create S3 service client
	svc := s3.New(sess)

	//Get List of Buckets

	BucketCount, err := GetAllBuckets(sess)
	if err != nil {
		fmt.Println("Got an error retrieving buckets:")
		fmt.Println(err)
		return
	}

	//Bucket Loop Begins
	for _, bucket := range BucketCount.Buckets {
		fmt.Println(*bucket.Name + ": " + bucket.CreationDate.Format("January 2, 2006 15:04:05 Monday"))

		// Get the list of items
		resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(*bucket.Name)})
		if err != nil {
			exitErrorf("Unable to list items in bucket %q, %v", bucket, err)
		}

		for _, item := range resp.Contents {

			fmt.Println("Name:         ", *item.Key)
			fmt.Println("Last modified:", *item.LastModified)
			fmt.Println("Size:         ", ByteCountDecimal(*item.Size))
			fmt.Println("Storage class:", *item.StorageClass)
			fmt.Println("")
		}

		fmt.Println("Found", len(resp.Contents), "items in bucket", *bucket.Name)
		fmt.Println("")
	}
	//Bucket Object Loop Ends
}

// main() Error Handle

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
