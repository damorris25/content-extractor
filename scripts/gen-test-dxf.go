//go:build ignore

// gen-test-dxf generates minimal DXF test fixtures.
package main

import (
	"os"
	"path/filepath"
)

func main() {
	dir := "testdata"
	os.MkdirAll(dir, 0o755)

	createSimpleDXF(filepath.Join(dir, "simple.dxf"))
}

func createSimpleDXF(path string) {
	// Minimal DXF with HEADER, TABLES (layers), and ENTITIES (text).
	dxf := `  0
SECTION
  2
HEADER
  9
$ACADVER
  1
AC1027
  9
$INSUNITS
 70
4
  0
ENDSEC
  0
SECTION
  2
TABLES
  0
TABLE
  2
LAYER
  0
LAYER
  2
Walls
 70
0
  0
LAYER
  2
Doors
 70
0
  0
LAYER
  2
Electrical
 70
0
  0
ENDTAB
  0
ENDSEC
  0
SECTION
  2
ENTITIES
  0
TEXT
  1
Floor Plan - Level 1
  0
TEXT
  1
Main Entrance
  0
MTEXT
  1
Conference Room\PCapacity: 20
  0
TEXT
  1
Emergency Exit
  0
INSERT
  2
DoorBlock
  0
DIMENSION
  1
12.5m
  0
ENDSEC
  0
EOF
`
	os.WriteFile(path, []byte(dxf), 0o644)
}
