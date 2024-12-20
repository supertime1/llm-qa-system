package pg

import (
	"fmt"
	"llm-qa-system/backend-service/src/proto"
	"math/big"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ToUUID converts a Postgres UUID to a uuid.UUID
func ToUUID(id pgtype.UUID) uuid.UUID {
	return uuid.UUID(id.Bytes)
}

// NewUUID generates a new random GUID
func NewUUID() pgtype.UUID {
	var id pgtype.UUID
	uid := uuid.New()
	id.Bytes = uid
	id.Valid = true
	return id
}

// ToPGUUID converts a protobuf uuid to a postgres uuid
func ToPGUUID(pbUUID *proto.UUID) pgtype.UUID {
	var id pgtype.UUID
	if pbUUID == nil {
		return id
	}
	id.Bytes = [16]byte(pbUUID.Value)
	id.Valid = true
	return id
}

// UUIDToPGUUID converts uuid.UUID to PGUUID
func UUIDToPGUUID(uid uuid.UUID) pgtype.UUID {
	var id pgtype.UUID
	if uid == uuid.Nil {
		return id
	}
	id.Bytes = uid
	id.Valid = true
	return id
}

// ToPGUUIDs converts a protobuf uuids to a postgres uuids
func ToPGUUIDs(pbUUIDs []*proto.UUID) []pgtype.UUID {
	r := make([]pgtype.UUID, 0, len(pbUUIDs))
	for _, pbUUID := range pbUUIDs {
		r = append(r, ToPGUUID(pbUUID))
	}
	return r
}

// ToInt2 -
func ToInt2(i int) pgtype.Int2 {
	return pgtype.Int2{
		Int16: int16(i),
		Valid: true,
	}
}

// ToInt4 -
func ToInt4(i int) pgtype.Int4 {
	return pgtype.Int4{
		Int32: int32(i),
		Valid: true,
	}
}

// ParseUUID parses a string formatted UUID and returns a UUID with error
func ParseUUID(uidString string) (pgtype.UUID, error) {
	uid, err := uuid.Parse(uidString)
	id := pgtype.UUID{
		Bytes: uid,
		Valid: err == nil,
	}
	return id, err
}

// TimestampCmp performs a standard compare returning the lowest timestamp first
func TimestampCmp(a, b pgtype.Timestamp) int {
	return a.Time.Compare(b.Time)
}

// TimestampReverseCmp performs a standard compare returning the lowest timestamp first
func TimestampReverseCmp(a, b pgtype.Timestamp) int {
	return b.Time.Compare(a.Time)
}

func Uint64ToNumeric(val uint64) pgtype.Numeric {
	// Convert uint64 to string
	uint64Str := fmt.Sprintf("%d", val)

	// Parse the string into a big.Int
	bigIntVal := new(big.Int)
	bigIntVal.SetString(uint64Str, 10) // Base 10

	// Create a pgtype.Numeric and set the value
	numeric := pgtype.Numeric{Int: bigIntVal, Valid: true}

	return numeric
}
