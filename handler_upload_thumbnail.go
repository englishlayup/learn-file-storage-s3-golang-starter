package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
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

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	mediaType := header.Header.Get("Content-Type")
	imageData, _ := io.ReadAll(file)
	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Video does not exist", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Authenticated user is not the video owner", err)
		return
	}

	thumbnailURL := fmt.Sprintf("data:%s;base64,%s", mediaType, base64.StdEncoding.EncodeToString(imageData))
	video.ThumbnailURL = &thumbnailURL

	cfg.db.UpdateVideo(video)

	respondWithJSON(w, http.StatusOK, database.Video{
		ID:                video.ID,
		ThumbnailURL:      video.ThumbnailURL,
		CreatedAt:         video.CreatedAt,
		UpdatedAt:         video.UpdatedAt,
		VideoURL:          video.VideoURL,
		CreateVideoParams: video.CreateVideoParams,
	})
}
