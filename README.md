# PDF2Images gRPC

This is a project that provides a gRPC-based service for converting PDF files into images. 
It leverages the power of gRPC, a high-performance, open-source framework for remote procedure calls, to offer
a seamless and efficient PDF-to-image conversion experience.

## Features

- PDF2Image Conversion: The project allows you to convert PDF files into various image formats, such as JPEG, PNG, or TIFF.
- gRPC Communication: It utilizes gRPC for communication between the client and server, ensuring fast and reliable data exchange.
- Easy Integration: With gRPC, integrating the PDF2Images service into your existing applications becomes straightforward,
enabling you to leverage its conversion capabilities seamlessly.

## Getting Started

To get started with the PDF2Images gRPC project, follow these steps:
- Clone the repository: git clone https://github.com/badfan/pdf2images
- Install the required Go dependencies.
- Build the project by running go build.

## API Reference

The project provides the following gRPC service methods:
ConvertPDF2Images: Converts a PDF file into images. It takes in the PDF file data and returns the converted images data.

The service definition and method details can be found in the proto/images.proto file.