package server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/agl/ewhales-v1-exporter/internal/config"
	"github.com/agl/ewhales-v1-exporter/internal/models"
	"github.com/agl/ewhales-v1-exporter/internal/serialize"
	pb "github.com/agl/ewhales-v1-exporter/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type exportServer struct {
	pb.UnimplementedExporterServiceServer
	cfg *config.Config
}

func (s *exportServer) ExportData(stream pb.ExporterService_ExportDataServer) error {
	log.Println("New export stream connected")
	
	// Receive first message which must be metadata
	req, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		log.Printf("Error receiving first message: %v", err)
		return err
	}

	metadata := req.GetMetadata()
	if metadata == nil {
		log.Println("First message was not metadata")
		return fmt.Errorf("first message must be metadata")
	}

	expectedLogbooks := int(metadata.ExpectedLogbooks)
	expectedEntries := int(metadata.ExpectedEntries)
	log.Printf("Expecting %d logbooks and %d entries", expectedLogbooks, expectedEntries)

	var pivotData models.PivotData

	// Receive the rest of the stream
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error receiving stream: %v", err)
			return err
		}

		batch := req.GetBatch()
		if batch != nil {
			for _, lb := range batch.Logbooks {
				pivotData.Logbooks = append(pivotData.Logbooks, models.Logbook{
					PostID:    uint(lb.PostId),
					LogbookID: lb.LogbookId,
				})
			}
			for _, entry := range batch.Entries {
				pivotData.LogbookEntries = append(pivotData.LogbookEntries, models.LogbookEntry{
					PostID:        uint(entry.PostId),
					LogbookID:     uint(entry.LogbookId),
					Bottom:        entry.Bottom,
					CloudCover:    entry.CloudCover,
					Depth:         entry.Depth,
					DepthUnit:     entry.DepthUnit,
					EntryDate:     entry.EntryDate,
					Landmark:      entry.Landmark,
					Latitude:      entry.Latitude,
					LocalTime:     entry.LocalTime,
					Longitude:     entry.Longitude,
					Page:          entry.Page,
					SeaState:      entry.SeaState,
					ShipHeading:   entry.ShipHeading,
					ShipSightings: entry.ShipSightings,
					Weather:       entry.Weather,
					WindDirection: entry.WindDirection,
					WindForce:     entry.WindForce,
				})
			}
		}
	}

	log.Printf("Received total %d logbooks and %d entries", len(pivotData.Logbooks), len(pivotData.LogbookEntries))

	// Serialize data to CSV
	serializer := &serialize.CSVSerializer{
		LogbooksFile:       "logbooks_" + s.cfg.CSVBaseName,
		LogbookEntriesFile: s.cfg.CSVBaseName,
		IdsToFields:        s.cfg.IdsToFields,
	}

	log.Println("Serializing to CSV...")
	err = serializer.Serialize(pivotData, nil)
	if err != nil {
		log.Printf("Error serializing: %v", err)
		return stream.SendAndClose(&pb.ExportResponse{
			Success: false,
			Message: fmt.Sprintf("failed to serialize: %v", err),
		})
	}

	log.Println("Successfully saved CSV files")
	return stream.SendAndClose(&pb.ExportResponse{
		Success: true,
		Message: "Successfully received and saved data",
	})
}

// GenerateSelfSignedCert generates a self-signed TLS cert for development
func GenerateSelfSignedCert(certPath, keyPath string) error {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"eWHALES Exporter"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	certOut, err := os.Create(certPath)
	if err != nil {
		return err
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	keyOut, err := os.Create(keyPath)
	if err != nil {
		return err
	}
	defer keyOut.Close()
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return nil
}

// RunServer starts the gRPC server
func RunServer(cfg *config.Config) error {
	if _, err := os.Stat(cfg.TLSCertFile); os.IsNotExist(err) {
		log.Println("TLS certs not found, generating self-signed certificates...")
		if err := GenerateSelfSignedCert(cfg.TLSCertFile, cfg.TLSKeyFile); err != nil {
			return fmt.Errorf("failed to generate certs: %v", err)
		}
		
		absCert, _ := filepath.Abs(cfg.TLSCertFile)
		absKey, _ := filepath.Abs(cfg.TLSKeyFile)
		log.Printf("Successfully generated and saved self-signed certificate at: %s", absCert)
		log.Printf("Successfully generated and saved self-signed private key at: %s", absKey)
	}

	creds, err := credentials.NewServerTLSFromFile(cfg.TLSCertFile, cfg.TLSKeyFile)
	if err != nil {
		return fmt.Errorf("failed to load TLS keys: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ListenPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterExporterServiceServer(grpcServer, &exportServer{cfg: cfg})

	log.Printf("Server listening on port %d", cfg.ListenPort)
	return grpcServer.Serve(lis)
}
