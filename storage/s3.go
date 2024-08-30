package storage

import (
	"bytes"
	"fmt"
	"image/png"
	"io"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/alvarofc/mode/types"
	"github.com/alvarofc/mode/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nfnt/resize"
	"github.com/patrickmn/go-cache"
)

// S3Client represents a client for interacting with Amazon S3 or compatible storage services
type S3Client struct {
	Client *s3.S3
	Cache  *cache.Cache
}

// NewS3Client creates and returns a new S3Client instance
// It sets up the AWS session and S3 client with the provided credentials and configuration
func NewS3Client(key, secret, url, region string) S3Client {
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(key, secret, ""),
		Endpoint:         aws.String(url),
		Region:           aws.String(region),
		S3ForcePathStyle: aws.Bool(true),
	}

	sess := session.Must(session.NewSession(s3Config))

	s3Client := s3.New(sess)

	// Create a cache with a default expiration time of 5 minutes and purge unused items every 10 minutes
	cacheInstance := cache.New(5*time.Minute, 10*time.Minute)

	return S3Client{
		Client: s3Client,
		Cache:  cacheInstance,
	}
}

// DownloadPhotoByKey retrieves a photo from S3 by its key
// It returns the photo data as a byte slice
func (s *S3Client) DownloadPhotoByKey(key string) ([]byte, error) {

	result, err := s.Client.GetObject((&s3.GetObjectInput{

		Bucket: aws.String(os.Getenv("BUCKET_NAME")),
		Key:    aws.String(key),
	}))

	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)

	if err != nil {
		return nil, err
	}

	return body, nil

}

//func (s *S3) UploadPhoto(id int) (types.Photo, error) {}

// DownloadSmallPhotoByKey retrieves a photo from S3 by its key and resizes it
// It returns the resized photo data as a byte slice
// The photo is resized to 800x600 pixels using the Lanczos3 resampling filter
func (s *S3Client) DownloadSmallPhotoByKey(key string) ([]byte, error) {
	pngData, err := s.DownloadPhotoByKey(key)
	if err != nil {
		return nil, fmt.Errorf("error getting original PNG: %w", err)
	}

	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, fmt.Errorf("error decoding PNG: %w", err)
	}

	resizedImg := resize.Resize(800, 600, img, resize.Lanczos3)

	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, resizedImg); err != nil {
		return nil, fmt.Errorf("error encoding PNG: %w", err)
	}

	return buffer.Bytes(), nil
}

// Helper function to process S3 objects
func (s *S3Client) processS3Objects(userID string, limit int64) ([]types.ImageInfo, error) {
	bucket := os.Getenv("BUCKET_NAME")
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(fmt.Sprintf("user_%s/", userID)),
	}

	var images []types.ImageInfo
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	err := s.Client.ListObjectsV2Pages(input,
		func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, item := range page.Contents {
				if !utils.IsImage(*item.Key) {
					continue
				}
				wg.Add(1)
				go func(item *s3.Object) {
					defer wg.Done()
					if utils.IsImage(*item.Key) {
						mu.Lock()
						images = append(images, types.ImageInfo{
							Key:      *item.Key,
							Size:     *item.Size,
							Modified: *item.LastModified,
						})
						mu.Unlock()
					}
				}(item)
			}
			return true // continue paging
		})

	if err != nil {
		errChan <- fmt.Errorf("error listing objects: %w", err)
	}

	wg.Wait()
	close(errChan)

	if err := <-errChan; err != nil {
		return nil, err
	}

	// Sort images by last modified date (newest first)
	sort.Slice(images, func(i, j int) bool {
		return images[i].Modified.After(images[j].Modified)
	})

	// Limit the number of images if necessary
	if limit > 0 && int64(len(images)) > limit {
		images = images[:limit]
	}

	return images, nil
}

// GetLastXPhotosForUser retrieves the last X photos for a specific user
func (s *S3Client) GetLastXPhotosForUser(userID string, photoNum int64) ([]types.ImageInfo, error) {
	cacheKey := fmt.Sprintf("last_%d_photos_%s", photoNum, userID)

	if cachedImages, found := s.Cache.Get(cacheKey); found {
		return cachedImages.([]types.ImageInfo), nil
	}

	images, err := s.processS3Objects(userID, photoNum)
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("no image files found for user %s", userID)
	}
	bucket := os.Getenv("BUCKET_NAME")

	// Generate presigned URLs
	var wg sync.WaitGroup
	errChan := make(chan error, len(images))

	for i := range images {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req, _ := s.Client.GetObjectRequest(&s3.GetObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(images[i].Key),
			})
			urlStr, err := req.Presign(1 * time.Hour)
			if err != nil {
				errChan <- fmt.Errorf("error generating presigned URL for image %d: %w", i, err)
				return
			}
			images[i].URL = urlStr
		}(i)
	}

	wg.Wait()
	close(errChan)

	if err := <-errChan; err != nil {
		return nil, err
	}

	// Cache the result
	s.Cache.Set(cacheKey, images, cache.DefaultExpiration)

	return images, nil
}

// GetLastPhotoForUser retrieves the most recent photo for a specific user
func (s *S3Client) GetLastPhotoForUser(userID string) (types.ImageInfo, error) {
	cacheKey := fmt.Sprintf("last_photo_%s", userID)

	if cachedImage, found := s.Cache.Get(cacheKey); found {
		return cachedImage.(types.ImageInfo), nil
	}

	images, err := s.processS3Objects(userID, 1)
	if err != nil {
		return types.ImageInfo{}, err
	}

	if len(images) == 0 {
		return types.ImageInfo{}, fmt.Errorf("no image files found for user %s", userID)
	}

	lastImage := images[0]

	// Generate presigned URL
	req, _ := s.Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("BUCKET_NAME")),
		Key:    aws.String(lastImage.Key),
	})
	urlStr, err := req.Presign(1 * time.Hour)
	if err != nil {
		return types.ImageInfo{}, fmt.Errorf("error generating presigned URL: %w", err)
	}
	lastImage.URL = urlStr

	// Cache the result
	s.Cache.Set(cacheKey, lastImage, cache.DefaultExpiration)

	return lastImage, nil
}
