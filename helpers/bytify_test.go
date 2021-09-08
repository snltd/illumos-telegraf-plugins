package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBytify(t *testing.T) {
	t.Parallel()

	tables := []struct {
		in  string
		out float64
	}{
		{"-", 0},
		{"5", 5},
		{"15b", 15},
		{"2K", 2048},
		{"2k", 2048},
		{"2.5K", 2560},
		{"800M", 838860800},
		{"6.12G", 6571299962.88},
		{"0.5T", 549755813888},
	}

	for _, table := range tables {
		result, _ := Bytify(table.in)
		require.Equal(t, table.out, result)
	}
}

func TestBytifyI(t *testing.T) {
	t.Parallel()

	tables := []struct {
		in  string
		out float64
	}{
		{"-", 0},
		{"5", 5},
		{"15b", 15},
		{"2K", 2000},
		{"2.5K", 2500},
		{"800M", 800000000},
		{"-6.12G", -6120000000},
		{"0.5T", 500000000000},
	}

	for _, table := range tables {
		result, _ := BytifyI(table.in)
		require.Equal(t, table.out, result)
	}
}

func TestUnBytify(t *testing.T) {
	t.Parallel()

	tables := []struct {
		in  float64
		out string
	}{
		{15, "15b"},
		{2048, "2.0Kb"},
		{2560, "2.5Kb"},
		{838860800, "800.0Mb"},
		{6120000000, "5.7Gb"},
	}

	for _, table := range tables {
		result := UnBytify(table.in)
		require.Equal(t, table.out, result)
	}
}

func TestUnBytifyI(t *testing.T) {
	t.Parallel()

	tables := []struct {
		in  float64
		out string
	}{
		{15, "15b"},
		{2000, "2.0Kib"},
		{2560, "2.6Kib"},
		{800000000, "800.0Mib"},
		{-6120000000, "-6.1Gib"},
		{500000000000, "500.0Gib"},
	}

	for _, table := range tables {
		result := UnBytifyI(table.in)
		require.Equal(t, table.out, result)
	}
}
