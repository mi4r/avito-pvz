package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	pvz_v1 "github.com/mi4r/avito-pvz/api/pvz/v1"
	"github.com/mi4r/avito-pvz/internal/storage"
	"github.com/mi4r/avito-pvz/internal/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const bufSize = 1024 * 1024

func TestServer_GetPVZList(t *testing.T) {
	// Create a mock storage
	mockStore := &mocks.Storage{}

	// Create test cases
	tests := []struct {
		name          string
		mockPVZs      []storage.PVZWithReceptions
		mockError     error
		expectedError bool
	}{
		{
			name: "successful get pvz list",
			mockPVZs: []storage.PVZWithReceptions{
				{
					PVZ: storage.PVZ{
						ID:               uuid.New(),
						RegistrationDate: time.Now(),
						City:             "Moscow",
					},
				},
				{
					PVZ: storage.PVZ{
						ID:               uuid.New(),
						RegistrationDate: time.Now().Add(-24 * time.Hour),
						City:             "Saint Petersburg",
					},
				},
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "storage error",
			mockPVZs:      nil,
			mockError:     assert.AnError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			mockStore.ExpectedCalls = nil
			mockStore.On("GetPVZsWithReceptions", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(tt.mockPVZs, tt.mockError)

			// Create server
			server := NewServer(mockStore)

			// Create a test context
			ctx := context.Background()

			// Call the method
			resp, err := server.GetPVZList(ctx, &pvz_v1.GetPVZListRequest{})

			// Check error
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Check response
			assert.Equal(t, len(tt.mockPVZs), len(resp.Pvzs))
			for i, pvz := range tt.mockPVZs {
				assert.Equal(t, pvz.PVZ.ID.String(), resp.Pvzs[i].Id)
				assert.Equal(t, pvz.PVZ.City, resp.Pvzs[i].City)
				assert.True(t, timestamppb.New(pvz.PVZ.RegistrationDate).AsTime().Equal(resp.Pvzs[i].RegistrationDate.AsTime()))
			}
		})
	}
}

func TestServer_Start(t *testing.T) {
	// Create a mock storage
	mockStore := &mocks.Storage{}

	// Setup mock expectations
	mockStore.On("GetPVZsWithReceptions", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]storage.PVZWithReceptions{}, nil)

	// Create a test server
	server := NewServer(mockStore)

	// Create a listener
	lis := bufconn.Listen(bufSize)

	// Start the server in a goroutine
	go func() {
		grpcServer := grpc.NewServer()
		pvz_v1.RegisterPVZServiceServer(grpcServer, server)
		if err := grpcServer.Serve(lis); err != nil {
			t.Errorf("Server exited with error: %v", err)
		}
	}()

	// Create a client connection
	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}), grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	// Create a client
	client := pvz_v1.NewPVZServiceClient(conn)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Call the method
	resp, err := client.GetPVZList(ctx, &pvz_v1.GetPVZListRequest{})
	require.NoError(t, err)
	assert.Empty(t, resp.Pvzs)
}
