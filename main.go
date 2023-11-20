package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"log"
	"net/http"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/disintegration/imaging"
	"github.com/gorilla/mux"
)

func main() {
	// Open the original image file
	r := mux.NewRouter()
	r.HandleFunc("/images", Image).Methods("POST")
	fmt.Println("Server  is getting started ....")
	log.Fatal(http.ListenAndServe(":8000", r))

}
func Image(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Error opening image os.Open:", err.Error())
		return
	}
	defer file.Close()
	
	sourceImagePath := filepath.Base(fileHeader.Filename)
	destinationBucket := "pixsell-image"
	destinationObjectName := "Resized image/" + sourceImagePath

	resizedImage, err := imaging.Decode(file)
	if err != nil {
		fmt.Println("Error opening image for resizing:", err)
		return
	}
	//yaha pe ek call karke vo bucket mein store kar
	if error := saveToBucket(resizedImage, destinationBucket, sourceImagePath); error != nil {
		log.Fatalf("Error saving image to bucket: %v", err)
	}
	json.NewEncoder(w).Encode("Image saved inside the bucket.")
	
	resizedImage = imaging.Resize(resizedImage, 300, 0, imaging.Lanczos)
	// Save the resized image to a new  //file idhar naya vala folder mein save kar
	if err := saveToBucket(resizedImage, destinationBucket, destinationObjectName); err != nil {
		log.Fatalf("Error saving image to bucket folder: %v", err)
	}

	log.Printf("Resized image saved to gs://%s/%s\n", destinationBucket, destinationObjectName)
	json.NewEncoder(w).Encode("Image resized and saved successfully.")
}
func saveToBucket(image image.Image, bucketName, objectName string) error {
	// Create a context and Google Cloud Storage client.
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	// Create an object handle for the destination bucket and object name.
	objectHandle := client.Bucket(bucketName).Object(objectName)

	// Create a storage.Writer for uploading the resized image.
	writer := objectHandle.NewWriter(ctx)

	// Save the resized image directly to the storage writer.
	if err := imaging.Encode(writer, image, imaging.JPEG); err != nil {
		return fmt.Errorf("error is %v", err.Error())
	}

	// Close the storage.Writer to complete the upload.
	if err := writer.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}
