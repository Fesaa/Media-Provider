package services

import (
	"bytes"
	//nolint:gosec
	"crypto/md5"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/disintegration/imaging"
	"github.com/rs/zerolog"
	_ "golang.org/x/image/webp"
	"image"
)

type ImageService interface {
	// Better check if candidateImg is similar to defaultImg with MSE. If the similarityThreshold (default 0.85)
	// is reached, returns the highest resolution one. Otherwise defaultImg
	Better(defaultImg, candidateImg []byte, similarityThresholds ...float64) ([]byte, bool, error)
	Similar(img1, img2 image.Image) float64
	MeanSquareError(img1, img2 image.Image) float64
	IsCover(data []byte) bool
	Equal(img1, img2 []byte) bool
}

type imageService struct {
	log zerolog.Logger
}

func ImageServiceProvider(log zerolog.Logger) ImageService {
	return &imageService{
		log: log.With().Str("handler", "image-service").Logger(),
	}
}

//nolint:gosec
func (i *imageService) Equal(img1, img2 []byte) bool {
	sum1 := md5.Sum(img1)
	sum2 := md5.Sum(img2)
	return bytes.Equal(sum1[:], sum2[:])
}

func (i *imageService) IsCover(data []byte) bool {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		i.log.Debug().Err(err).Msg("can't decode image, assuming it's fine")
		return true
	}

	// Reference: https://wiki.kavitareader.com/guides/admin-settings/media/#cover-image-size
	// The ratio computed from values above ~ 0.74. We up the threshold to be a little more forgiving
	width, height := float64(img.Bounds().Dx()), float64(img.Bounds().Dy())
	ratio := width / max(1, height)
	return ratio <= 0.8
}

func (i *imageService) Better(defaultImg, candidateImg []byte, similarityThresholds ...float64) ([]byte, bool, error) {
	similarityThreshold := utils.OrDefault(similarityThresholds, 0.85)

	img1, _, err := image.Decode(bytes.NewReader(defaultImg))
	if err != nil {
		return nil, false, err
	}
	img2, _, err := image.Decode(bytes.NewReader(candidateImg))
	if err != nil {
		return nil, false, err
	}

	similarity := i.Similar(img1, img2)

	if similarity < similarityThreshold {
		i.log.Debug().Float64("similarity", similarity).Msg("image similarity threshold not reached, returning default")
		return defaultImg, false, nil
	}

	if i.imgResolution(img1) > i.imgResolution(img2) {
		i.log.Debug().Msg("default image has a higher resolution, returning default")
		return defaultImg, false, nil
	}

	i.log.Debug().Float64("similarity", similarity).Msg("candidate image is similar enough, and has a better resolution, returning candidate")
	return candidateImg, true, nil
}

func (i *imageService) Similar(img1, img2 image.Image) float64 {
	mse := i.MeanSquareError(img1, img2)
	normalizedMse := min(1, mse/65025)

	i.log.Trace().
		Float64("mse", mse).
		Float64("normalized mse", normalizedMse).
		Send()
	return max(0, 1-normalizedMse)
}

func (i *imageService) MeanSquareError(img1, img2 image.Image) float64 {
	if !img1.Bounds().Eq(img2.Bounds()) {
		img2 = imaging.Resize(img2, img1.Bounds().Dx(), img1.Bounds().Dy(), imaging.Lanczos)
	}

	var sumDiff float64

	for y := range img1.Bounds().Dy() {
		for x := range img1.Bounds().Dx() {
			r1, g1, b1, _ := img1.At(x, y).RGBA()
			r2, g2, b2, _ := img2.At(x, y).RGBA()

			r1, g1, b1 = r1>>8, g1>>8, b1>>8
			r2, g2, b2 = r2>>8, g2>>8, b2>>8

			diffR := float64(r1) - float64(r2)
			diffG := float64(g1) - float64(g2)
			diffB := float64(b1) - float64(b2)

			diff := diffR*diffR + diffG*diffG + diffB*diffB

			sumDiff += diff
		}
	}

	return sumDiff / (float64(i.imgResolution(img1)))
}

// ImgResolution returns the product of Dx and Dy
func (i *imageService) imgResolution(img image.Image) int {
	return img.Bounds().Dx() * img.Bounds().Dy()
}
