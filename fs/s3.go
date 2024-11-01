package fs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3ReadSeeker is a custom implementation of io.ReadSeeker for reading from S3 with seeking capability.
type S3ReadSeeker struct {
	client    *s3.S3
	bucket    string
	key       string
	offset    int64
	remaining io.ReadCloser
}

// NewS3ReadSeeker creates a new S3ReadSeeker.
func NewS3ReadSeeker(client *s3.S3, bucket, key string) (*S3ReadSeeker, error) {
	return &S3ReadSeeker{
		client: client,
		bucket: bucket,
		key:    key,
	}, nil
}

func (s *S3ReadSeeker) Read(p []byte) (n int, err error) {
	if s.remaining == nil {
		return 0, io.EOF
	}
	n, err = s.remaining.Read(p)
	if err == io.EOF {
		s.remaining.Close()
		s.remaining = nil
	}
	s.offset += int64(n)
	return
}

func (s *S3ReadSeeker) Seek(offset int64, whence int) (int64, error) {
	var err error
	switch whence {
	case io.SeekStart:
		s.offset = offset
	case io.SeekCurrent:
		s.offset += offset
	case io.SeekEnd:
		head, err := s.client.HeadObject(&s3.HeadObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(s.key),
		})
		if err != nil {
			return 0, err
		}
		s.offset = *head.ContentLength + offset
	default:
		return 0, errors.New("invalid whence")
	}

	if s.remaining != nil {
		s.remaining.Close()
	}
	getObjInput := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key),
		Range:  aws.String(fmt.Sprintf("bytes=%d-", s.offset)),
	}
	getObjOutput, err := s.client.GetObject(getObjInput)
	if err != nil {
		return 0, err
	}
	s.remaining = getObjOutput.Body
	return s.offset, nil
}

func (s *S3ReadSeeker) Close() error {
	if s.remaining != nil {
		return s.remaining.Close()
	}
	return nil
}

// S3VFS implements the VFS interface using Amazon S3.
type S3VFS struct {
	bucket   string
	client   *s3.S3
	uploader *s3manager.Uploader
	subdir   string
}

// NewS3VFS creates a new S3VFS.
func NewS3VFS(subdir string) (*S3VFS, error) {
	bucket := os.Getenv("AWS_BUCKET")
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")

	if bucket == "" || accessKey == "" || secretKey == "" || region == "" {
		return nil, errors.New("missing required environment variables: AWS_BUCKET, AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION")
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("error creating session: %v", err)
	}

	uploader := s3manager.NewUploader(sess)

	return &S3VFS{
		bucket:   bucket,
		client:   s3.New(sess),
		uploader: uploader,
		subdir:   subdir,
	}, nil
}

func (v *S3VFS) fullPath(key string) string {
	// we need to strip off the prefix, because
	// inside the actual bucket, the data starts
	// after /files/
	//
	// in the volume, it's like:
	// ./persistent/files/
	// ./persistent/schema.db
	//
	// and volumes start at ./persistent
	//
	// but with S3, the schema.db is persistent locally
	// and /files/ is effectively replaced with s3://
	// if you were using the cli.
	if key == "/files/" || key == "/files" {
		return "/"
	}
	if key == "/files" {
		return "/"
	}
	return key[len("/files/"):]
}

func (v *S3VFS) Open(fullName string) (io.ReadCloser, error) {
	fullPath := v.fullPath(fullName)
	result, err := v.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(v.bucket),
		Key:    aws.String(fullPath),
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

func (v *S3VFS) ReadDir(fullName string) ([]fs.DirEntry, error) {
	fullPath := v.fullPath(fullName)
	if !strings.HasSuffix(fullPath, "/") {
		fullPath += "/"
	}
	// if it does not have a trailing slash, then add one
	// except use blank for the root
	if fullPath == "/" {
		fullPath = ""
	}
	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(v.bucket),
		Prefix:    aws.String(fullPath),
		Delimiter: aws.String("/"),
	}
	result, err := v.client.ListObjectsV2(input)
	if err != nil {
		fmt.Printf("ReadDir %s problem: %v\n", fullPath, err)
		return nil, err
	}

	//fmt.Printf("ReadDir %s result: %+v\n", fullPath, result) // Debugging output

	// CommonPrefixes is a euphemism for a list of directories,
	// though they only exist as a side-effect of where the file is.
	var entries []fs.DirEntry
	for _, item := range result.CommonPrefixes {
		entries = append(
			entries,
			S3DirEntry{
				fullPath: *item.Prefix,
				obj: s3.Object{
					Key: item.Prefix,
				},
				vfs: v,
			})
	}
	// And these are the files in the directory
	for _, item := range result.Contents {
		if *item.Key != fullPath {
			entries = append(
				entries,
				S3DirEntry{
					fullPath: *item.Key,
					obj:      *item,
					vfs:      v,
				},
			)
		}
	}
	//fmt.Printf("ReadDir %s entries: %+v\n", fullPath, entries)
	return entries, nil
}

func (v *S3VFS) Remove(fullName string) error {
	fullPath := v.fullPath(fullName)
	_, err := v.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(v.bucket),
		Key:    aws.String(fullPath),
	})
	if err != nil {
		return err
	}
	return v.client.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(v.bucket),
		Key:    aws.String(fullPath),
	})
}

func (v *S3VFS) IsExist(fullName string) bool {
	if fullName == "/" {
		return true
	}
	// if it ends with a slash, it might not exist, but it's a dir path
	fullPath := v.fullPath(fullName)
	fullPath = strings.TrimSuffix(fullPath, "/")
	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(v.bucket),
		Prefix:    aws.String(fullPath),
		Delimiter: aws.String("/"),
	}
	//fullPath is guaranteed to end with a slash
	result, err := v.client.ListObjectsV2(input)
	if err != nil {
		fmt.Printf("IsExist %s problem: %v\n", fullPath, err)
		return false
	}
	//fmt.Printf("IsExist %s result: %+v\n", fullPath, result)
	if *result.KeyCount > 0 {
		return true
	}
	return false
}

func (v *S3VFS) IsNotExist(fullName string) bool {
	return !v.IsExist(fullName)
}

func (v *S3VFS) IsDir(fullName string) bool {
	// if it ends with a slash, it might not exist, but it's a dir path
	fullPath := v.fullPath(fullName)
	if strings.HasSuffix(fullPath, "/") {
		return true
	}
	// now we have to check, because this is ambiguous
	fullPath += "/"
	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(v.bucket),
		Prefix:    aws.String(fullPath),
		Delimiter: aws.String("/"),
	}
	//fullPath is guaranteed to end with a slash
	result, err := v.client.ListObjectsV2(input)
	if err != nil {
		fmt.Printf("IsDir %s problem: %v\n", fullPath, err)
		return false
	}
	//fmt.Printf("IsDir %s result: %+v\n", fullPath, result)
	if len(result.Contents) > 0 {
		if len(fullPath) < len(*result.Contents[0].Key) {
			return true
		}
	}
	return false
}

func (v *S3VFS) Create(fullName string) (io.WriteCloser, error) {
	fullPath := v.fullPath(fullName)
	pr, pw := io.Pipe()
	go func() {
		_, err := v.uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(v.bucket),
			Key:    aws.String(fullPath),
			Body:   pr,
		})
		pr.CloseWithError(err)
	}()
	return pw, nil
}

func (v *S3VFS) Size(fullName string) int64 {
	fullPath := v.fullPath(fullName)
	head, err := v.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(v.bucket),
		Key:    aws.String(fullPath),
	})
	if err != nil {
		log.Printf("Size %s problem: %v\n", fullPath, err)
		return -1
	}
	return *head.ContentLength
}

func (v *S3VFS) Date(fullName string) string {
	fullPath := v.fullPath(fullName)
	head, err := v.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(v.bucket),
		Key:    aws.String(fullPath),
	})
	if err != nil {
		log.Printf("Date %s problem: %v\n", fullPath, err)
		return ""
	}
	return head.LastModified.String()
}

func (v *S3VFS) ServeFile(w http.ResponseWriter, r *http.Request, fullName string) {
	fullPath := v.fullPath(fullName)
	readSeeker, err := NewS3ReadSeeker(v.client, v.bucket, fullPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer readSeeker.Close()

	http.ServeContent(w, r, fullName, time.Now(), readSeeker)
}

func (v *S3VFS) ReadFile(fullName string) ([]byte, error) {
	fullPath := v.fullPath(fullName)
	result, err := v.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(v.bucket),
		Key:    aws.String(fullPath),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, result.Body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (v *S3VFS) PdfThumbnail(fullName string) (io.Reader, error) {
	return nil, errors.New("PdfThumbnail not implemented")
}

func (v *S3VFS) MakeThumbnail(fullName string) (io.Reader, error) {
	return nil, errors.New("MakeThumbnail not implemented")
}

func (v *S3VFS) VideoThumbnail(fullName string) (io.Reader, error) {
	return nil, errors.New("VideoThumbnail not implemented")
}

func (v *S3VFS) FileServer() http.Handler {
	return http.StripPrefix("/", http.FileServer(http.Dir("/")))
}

type S3DirEntry struct {
	fullPath string
	obj      s3.Object
	vfs      *S3VFS
}

func (d S3DirEntry) Name() string {
	return path.Base(*d.obj.Key)
}

func (d S3DirEntry) IsDir() bool {
	return strings.HasSuffix(*d.obj.Key, "/")
}

func (d S3DirEntry) Type() fs.FileMode {
	if d.IsDir() {
		return fs.ModeDir
	}
	return 0
}

func (d S3DirEntry) Info() (fs.FileInfo, error) {
	return NewS3FileInfo(d.fullPath, d.obj, d.vfs), nil
}

type S3FileInfo struct {
	fullPath string
	obj      s3.Object
	vfs      *S3VFS
}

func NewS3FileInfo(fullPath string, obj s3.Object, vfs *S3VFS) S3FileInfo {
	return S3FileInfo{
		fullPath: fullPath,
		obj:      obj,
		vfs:      vfs,
	}
}

func (f S3FileInfo) Name() string {
	return path.Base(*f.obj.Key)
}

func (f S3FileInfo) Size() int64 {
	return f.vfs.Size("/files" + f.fullPath)
}

func (f S3FileInfo) Mode() fs.FileMode {
	if strings.HasSuffix(*f.obj.Key, "/") {
		return fs.ModeDir
	}
	return 0
}

func (f S3FileInfo) ModTime() time.Time {
	return *f.obj.LastModified
}

func (f S3FileInfo) IsDir() bool {
	return strings.HasSuffix(*f.obj.Key, "/")
}

func (f S3FileInfo) Sys() interface{} {
	return nil
}
