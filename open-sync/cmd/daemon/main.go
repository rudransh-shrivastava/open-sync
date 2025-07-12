package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 10 * 1024 * 1024 * 1024 // 10GB
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	dst, err := os.Create("./uploads/" + header.Filename)
	if err != nil {
		http.Error(w, "Failed to create file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status": "success", "filename": "` + header.Filename + `"}`))
}

func main() {
	if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/upload", uploadHandler)
	log.Println("Server started on :8080 (supports large file uploads)")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
