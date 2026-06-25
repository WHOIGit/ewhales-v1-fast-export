package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"

	"github.com/agl/ewhales-v1-exporter/internal/config"
	"github.com/agl/ewhales-v1-exporter/internal/models"
	pb "github.com/agl/ewhales-v1-exporter/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type APIClient struct {
	client pb.ExporterServiceClient
	stream pb.ExporterService_ExportDataClient
	conn   *grpc.ClientConn
}

func NewAPIClient(cfg *config.Config) (*APIClient, error) {
	// For testing with self-signed certs, we use InsecureSkipVerify
	// In production with Let's Encrypt, this should be false or use proper RootCAs
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	creds := credentials.NewTLS(tlsConfig)

	conn, err := grpc.Dial(cfg.ServerAddress, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %v", err)
	}

	client := pb.NewExporterServiceClient(conn)
	return &APIClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *APIClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *APIClient) StartStream(expectedLogbooks, expectedEntries int) error {
	stream, err := c.client.ExportData(context.Background())
	if err != nil {
		return fmt.Errorf("failed to start stream: %v", err)
	}
	c.stream = stream

	// Send metadata
	err = c.stream.Send(&pb.ExportRequest{
		Payload: &pb.ExportRequest_Metadata{
			Metadata: &pb.ExportMetadata{
				ExpectedLogbooks: int64(expectedLogbooks),
				ExpectedEntries:  int64(expectedEntries),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send metadata: %v", err)
	}

	return nil
}

func (c *APIClient) SendBatch(data *models.PivotData) error {
	if c.stream == nil {
		return fmt.Errorf("stream not started")
	}

	const maxChunkSize = 2000

	// Send logbooks in chunks
	for i := 0; i < len(data.Logbooks); i += maxChunkSize {
		end := i + maxChunkSize
		if end > len(data.Logbooks) {
			end = len(data.Logbooks)
		}

		batch := &pb.ExportBatch{
			Logbooks: data.Logbooks[i:end],
		}

		if err := c.stream.Send(&pb.ExportRequest{
			Payload: &pb.ExportRequest_Batch{
				Batch: batch,
			},
		}); err != nil {
			return fmt.Errorf("failed to send logbooks batch: %v", err)
		}
	}

	// Send entries in chunks
	for i := 0; i < len(data.LogbookEntries); i += maxChunkSize {
		end := i + maxChunkSize
		if end > len(data.LogbookEntries) {
			end = len(data.LogbookEntries)
		}

		batch := &pb.ExportBatch{
			Entries: data.LogbookEntries[i:end],
		}

		if err := c.stream.Send(&pb.ExportRequest{
			Payload: &pb.ExportRequest_Batch{
				Batch: batch,
			},
		}); err != nil {
			return fmt.Errorf("failed to send entries batch: %v", err)
		}
	}

	return nil
}

func (c *APIClient) FinishStream() error {
	if c.stream == nil {
		return nil
	}

	resp, err := c.stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("failed to close stream and receive response: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("server reported failure: %s", resp.Message)
	}

	log.Printf("Server response: %s", resp.Message)
	return nil
}
