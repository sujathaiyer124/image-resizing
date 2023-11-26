package ResizeImage

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"log"

	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/disintegration/imaging"
	//"github.com/sujathaiyer124/image-resizing"
	//"github.com/gorilla/mux"
)

type PubSubMessage struct {
	Data struct {
		Message  string `json:"message"`
		FileName string `json:"fileName"`
	} `json:"data"`
}

func init() {
	functions.CloudEvent("ResizeImageToBuckets", ResizeImageToBucket)
}

// func main() {
// //Open the original image file
// r := mux.NewRouter()
// r.HandleFunc("/images", Images).Methods("POST")
// fmt.Println("Server  is getting started ....")
// log.Fatal(http.ListenAndServe(":8000", r))

// }

// entry point is ResizeImageToBucket
func ResizeImageToBucket(ctx context.Context, m event.Event) error {

	// var data map[string]interface{}
	var pubsubMessage PubSubMessage
	data, err := base64.StdEncoding.DecodeString(string(m.Data()))
	if err != nil {
		log.Printf("Error decoding base64 data: %v", err)
		return nil
	}

	if err := json.Unmarshal(data, &pubsubMessage); err != nil {
		log.Printf("Error unmarshalling Pub/Sub message data: %v", err)
		return nil
	}
	log.Println("Data is", string(data))
	log.Printf("Unmarshalled Pub/Sub message: %+v", pubsubMessage)

	msg := pubsubMessage.Data.Message
	log.Println("The message is", msg)

	imageName := pubsubMessage.Data.FileName
	log.Println("the imagename is", imageName)

	//sourceImagePath := filepath.Base(fileHeader.Filename)
	destinationBucket := "pixsell-image"
	destinationObjectName := "Resized image/" + imageName

	image, err := downloadImage(ctx, destinationBucket, imageName)
	if err != nil {
		log.Printf("Error downloading image: %v", err)
		return nil
	}

	resizedImage := imaging.Resize(image, 250, 250, imaging.Lanczos)
	// Here it stores in bucket/folder
	if err := saveToBucket(resizedImage, destinationBucket, destinationObjectName); err != nil {
		log.Fatalf("Error saving image to bucket folder: %v", err)
	}

	log.Printf("Resized image saved to gs://%s/%s\n", destinationBucket, destinationObjectName)
	//json.NewEncoder(w).Encode("Image resized and saved successfully.")
	return nil
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
func downloadImage(ctx context.Context, bucketName, objectName string) (image.Image, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	objectHandle := client.Bucket(bucketName).Object(objectName)

	// Download the image from the bucket
	imgData, err := objectHandle.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("error reading image from bucket: %v", err)
	}
	defer imgData.Close()

	// Decode the image data
	img, _, err := image.Decode(imgData)
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %v", err)
	}

	return img, nil
}
