# Generate a QR code in Go

This is a small repository showing how to generate a QR code, with an optional watermark, in Go.

This project is the complete code behind the ["How to Generate a QR Code with Go"][tutorial-url] tutorial on the Twilio blog.

## Prerequisites

To follow along with the tutorial, you don't need much, just the following things:

- [Go][go-url] (a recent version, or the latest, 1.20.5)
- [Curl][curl-url] or [Postman][postman-url]
- A smartphone with a QR code scanner (which, these days, most of them should have)

## Start the application

To start the application, run the following command:

```bash
go run main.go
```

## Generate a QR code

To generate a QR code, send a POST request to http://localhost:8080/generate with two POST variables:

- **size**: This sets the width and height of the QR code
- **url**: This is the URL that the QR code will embed

The curl example, below, shows how to create a QR code 256x256 px that embeds "https://arstechnica.com", and outputs the generated QR code to _data/qrcode.png_.

```bash
curl -X POST \
    --form "size=256" \
    --form "url=https://arstechnica.com" \
    --output data/qrcode.png \
    http://localhost:8080/generate
```

You can also watermark the QR code, by uploading a PNG file using the `watermark` POST variable.
Below is an example of how to do so with curl.

```bash
curl -X POST \
    --form "size=256" \
    --form "url=https://matthewsetter.com" \
    --form "watermark=@data/twilio-logo.png" \
    --output data/qrcode.png \
    http://localhost:8080/generate
```

[tutorial-url]: https://www.twilio.com/blog/generate-qr-code-with-go
[go-url]: https://go.dev/
[curl-url]: https://curl.se/
[postman-url]: https://www.postman.com/downloads/
