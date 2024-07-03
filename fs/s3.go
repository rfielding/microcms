package fs

import (
    "bytes"
    "errors"
    "fmt"
    "io"
    "io/fs"
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
    buf       *bytes.Reader
    s3Stream  io.ReadCloser
    offset    int64
    bytesRead int64
}

// NewS3ReadSeeker creates a new S3ReadSeeker.
func NewS3ReadSeeker(client *s3.S3, bucket, key string, buf *bytes.Reader, s3Stream io.ReadCloser) *S3ReadSeeker {
    return &S3ReadSeeker{
        client:   client,
        bucket:   bucket,
        key:      key,
        buf:      buf,
        s3Stream: s3Stream,
        offset:   0,
    }
}

func (s *S3ReadSeeker) Read(p []byte) (n int, err error) {
    if s.buf != nil {
        n, err = s.buf.Read(p)
        s.bytesRead += int64(n)
        if err == io.EOF {
            s.buf = nil
            err = nil
        }
        if err != nil {
            return
        }
    }
    n, err = s.s3Stream.Read(p)
    s.bytesRead += int64(n)
    return
}

func (s *S3ReadSeeker) Seek(offset int64, whence int) (int64, error) {
    if s.buf == nil {
        rangeStart := s.offset + offset
        if whence == io.SeekStart {
            rangeStart = offset
        } else if whence == io.SeekEnd {
            head, err := s.client.HeadObject(&s3.HeadObjectInput{
                Bucket: aws.String(s.bucket),
                Key:    aws.String(s.key),
            })
            if err != nil {
                return 0, err
            }
            rangeStart = *head.ContentLength + offset
        }

        s.offset = rangeStart

        s3Stream, err := s.client.GetObject(&s3.GetObjectInput{
            Bucket: aws.String(s.bucket),
            Key:    aws.String(s.key),
            Range:  aws.String(fmt.Sprintf("bytes=%d-", rangeStart)),
        })
        if err != nil {
            return 0, err
        }
        s.s3Stream.Close()
        s.s3Stream = s3Stream.Body

        s.bytesRead = rangeStart
        return rangeStart, nil
    }

    pos, err := s.buf.Seek(offset, whence)
    s.bytesRead = pos
    return pos, err
}

func (s *S3ReadSeeker) Close() error {
    return s.s3Stream.Close()
}

// S3VFS implements the VFS interface using Amazon S3.
type S3VFS struct {
    bucket string
    client *s3.S3
    uploader *s3manager.Uploader
    subdir string
}

// NewS3VFS creates a new S3VFS.
func NewS3VFS(subdir string) (*S3VFS, error) {
    bucket := os.Getenv("AWS_BUCKET")
    accessKey := os.Getenv("AWS_ACCESS_KEY")
    secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
    region := os.Getenv("AWS_REGION")

    if bucket == "" || accessKey == "" || secretKey == "" || region == "" {
        return nil, errors.New("Missing required environment variables: AWS_BUCKET, AWS_ACCESS_KEY, AWS_SECRET_ACCESS_KEY, AWS_REGION")
    }

    sess, err := session.NewSession(&aws.Config{
        Region:      aws.String(region),
        Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
    })
    if err != nil {
        return nil, fmt.Errorf("Error creating session: %v", err)
    }

    uploader := s3manager.NewUploader(sess)

    return &S3VFS{
        bucket: bucket,
        client: s3.New(sess),
        uploader: uploader,
        subdir: subdir,
    }, nil
}

func (v *S3VFS) fullPath(key string) string {
    if v.subdir == "" {
        return key
    }
    return path.Join(v.subdir, key)
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
    input := &s3.ListObjectsV2Input{
        Bucket: aws.String(v.bucket),
        Prefix: aws.String(fullPath),
    }
    result, err := v.client.ListObjectsV2(input)
    if err != nil {
        return nil, err
    }

    var entries []fs.DirEntry
    for _, item := range result.Contents {
        entries = append(entries, s3DirEntry{*item})
    }
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
    fullPath := v.fullPath(fullName)
    _, err := v.client.HeadObject(&s3.HeadObjectInput{
        Bucket: aws.String(v.bucket),
        Key:    aws.String(fullPath),
    })
    return err == nil
}

func (v *S3VFS) IsNotExist(fullName string) bool {
    return !v.IsExist(fullName)
}

func (v *S3VFS) IsDir(fullName string) bool {
    fullPath := v.fullPath(fullName)
    input := &s3.ListObjectsV2Input{
        Bucket: aws.String(v.bucket),
        Prefix: aws.String(fullPath),
    }
    result, err := v.client.ListObjectsV2(input)
    if err != nil {
        return false
    }
    return len(result.Contents) > 1 || (len(result.Contents) == 1 && strings.HasSuffix(*result.Contents[0].Key, "/"))
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
        return -1
    }
    return *head.ContentLength
}

func (v *S3VFS) ServeFile(w http.ResponseWriter, r *http.Request, fullName string) {
    fullPath := v.fullPath(fullName)
    input := &s3.GetObjectInput{
        Bucket: aws.String(v.bucket),
        Key:    aws.String(fullPath),
    }
    result, err := v.client.GetObject(input)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Create a buffer for the initial seek
    buf := new(bytes.Buffer)
    _, err = io.CopyN(buf, result.Body, 4096) // Read the first 4KB into the buffer
    if err != nil && err != io.EOF {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Use the custom ReadSeeker
    readSeeker := NewS3ReadSeeker(v.client, v.bucket, fullPath, bytes.NewReader(buf.Bytes()), result.Body)
    defer readSeeker.Close()

    http.ServeContent(w, r, fullName, *result.LastModified, readSeeker)
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

type s3DirEntry struct {
    obj s3.Object
}

func (d s3DirEntry) Name() string {
    return path.Base(*d.obj.Key)
}

func (d s3DirEntry) IsDir() bool {
    return strings.HasSuffix(*d.obj.Key, "/")
}

func (d s3DirEntry) Type() fs.FileMode {
    if d.IsDir() {
        return fs.ModeDir
    }
    return 0
}

func (d s3DirEntry) Info() (fs.FileInfo, error) {
    return s3FileInfo{d.obj}, nil
}

type s3FileInfo struct {
    obj s3.Object
}

func (f s3FileInfo) Name() string {
    return path.Base(*f.obj.Key)
}

func (f s3FileInfo) Size() int64 {
    return *f.obj.Size
}

func (f s3FileInfo) Mode() fs.FileMode {
    if strings.HasSuffix(*f.obj.Key, "/") {
        return fs.ModeDir
    }
    return 0
}

func (f s3FileInfo) ModTime() time.Time {
    return *f.obj.LastModified
}

func (f s3FileInfo) IsDir() bool {
    return strings.HasSuffix(*f.obj.Key, "/")
}

func (f s3FileInfo) Sys() interface{} {
    return nil
}

