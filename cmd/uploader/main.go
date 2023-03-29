package main

import (
	"fmt"
	"os"
	"io"
	"sync"
	"github.com/lauronicolas/curso-go/eventos/pkg/rabbitmq"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws"
)

var (
	s3Client *s3.S3
	s3Bucket string
	wg sync.WaitGroup
)

func init(){
	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String("us-east-1"),
			Credentials: credentials.NewStaticCredentials(
				"",
				"",
				"",
			),
		},
	)

	if err != nil {
		panic(err)
	}
	s3Client = s3.New(sess)
	s3Bucket = "example-goexpert-bucket"
}

func main() {
	dir, err := os.Open("./tmp")
	if err != nil {
		panic(err)
	}
	defer dir.Close()

	uploadControl := make(chan string, 100)
	errorFileUpload := make(chan string, 10)

	channel, err := rabbitmq.OpenChannel()
	if err != nil {
		panic(err)
	}
	defer channel.Close()

	go func(){
		for{
			select{
			case fileName := <- uploadControl:
				body := fmt.Sprintf("Arquivo %s inserido com sucesso no S3", fileName)
				rabbitmq.Publish(channel, body, "amq.direct")
			}
		}
	}()

	go func(){
		for{
			select{
			case fileName := <- errorFileUpload:
				uploadControl <- fileName
				wg.Add(1)
				go uploadFile(fileName, uploadControl, errorFileUpload)
			}
		}
	}()

	for {
		files, err := dir.ReadDir(1)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("Error reading directory: %s\n", err)
			continue
		}
		wg.Add(1)
		uploadControl <- files[0].Name()
		go uploadFile(files[0].Name(), uploadControl, errorFileUpload)
	}
	wg.Wait()
}

func uploadFile(fileName string, uploadControl <-chan string, errorFileUpload chan<- string) {
	defer wg.Done()
	completeFileName := fmt.Sprintf("./tmp/%s", fileName)
	f, err := os.Open(completeFileName)
	if err != nil {
		fmt.Printf("Error opening file %s\n", completeFileName)
		<-uploadControl
		errorFileUpload <- fileName
		return
	}
	defer f.Close()

	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key: aws.String(fileName),
		Body: f,
	})
	if err != nil {
		fmt.Printf("Error uploading file %s\n", completeFileName)
		<-uploadControl
		errorFileUpload <- fileName
		return
	}
	fmt.Printf("File %s uploaded successfully\n", completeFileName)
	<-uploadControl
}
