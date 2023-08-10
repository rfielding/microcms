package handler

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/rfielding/microcms/fs"
	"github.com/rfielding/microcms/utils"
)

// detectLabels gets labels from the rekognition API for an image at the given file path.
func detectLabels(file string) (io.Reader, error) {
	svc := rekognition.New(session.New())

	imageBytes, err := fs.F.ReadFile(file)
	if err != nil {
		return nil, err
	}

	input := &rekognition.DetectLabelsInput{
		Image: &rekognition.Image{
			Bytes: imageBytes,
		},
		MaxLabels:     aws.Int64(123),
		MinConfidence: aws.Float64(70.000000),
	}

	result, err := svc.DetectLabels(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rekognition.ErrCodeInvalidS3ObjectException:
				fmt.Println(rekognition.ErrCodeInvalidS3ObjectException, aerr.Error())
			case rekognition.ErrCodeInvalidParameterException:
				fmt.Println(rekognition.ErrCodeInvalidParameterException, aerr.Error())
			case rekognition.ErrCodeImageTooLargeException:
				fmt.Println(rekognition.ErrCodeImageTooLargeException, aerr.Error())
			case rekognition.ErrCodeAccessDeniedException:
				fmt.Println(rekognition.ErrCodeAccessDeniedException, aerr.Error())
			case rekognition.ErrCodeInternalServerError:
				fmt.Println(rekognition.ErrCodeInternalServerError, aerr.Error())
			case rekognition.ErrCodeThrottlingException:
				fmt.Println(rekognition.ErrCodeThrottlingException, aerr.Error())
			case rekognition.ErrCodeProvisionedThroughputExceededException:
				fmt.Println(rekognition.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case rekognition.ErrCodeInvalidImageFormatException:
				fmt.Println(rekognition.ErrCodeInvalidImageFormatException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		}
		return nil, err
	}

	pipeReader, pipeWriter := io.Pipe()
	go func() {
		pipeWriter.Write([]byte(utils.AsJson(result)))
		pipeWriter.Close()
	}()
	return pipeReader, nil
}

func detectCeleb(file string) (io.Reader, error) {
	svc := rekognition.New(session.New())

	imageBytes, err := fs.F.ReadFile(file)
	if err != nil {
		return nil, err
	}

	input := &rekognition.RecognizeCelebritiesInput{
		Image: &rekognition.Image{
			Bytes: imageBytes,
		},
	}

	result, err := svc.RecognizeCelebrities(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rekognition.ErrCodeInvalidS3ObjectException:
				fmt.Println(rekognition.ErrCodeInvalidS3ObjectException, aerr.Error())
			case rekognition.ErrCodeInvalidParameterException:
				fmt.Println(rekognition.ErrCodeInvalidParameterException, aerr.Error())
			case rekognition.ErrCodeImageTooLargeException:
				fmt.Println(rekognition.ErrCodeImageTooLargeException, aerr.Error())
			case rekognition.ErrCodeAccessDeniedException:
				fmt.Println(rekognition.ErrCodeAccessDeniedException, aerr.Error())
			case rekognition.ErrCodeInternalServerError:
				fmt.Println(rekognition.ErrCodeInternalServerError, aerr.Error())
			case rekognition.ErrCodeThrottlingException:
				fmt.Println(rekognition.ErrCodeThrottlingException, aerr.Error())
			case rekognition.ErrCodeProvisionedThroughputExceededException:
				fmt.Println(rekognition.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case rekognition.ErrCodeInvalidImageFormatException:
				fmt.Println(rekognition.ErrCodeInvalidImageFormatException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		}
		return nil, err
	}

	pipeReader, pipeWriter := io.Pipe()
	go func() {
		pipeWriter.Write([]byte(utils.AsJson(result)))
		pipeWriter.Close()
	}()
	return pipeReader, nil
}

func detectModeration(file string) (io.Reader, error) {
	svc := rekognition.New(session.New())

	imageBytes, err := fs.F.ReadFile(file)
	if err != nil {
		return nil, err
	}

	input := &rekognition.DetectModerationLabelsInput{
		Image: &rekognition.Image{
			Bytes: imageBytes,
		},
	}

	result, err := svc.DetectModerationLabels(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rekognition.ErrCodeInvalidS3ObjectException:
				fmt.Println(rekognition.ErrCodeInvalidS3ObjectException, aerr.Error())
			case rekognition.ErrCodeInvalidParameterException:
				fmt.Println(rekognition.ErrCodeInvalidParameterException, aerr.Error())
			case rekognition.ErrCodeImageTooLargeException:
				fmt.Println(rekognition.ErrCodeImageTooLargeException, aerr.Error())
			case rekognition.ErrCodeAccessDeniedException:
				fmt.Println(rekognition.ErrCodeAccessDeniedException, aerr.Error())
			case rekognition.ErrCodeInternalServerError:
				fmt.Println(rekognition.ErrCodeInternalServerError, aerr.Error())
			case rekognition.ErrCodeThrottlingException:
				fmt.Println(rekognition.ErrCodeThrottlingException, aerr.Error())
			case rekognition.ErrCodeProvisionedThroughputExceededException:
				fmt.Println(rekognition.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case rekognition.ErrCodeInvalidImageFormatException:
				fmt.Println(rekognition.ErrCodeInvalidImageFormatException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		}
		return nil, err
	}

	pipeReader, pipeWriter := io.Pipe()
	go func() {
		pipeWriter.Write([]byte(utils.AsJson(result)))
		pipeWriter.Close()
	}()
	return pipeReader, nil
}
