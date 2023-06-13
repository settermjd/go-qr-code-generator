package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/nfnt/resize"
	qrcode "github.com/skip2/go-qrcode"
)

const MAX_UPLOAD_SIZE = 1024 * 1024 // 1MB
const WATERMARK_WIDTH = 64
const WATERMARK_FILENAME = "uploads/watermark.png"
const QRCODE_FILENAME = "data/qrcode.png"

type simpleQRCode struct {
	Content string
	Size    int
}

func (code *simpleQRCode) Generate() ([]byte, error) {
	png, err := qrcode.Encode(code.Content, qrcode.Medium, code.Size)
	if err != nil {
		return nil, fmt.Errorf("could not generate a QR code: %v", err)
	}
	return png, nil
}

func (code *simpleQRCode) GenerateWithWatermark() ([]byte, error) {
	err := qrcode.WriteFile(code.Content, qrcode.Medium, code.Size, QRCODE_FILENAME)
	if err != nil {
		return nil, err
	}

	err = code.addWatermark(code.Size)
	if err != nil {
		return nil, fmt.Errorf("could not add watermark to QR code: %v", err)
	}

	fileData, err := os.ReadFile(QRCODE_FILENAME)
	if err != nil {
		return nil, fmt.Errorf("could not open watermarked QR code: %v", err)
	}

	return fileData, nil
}

func (code *simpleQRCode) addWatermark(size int) error {
	qrCode, err := os.Open(QRCODE_FILENAME)
	if err != nil {
		return fmt.Errorf("could not open QR code file: %v", err)
	}
	defer qrCode.Close()

	originalImage, err := png.Decode(qrCode)
	if err != nil {
		return fmt.Errorf("could not decode QR code file: %v", err)
	}

	watermark, err := os.Open(WATERMARK_FILENAME)
	if err != nil {
		return fmt.Errorf("could not open watermark image: %v", err)
	}
	defer watermark.Close()

	watermarkImage, err := png.Decode(watermark)
	if err != nil {
		return fmt.Errorf("could not decode watermark image: %v", err)
	}

	// Center the watermark on the QR code
	offset := image.Pt(((size / 2) - 32), ((size / 2) - 32))

	// Use same size as source image has
	b := originalImage.Bounds()
	m := image.NewRGBA(b)

	draw.Draw(m, b, originalImage, image.Point{}, draw.Src)
	draw.Draw(
		m,
		watermarkImage.Bounds().Add(offset),
		watermarkImage,
		image.Point{},
		draw.Over,
	)

	out, err := os.Create(QRCODE_FILENAME)
	if err != nil {
		return fmt.Errorf("could not create watermarked QR code: %v", err)
	}
	defer out.Close()

	png.Encode(out, m)

	return nil
}

func uploadWatermarkImage(file multipart.File) error {
	dst, err := os.Create(WATERMARK_FILENAME)
	if err != nil {
		return fmt.Errorf("could not upload the watermark file: %v", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return fmt.Errorf("could not upload the watermark file: %v", err)
	}

	return nil
}

func resizeWatermarkImage(width uint) error {
	image, err := os.Open(WATERMARK_FILENAME)
	if err != nil {
		return fmt.Errorf("could not open watermark image: %v", err)
	}

	decodedImage, err := png.Decode(image)
	if err != nil {
		return fmt.Errorf("could not decode watermark image: %v", err)
	}
	image.Close()

	m := resize.Resize(width, 0, decodedImage, resize.Lanczos3)
	resizedImage, err := os.Create(WATERMARK_FILENAME)
	if err != nil {
		return fmt.Errorf("could not resize watermark image: %v", err)
	}
	defer resizedImage.Close()

	png.Encode(resizedImage, m)

	return nil
}

func detectContentType(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("could not open file: %v", err)
	}

	fileHeader := make([]byte, 512)
	if _, err := file.Read(fileHeader); err != nil {
		return "", fmt.Errorf("could not read file: %v", err)
	}

	return http.DetectContentType(fileHeader), nil
}

func createErrorResponse(message string) []byte {
	resp := make(map[string]string)
	resp["error"] = "Only PNG files are supported."
	jsonResp, err := json.Marshal(resp)

	if err != nil {
		log.Fatalln("Could not generate error message.")
	}

	return jsonResp
}

func handleRequest(writer http.ResponseWriter, request *http.Request) {
	request.ParseMultipartForm(10 << 20)
	size := request.FormValue("size")
	url := request.FormValue("url")

	if url == "" {
		writer.Write(createErrorResponse("Could not determine the desired QR code content."))
		writer.WriteHeader(400)
		return
	}

	qrCodeSize, err := strconv.Atoi(size)
	if err != nil || size == "" {
		writer.Write(createErrorResponse(fmt.Sprint("Could not determine the desired QR code size:", err)))
		writer.WriteHeader(400)
		return
	}

	qrCode := simpleQRCode{Content: url, Size: qrCodeSize}

	watermark, _, err := request.FormFile("watermark")
	if err != nil && errors.Is(err, http.ErrMissingFile) {
		fmt.Println("Error retrieving the uploaded watermark image or no watermark image was uploaded. Error details: ", err)
		codeData, err := qrCode.Generate()
		if err != nil {
			writer.Write(createErrorResponse(fmt.Sprintf("Could not generate QR code. %v", err)))
			writer.WriteHeader(400)
			return
		}
		writer.Header().Add("Content-Type", "image/png")
		writer.Write(codeData)
		return
	}

	err = uploadWatermarkImage(watermark)
	if err != nil {
		writer.Write(createErrorResponse("Could not generate QR code with the provided watermark image."))
		writer.WriteHeader(400)
		return
	}

	contentType, err := detectContentType(WATERMARK_FILENAME)
	if err != nil {
		writer.Write(createErrorResponse(
			fmt.Sprintf("Provided watermark image is not a PNG image. %v. Content type is: %s", err, contentType)),
		)
		writer.WriteHeader(400)
		return
	}

	resizeWatermarkImage(WATERMARK_WIDTH)

	defer watermark.Close()
	fileData, err := qrCode.GenerateWithWatermark()
	if err != nil {
		writer.Write(createErrorResponse("Could not generate QR code with the provided watermark image."))
		writer.WriteHeader(400)
		return
	}
	writer.Write(fileData)
}

func main() {
	http.HandleFunc("/generate", handleRequest)
	http.ListenAndServe(":8080", nil)
}
