package crudp

import (
	"testing"

	"github.com/tinywasm/binary"
)

// BenchUser is a simple structure for benchmarks
type BenchUser struct {
	ID    uint32
	Name  string
	Email string
	Age   uint8
}

func (u *BenchUser) Create(data ...any) any {
	created := make([]*BenchUser, 0, len(data))
	for _, item := range data {
		user := item.(*BenchUser)
		user.ID = 123
		created = append(created, user)
	}
	return created
}

func (u *BenchUser) Read(data ...any) any {
	results := make([]*BenchUser, 0, len(data))
	for _, item := range data {
		user := item.(*BenchUser)
		results = append(results, &BenchUser{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Age:   user.Age,
		})
	}
	return results
}

func (u *BenchUser) Update(data ...any) any {
	updated := make([]*BenchUser, 0, len(data))
	for _, item := range data {
		user := item.(*BenchUser)
		user.Name = "Updated " + user.Name
		updated = append(updated, user)
	}
	return updated
}

func (u *BenchUser) Delete(data ...any) any {
	return len(data)
}

// Global variables to prevent compiler optimizations
var (
	globalCrudP     *CrudP
	globalBatchResp *BatchResponse
	globalUser      = &BenchUser{
		ID:    1,
		Name:  "BenchUser",
		Email: "bench@example.com",
		Age:   25,
	}
)

// BenchmarkCrudPSetup measures allocations for CRUDP initialization
func BenchmarkCrudPSetupShared(b *testing.B) {
	var cp *CrudP

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		cp = New(binary.Encode, binary.Decode)
		if err := cp.RegisterHandlers(&BenchUser{}); err != nil {
			b.Fatalf("RegisterHandlers failed: %v", err)
		}
	}

	globalCrudP = cp // Prevent optimization
}

// BenchmarkCrudPExecute measures allocations for executing a batch request
func BenchmarkCrudPExecuteShared(b *testing.B) {
	cp := New(binary.Encode, binary.Decode)
	if err := cp.RegisterHandlers(&BenchUser{}); err != nil {
		b.Fatalf("RegisterHandlers failed: %v", err)
	}

	var userData []byte
	binary.Encode(globalUser, &userData)
	req := &BatchRequest{
		Packets: []Packet{
			{Action: 'c', HandlerID: 0, ReqID: "bench", Data: [][]byte{userData}},
		},
	}

	var resp *BatchResponse
	var err error

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resp, err = cp.Execute(req)
		if err != nil {
			b.Fatalf("Execute failed: %v", err)
		}
	}

	globalBatchResp = resp // Prevent optimization
}
