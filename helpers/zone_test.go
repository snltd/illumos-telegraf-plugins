package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestZoneMapNames(t *testing.T) {
	t.Parallel()

	zoneMap := ParseZones(zoneadmOutput)
	require.ElementsMatch(t, []string{"global", "cube-media", "cube-ws"}, zoneMap.Names())
}

func TestZoneMapRunning(t *testing.T) {
	t.Parallel()

	zoneMap := ParseZones(zoneadmOutput)
	require.ElementsMatch(t, []string{"global", "cube-media"}, zoneMap.InState("running"))
	require.ElementsMatch(t, []string{"cube-ws"}, zoneMap.InState("installed"))
	require.ElementsMatch(t, []string{}, zoneMap.InState("configured"))
}

func TestParseZone(t *testing.T) {
	t.Parallel()

	result, err := parseZone("0:global:running:/::ipkg:shared:0")
	require.Nil(t, err)

	require.Equal(
		t,
		Zone{0, "global", "running", "/", "", "ipkg", "shared", 0},
		result,
	)

	result, err = parseZone(
		"42:mz1:running:/zones/mz1:c624d04f-d0d9-e1e6-822e-acebc78ec9ff:lipkg:excl:128",
	)

	require.Nil(t, err)
	require.Equal(
		t,
		Zone{
			42,
			"mz1",
			"running",
			"/zones/mz1",
			"c624d04f-d0d9-e1e6-822e-acebc78ec9ff",
			"lipkg",
			"excl",
			128,
		},
		result,
	)

	result, err = parseZone("some:random:string")

	require.Equal(t, Zone{}, result)
	require.Error(t, err)
}

func TestParseZones(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		ZoneMap{
			"global": Zone{0, "global", "running", "/", "", "ipkg", "shared", 0},
			"cube-media": Zone{
				42,
				"cube-media",
				"running",
				"/zones/cube-media",
				"c624d04f-d0d9-e1e6-822e-acebc78ec9ff",
				"lipkg",
				"excl",
				128,
			},
			"cube-ws": Zone{
				44,
				"cube-ws",
				"installed",
				"/zones/cube-ws",
				"0f9c56f4-9810-6d45-f801-d34bf27cc13f",
				"pkgsrc",
				"excl",
				179,
			},
		},
		ParseZones(zoneadmOutput),
	)
}

func TestZoneByID(t *testing.T) {
	t.Parallel()

	zoneMap := ParseZones(zoneadmOutput)
	require.ElementsMatch(t, []string{"global", "cube-media", "cube-ws"}, zoneMap.Names())

	zoneData, err := zoneMap.ZoneByID(42)

	require.Nil(t, err)
	require.Equal(
		t,
		Zone{
			ID:      42,
			Name:    "cube-media",
			Status:  "running",
			Path:    "/zones/cube-media",
			UUID:    "c624d04f-d0d9-e1e6-822e-acebc78ec9ff",
			Brand:   "lipkg",
			IPType:  "excl",
			DebugID: 128,
		},
		zoneData)

	_, err = zoneMap.ZoneByID(101)
	require.Error(t, err)
}

func TestParseZoneVnics(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		ZoneVnicMap{
			"www_records0": Vnic{
				Name:  "www_records0",
				Zone:  "cube-www-records",
				Link:  "rge0",
				Speed: 1000,
			},
		},
		ParseZoneVnics("www_records0:cube-www-records:rge0:1000"),
	)

	require.Equal(
		t,
		ZoneVnicMap{},
		ParseZoneVnics(""),
	)
}

func TestParseZoneVnic(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		Vnic{
			Name:  "www_records0",
			Zone:  "cube-www-records",
			Link:  "rge0",
			Speed: 1000,
		},
		parseZoneVnic("www_records0:cube-www-records:rge0:1000"),
	)
}

var zoneadmOutput = `0:global:running:/::ipkg:shared:0
42:cube-media:running:/zones/cube-media:c624d04f-d0d9-e1e6-822e-acebc78ec9ff:lipkg:excl:128
44:cube-ws:installed:/zones/cube-ws:0f9c56f4-9810-6d45-f801-d34bf27cc13f:pkgsrc:excl:179
-:gold-lipkg-b:installed:/zones/gold-lipkg-b:4e9ce283-b99b-47a4-b8f1-dfc09b3cfd99:lipkg:excl:219
`
