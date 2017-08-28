package rpnr

import (
	"C"
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"io/ioutil"
	"os"
	"unsafe"

	"github.com/lazywei/go-opencv/opencv"
	"github.com/otiai10/gosseract"
)
import "errors"

const numberLength int = 9

// допилиный ToImage ибо либный не работал
func toImageMy(img *opencv.IplImage) image.Image {
	var height, width, _, step int = img.Height(), img.Width(), img.Channels(), img.WidthStep()
	out := image.NewNRGBA(image.Rect(0, 0, width, height))
	if img.Depth() != opencv.IPL_DEPTH_8U {
		return nil // TODO return error
	}
	// Turn opencv.Iplimage.imageData(*char) to slice
	var limg_ptr unsafe.Pointer = img.ImageData()
	var data []C.char = (*[1 << 30]C.char)(limg_ptr)[:height*step : height*step]
	c := color.NRGBA{R: uint8(0), G: uint8(0), B: uint8(0), A: uint8(255)}
	// Iteratively assign imageData's color to Go's image
	for y := 0; y < height; y++ {
		for x := 0; x < step; x++ {
			c.B = uint8(data[y*step+x])
			c.G = uint8(data[y*step+x])
			c.R = uint8(data[y*step+x])
			out.SetNRGBA(x, y, c)
		}
	}
	return out
}

func getLicensePlate(image *opencv.IplImage) *opencv.IplImage {
	cascade := opencv.LoadHaarClassifierCascade("test.xml")
	if cascade == nil {
		return nil
	}
	numbers := cascade.DetectObjects(image)
	if numbers == nil {
		return nil
	}
	//Getting the plate from the whole picture
	var plate *opencv.IplImage
	defer plate.Release()
	for _, value := range numbers {
		plate = opencv.Crop(image, value.X(), value.Y(), value.Width(), value.Height())
		opencv.Threshold(plate, plate, 120, 255, opencv.CV_THRESH_BINARY)
	}
	return plate
}
func getSliceOfContours(plate *opencv.IplImage) ([]opencv.Rect, int) {

	//Getting all contours on the plate as Seq
	point := opencv.Point{X: 0, Y: 0}
	borderSequance := plate.FindContours(opencv.CV_RETR_LIST, opencv.CV_CHAIN_APPROX_SIMPLE, point)
	defer borderSequance.Release()
	numberOfBorders := 0
	copyBorderSequance := borderSequance
	defer copyBorderSequance.Release()

	for i := 0; borderSequance != nil; i++ {
		numberOfBorders++
		borderSequance = borderSequance.HNext()
	}

	//Getting all contours on the plate as slice (array)
	borderArray := make([]opencv.Rect, numberOfBorders)
	for i := 0; copyBorderSequance != nil; i++ {
		borderArray[i] = opencv.BoundingRect(unsafe.Pointer(copyBorderSequance))
		copyBorderSequance = copyBorderSequance.HNext()
	}
	return borderArray, numberOfBorders
}
func getSortedSliceOfBorders(borderArray []opencv.Rect, numberOfBorders int) []opencv.Rect {
	//Sorting borders by the size
	for i := 0; i < numberOfBorders-1; i++ {
		for j := 0; j < numberOfBorders-i-1; j++ {
			if borderArray[j].Height()*borderArray[j].Width() > borderArray[j+1].Height()*borderArray[j+1].Width() {
				borderArray[j], borderArray[j+1] = borderArray[j+1], borderArray[j]
			}
		}
	}

	//Sorting borders by the pozition
	sortedBorders := borderArray[numberOfBorders-11 : numberOfBorders-2]
	for i := 0; i < numberLength; i++ {
		for j := 0; j < numberLength-i-1; j++ {
			if sortedBorders[j].X() > sortedBorders[j+1].X() {
				sortedBorders[j], sortedBorders[j+1] = sortedBorders[j+1], sortedBorders[j]
			}
		}
	}
	return sortedBorders
}
func getLicensePlateNumber(plate *opencv.IplImage, sortedBorders []opencv.Rect) string {
	//Recognition of each character
	alphabet := "ABEKMHOPCTYX0123456789"
	result := ""
	flag := false
	for i := 0; i < numberLength; i++ {
		symbol := opencv.Crop(plate, sortedBorders[i].X(), sortedBorders[i].Y(), sortedBorders[i].Width(), sortedBorders[i].Height())

		tmpfile, _ := ioutil.TempFile("src", "symb")
		var buf bytes.Buffer
		op := jpeg.Options{Quality: 99}
		jpeg.Encode(&buf, toImageMy(symbol), &op)
		tmpfile.Write(buf.Bytes())

		character := gosseract.Must(gosseract.Params{
			Src:       tmpfile.Name(),
			Languages: "eng",
		})
		symbol.Release()
		os.Remove(tmpfile.Name())
		flag = true
		if len(character) == 0 {
			continue
		}
		for i := 0; i < len(alphabet); i++ {
			if character[0] == alphabet[i] {
				character = character[0:1]
				result += character
				flag = false
				break
			}
		}
		if flag {
			result += "*"
		}

	}

	return result
}

func GetPlateNumber(filename string) (string, float64, error) {
	image := opencv.LoadImage(filename, 0)
	if image == nil {
		return "", 0, errors.New("Couldn`t open file")
	}
	defer image.Release()

	plate := getLicensePlate(image)
	defer plate.Release()
	if plate == nil {
		return "", 0, errors.New("Couldn`t detect plates")
	}
	copyPlate := plate.Clone()
	defer copyPlate.Release()

	result := getLicensePlateNumber(plate, getSortedSliceOfBorders(getSliceOfContours(copyPlate)))
	if len(result) < numberLength-1 {
		result = "Can't recognize"
	}
	return result, float64(image.Width()) / float64(image.Height()), nil
}
