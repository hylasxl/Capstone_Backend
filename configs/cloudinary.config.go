package configs

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"os"
)

type CloudinaryService struct {
	Client *cloudinary.Cloudinary
	Ctx    context.Context
}

func InitCloudinary(ctx context.Context) (*CloudinaryService, error) {
	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudinary: %v", err)
	}
	return &CloudinaryService{Client: cld, Ctx: ctx}, nil
}

func (cs *CloudinaryService) UploadAvatarImage(data []byte) (string, error) {
	if len(data) > 10*1024*1024 {
		var compressData []byte
		var err error
		compressData, err = cs.CompressImage(data)
		if err != nil {
			return "", err
		}
		data = compressData
	}
	imageReader := bytes.NewReader(data)
	uploadParams := uploader.UploadParams{
		Folder: "user_profile_image",
	}
	uploadResult, err := cs.Client.Upload.Upload(cs.Ctx, imageReader, uploadParams)
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %v", err)
	}
	return uploadResult.SecureURL, nil
}

func (cs *CloudinaryService) DeleteFile(publicID string) error {
	_, err := cs.Client.Upload.Destroy(cs.Ctx, uploader.DestroyParams{PublicID: publicID})
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}
	return nil
}

func (cs *CloudinaryService) CheckConnection() error {
	_, err := cs.Client.Admin.Ping(cs.Ctx)
	if err != nil {
		return fmt.Errorf("failed to check connection: %v", err)
	}
	return nil
}

func (cs *CloudinaryService) CompressImage(data []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}
	img = resize.Thumbnail(1200, 1200, img, resize.Lanczos3)

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 75})
	if err != nil {
		return nil, fmt.Errorf("failed to compress image: %v", err)
	}
	return buf.Bytes(), nil
}
