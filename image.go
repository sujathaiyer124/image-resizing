package image

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/disintegration/imaging"
	//"github.com/sujathaiyer124/image-resizing"
)

func init() {
	functions.HTTP("ImageResize", ImagesResize)
}

// func main() {
// 	//Open the original image file
// 	r := mux.NewRouter()
// 	r.HandleFunc("/images", ImagesResize).Methods("POST")
// 	fmt.Println("Server  is getting started ....")
// 	log.Fatal(http.ListenAndServe(":8000", r))

// }

// entry point is ImageResize
func ImagesResize(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Max-Age", "3600")
	w.Header().Set("Content-Type", "application/json")
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
	projectId := "excellent-math-403109"
	topicID := "resize"

	sourceImagePath := filepath.Base(fileHeader.Filename)
	destinationBucket := "pixsell-image"
	resizedImage, err := imaging.Decode(file)
	if err != nil {
		fmt.Println("Error opening image for resizing:", err)
		return
	}
	//Here it stores in bucket
	if error := saveToBucket(resizedImage, destinationBucket, sourceImagePath); error != nil {
		log.Fatalf("Error saving image to bucket: %v", err)
	}
	json.NewEncoder(w).Encode("Image saved inside the bucket.")
	publishMessage(w, projectId, topicID)
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
func publishMessage(w io.Writer, projectID, topicID string) error {

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %w", err)
	}
	defer client.Close()

	t := client.Topic(topicID)
	result := t.Publish(ctx, &pubsub.Message{
		Data: []byte("Image saved in the bucket"),
	})
	// Block until the result is returned and a server-generated
	// ID is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("get: %w", err)
	}
	fmt.Fprintf(w, "Published message with custom attributes; msg ID: %v\n", id)
	return nil
}
