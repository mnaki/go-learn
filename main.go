package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/disintegration/gift"
	"github.com/gin-gonic/gin"
)

type Filter struct {
	brightness float32
	contrast   float32
	saturation float32
}

func FilterService(filePath string) []string {
	fmt.Printf("filePath: %v\n", filePath)
	src := loadImage(filePath)
	imageMap := filterImage(src)
	var finalPathes []string
	for name, img := range imageMap {
		imageFilePath := path.Join("./.tmp", "dst_"+name+".jpg")
		saveImage(imageFilePath, &img)
		finalPathes = append(finalPathes, imageFilePath)
	}
	return finalPathes
}

func filterImage(src image.Image) map[string]image.NRGBA {

	brightness := []float32{-10, 10}
	contrast := []float32{10, 30}
	saturation := []float32{10, 30}

	var images = make(map[string]image.NRGBA)
	for _, b := range brightness {
		for _, c := range contrast {
			for _, s := range saturation {
				name := fmt.Sprintf("B%v_C%v_S%v", b, c, s)
				images[name] = applyFilter(Filter{b, c, s}, src, name)
			}
		}
	}
	return images
}

func applyFilter(filter Filter, src image.Image, name string) image.NRGBA {
	g := gift.New(
		gift.ResizeToFit(1024, 1024, gift.LanczosResampling),
		gift.ColorBalance(0, 5, 10),
		gift.Brightness(filter.brightness),
		gift.Contrast(filter.contrast),
		gift.Saturation(filter.saturation),
		gift.Rotate270(),
		gift.Convolution(
			[]float32{
				0, 0, 0,
				-1, 1, 1,
				0, 1, 0,
			},
			false, false, false, 0.0,
		),
	)
	fmt.Println(g)
	dst := image.NewNRGBA(g.Bounds(src.Bounds()))
	g.Draw(dst, src)
	return *dst
}

func loadImage(filename string) image.Image {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("os.Open failed: %v", err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		log.Fatalf("image.Decode failed: %v", err)
	}
	return img
}

func saveImage(filename string, img image.Image) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("os.Create failed: %v", err)
	}
	defer f.Close()
	opt := &jpeg.Options{Quality: 100}
	err = jpeg.Encode(f, img, opt)
	if err != nil {
		log.Fatalf("png.Encode failed: %v", err)
	}
}

func main() {
	router := gin.Default()
	router.POST("/filter_image", func(context *gin.Context) {
		file, err := context.FormFile("file")
		log.Println("file:", file.Filename)
		if err != nil {
			log.Fatalln(err)
		}
		filePath := path.Join("./.tmp", file.Filename)
		context.SaveUploadedFile(file, filePath)
		savedImagePathes := FilterService(filePath)
		context.String(http.StatusOK, fmt.Sprintf("%v", savedImagePathes))
	})
	router.Run(":3000")
}
