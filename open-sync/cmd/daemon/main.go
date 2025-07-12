package main

import (
	"flag"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	pb "github.com/rudransh-shrivastava/open-sync/open-sync/pkg/protobuf"
	"google.golang.org/protobuf/proto"
)

const chunkSize = 1024 // 1KB

type fileTransfer struct {
	file           *os.File
	totalChunks    int64
	receivedChunks map[int64]bool
	lastPacketTime time.Time
}

func udpListener(conn *net.UDPConn) {
	transfers := make(map[string]*fileTransfer)
	buffer := make([]byte, 65507) // Max UDP packet size

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			continue
		}

		var packet pb.Packet
		if err := proto.Unmarshal(buffer[:n], &packet); err != nil {
			log.Printf("Failed to unmarshal packet from %s: %v", remoteAddr, err)
			continue
		}

		switch p := packet.Payload.(type) {
		case *pb.Packet_Metadata:
			md := p.Metadata
			log.Printf("Received metadata for %s from %s", md.FileName, remoteAddr)

			filePath := filepath.Join("downloads", md.FileName)
			file, err := os.Create(filePath)
			if err != nil {
				log.Printf("Failed to create file %s: %v", filePath, err)
				continue
			}
			if err := file.Truncate(md.FileSize); err != nil {
				log.Printf("Failed to pre-allocate file size for %s: %v", filePath, err)
				file.Close()
				continue
			}

			transfers[remoteAddr.String()] = &fileTransfer{
				file:           file,
				totalChunks:    md.TotalChunks,
				receivedChunks: make(map[int64]bool),
				lastPacketTime: time.Now(),
			}

		case *pb.Packet_Chunk:
			chunk := p.Chunk
			transfer, ok := transfers[remoteAddr.String()]
			if !ok {
				log.Printf("Received chunk from %s without metadata", remoteAddr)
				continue
			}

			offset := chunk.SequenceNumber * chunkSize
			_, err := transfer.file.WriteAt(chunk.Data, offset)
			if err != nil {
				log.Printf("Failed to write chunk %d for %s: %v", chunk.SequenceNumber, transfer.file.Name(), err)
				continue
			}

			transfer.receivedChunks[chunk.SequenceNumber] = true
			transfer.lastPacketTime = time.Now()

			if int64(len(transfer.receivedChunks)) == transfer.totalChunks {
				log.Printf("Finished receiving %s from %s", transfer.file.Name(), remoteAddr)
				transfer.file.Close()
				delete(transfers, remoteAddr.String())
			}
		}
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 10 * 1024 * 1024 * 1024 // 10GB
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	receiverAddr := r.FormValue("receiver_addr")
	if receiverAddr == "" {
		http.Error(w, "Missing receiver_addr", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	udpAddr, err := net.ResolveUDPAddr("udp", receiverAddr)
	if err != nil {
		http.Error(w, "Invalid receiver address: "+err.Error(), http.StatusBadRequest)
		return
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		http.Error(w, "Failed to connect to receiver: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	fileSize := int64(len(fileBytes))
	totalChunks := int64(math.Ceil(float64(fileSize) / float64(chunkSize)))

	metadata := &pb.Packet{
		Payload: &pb.Packet_Metadata{
			Metadata: &pb.Metadata{
				FileName:    header.Filename,
				FileSize:    fileSize,
				TotalChunks: totalChunks,
			},
		},
	}
	metaData, err := proto.Marshal(metadata)
	if err != nil {
		http.Error(w, "Failed to marshal metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = conn.Write(metaData)
	if err != nil {
		http.Error(w, "Failed to send metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for i := int64(0); i < totalChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > fileSize {
			end = fileSize
		}

		chunk := &pb.Packet{
			Payload: &pb.Packet_Chunk{
				Chunk: &pb.Chunk{
					SequenceNumber: i,
					Data:           fileBytes[start:end],
				},
			},
		}
		chunkData, err := proto.Marshal(chunk)
		if err != nil {
			log.Printf("Failed to marshal chunk %d: %v", i, err)
			continue
		}
		_, err = conn.Write(chunkData)
		if err != nil {
			log.Printf("Failed to send chunk %d: %v", i, err)
		}
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status": "transfer initiated"}`))
}

func main() {
	apiPort := flag.String("api-port", "8080", "Port for the HTTP API server")
	udpPort := flag.String("udp-port", "8081", "Port for the UDP file transfer server")
	flag.Parse()

	if err := os.MkdirAll("./downloads", os.ModePerm); err != nil {
		log.Fatal(err)
	}

	addr, err := net.ResolveUDPAddr("udp", ":"+*udpPort)
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on UDP port: %v", err)
	}
	defer conn.Close()

	go udpListener(conn)

	http.HandleFunc("/upload", uploadHandler)
	log.Printf("API server started on :%s", *apiPort)
	log.Fatal(http.ListenAndServe(":"+*apiPort, nil))
}

