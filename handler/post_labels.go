package handler

import (
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/rfielding/microcms/db"
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

func indexTextFile(
	path string,
	name string,
	part int,
	originalPath string,
	originalName string,
	content []byte,
) error {
	// index the file -- if we are appending, we should only incrementally index
	_, err := db.TheDB.Exec(
		`INSERT INTO filesearch (cmd, path, name, part, original_path, original_name, content) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"files",
		path,
		name,
		part,
		originalPath,
		originalName,
		content,
	)
	if err != nil {
		return fmt.Errorf("ERR while indexing files %s%s: %v", path, name, err)
	}
	return nil
}

func splitByCaseChanges(s string) string {
	var result []string
	var wordStart int

	for i, r := range s {
		if i == 0 {
			continue
		}

		// Detect changes between uppercase, lowercase, and digits
		prevRune := rune(s[i-1])
		changeDetected := (unicode.IsUpper(r) && !unicode.IsUpper(prevRune)) ||
			(unicode.IsDigit(r) != unicode.IsDigit(prevRune))

		if changeDetected {
			result = append(result, s[wordStart:i])
			wordStart = i
		}
	}

	// Add the remaining part
	result = append(result, s[wordStart:])

	return strings.Join(result, " ")
}

func generateNameMeta(path string, name string) []byte {
	// put in name in a way that works with keyword searches
	var words []string
	words = append(words, name)
	name2 := strings.ReplaceAll(name, "_", " ")
	name2 = strings.ReplaceAll(name2, ".", " ")
	name2 = strings.ReplaceAll(name2, "-", " ")
	words = append(words, name2)
	words = append(words, splitByCaseChanges(name2))
	return []byte(strings.ToLower(strings.Join(words, " ")))
}

func indexFileName(
	path string,
	name string,
) error {
	// index the file -- if we are appending, we should only incrementally index
	md := generateNameMeta(path, name)
	//log.Printf("index %s with: %s", name, string(md))
	_, err := db.TheDB.Exec(
		`INSERT INTO filesearch (cmd, path, name, part, original_path, original_name, content) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"files",
		path,
		name,
		0,
		path,
		name,
		md,
	)
	if err != nil {
		return fmt.Errorf("ERR while indexing files %s%s: %v", path, name, err)
	}
	return nil
}
