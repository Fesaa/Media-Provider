package services

import (
	"bytes"
	"context"
	"fmt"
	"image"

	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/internal/tracing"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const (
	webpFormatKey = "webp"
)

var (
	encodingOptions = &webp.Options{
		Lossless: true,
		Quality:  80,
	}
)

type ImageService interface {
	// Better check if candidateImg is similar to defaultImg with MSE. If the similarityThreshold (default 0.85)
	// is reached, returns the highest resolution one. Otherwise defaultImg
	Better(defaultImg, candidateImg []byte, similarityThresholds ...float64) ([]byte, bool, error)
	Similar(img1, img2 image.Image) float64
	MeanSquareError(img1, img2 image.Image) float64
	IsCover(data []byte) bool
	ToImage(data []byte) (image.Image, error)
	ConvertToWebp(ctx context.Context, data []byte) ([]byte, bool)
}

type imageService struct {
	log        zerolog.Logger
	unitOfWork *db.UnitOfWork
}

func ImageServiceProvider(log zerolog.Logger, unitOfWork *db.UnitOfWork) ImageService {
	return &imageService{
		log:        log.With().Str("handler", "image-service").Logger(),
		unitOfWork: unitOfWork,
	}
}

func (i *imageService) ConvertToWebp(ctx context.Context, data []byte) ([]byte, bool) {
	ctx, span := tracing.TracerServices.Start(ctx, tracing.SpanServicesImagesWebp)
	defer span.End()

	p, err := i.unitOfWork.Preferences.GetPreferences(ctx)
	if err != nil {
		i.log.Warn().Err(err).Msg("pref.Get() failed, not converting to webp")
		return data, false
	}

	if !p.ConvertToWebp {
		return data, false
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("image.Decode() failed: %s", err.Error()))
		return data, false
	}

	span.SetAttributes(
		attribute.Int("input.size", len(data)),
		attribute.String("image.format", format),
		attribute.Bool("converted", format != webpFormatKey),
	)

	if format == webpFormatKey {
		return data, false
	}

	buf := new(bytes.Buffer)
	if err = webp.Encode(buf, img, encodingOptions); err != nil {
		i.log.Error().Err(err).Msg("webp.Encode() failed")
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("webp.Encode() failed: %s", err.Error()))
		return data, false
	}

	return buf.Bytes(), true
}

func (i *imageService) ToImage(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return img, nil
}

func (i *imageService) IsCover(data []byte) bool {
	img, err := i.ToImage(data)
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

	img1, err := i.ToImage(defaultImg)
	if err != nil {
		return nil, false, err
	}
	img2, err := i.ToImage(candidateImg)
	if err != nil {
		return nil, false, err
	}

	similarity := i.Similar(img1, img2)

	if similarity < similarityThreshold {
		i.log.Trace().Float64("similarity", similarity).Msg("image similarity threshold not reached, returning default")
		return defaultImg, false, nil
	}

	if i.imgResolution(img1) > i.imgResolution(img2) {
		i.log.Trace().Msg("default image has a higher resolution, returning default")
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
