package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	log.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20

	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse body as multipart/form-data", err)
		return
	}

	image, imageHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't get thumbnail", err)
		return
	}
	defer image.Close()

	mediaType, _, err := mime.ParseMediaType(imageHeader.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse header", err)
	}
	if !(mediaType == "image/jpeg" || mediaType == "image/png") {
		respondWithError(w, http.StatusUnsupportedMediaType, "Uploaded file is not an image", nil)
		return
	}

	thumbnailFilePath := cfg.createAsset(mediaType)

	thumbnailFile, err := os.Create(thumbnailFilePath)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't create thumbnail file", err)
		return
	}
	defer thumbnailFile.Close()

	_, err = io.Copy(thumbnailFile, image)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't copy thumbnail", err)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get video details from db", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Video no owned by auth'd user", nil)
		return
	}

	thumbnailFileURL := fmt.Sprintf("http://localhost:%s/%s", cfg.port, thumbnailFilePath)

	video.ThumbnailURL = &thumbnailFileURL
	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video details", err)
		return
	}

	respondWithJSON(w, http.StatusOK, struct{}{})
}
