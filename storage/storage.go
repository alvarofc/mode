package storage

import "github.com/alvarofc/mode/types"

type Storage interface {
	GetUserById(id int) (types.User, error)
	GetUserByEmail(email string) (types.User, error)
	CreateUser(email, password string) error
}

type S3 interface {
	DownloadPhotoByKey(key string) ([]byte, error)
	//GetWEBPPhotoByKey(key string) ([]byte, error)
	DownloadSmallPhotoByKey(key string) ([]byte, error)
	GetLastXPhotosForUser(userID string, photoNum int64) ([]types.ImageInfo, error)
	GetLastPhotoForUser(userID string) (types.ImageInfo, error)
}
