package main

import (
	"fmt"
	"github.com/Tnze/go-mc/nbt"
	"io/ioutil"
	"os"
	"time"
)

// Chunk represents a Minecraft chunk with x, z coordinates and time
type Chunk struct {
	X    int32 `nbt:"x"`
	Z    int32 `nbt:"z"`
	Time int64 `nbt:"time"`
}

// ChunkData represents the structure of NBT data to hold chunk coordinates
type ChunkData struct {
	MaxClaimChunks     int32   `nbt:"max_claim_chunks"`
	MaxForceLoadChunks int32   `nbt:"max_force_load_chunks"`
	LastLoginTime      int64   `nbt:"last_login_time"`
	ChunksOverworld    []Chunk `nbt:"chunks"`
}

type State struct {
	Forceloaded    bool  `nbt:"forceloaded"`
	SubConfigIndex int32 `nbt:"subConfigIndex"`
}

type Position struct {
	X int32 `nbt:"x"`
	Y int32 `nbt:"y"`
}

type Claim struct {
	Claims []Position `nbt:"positions"`
	State  State      `nbt:"state"`
}

type Dimension struct {
	Claims []Claim `nbt:"claims"`
}

type OpenPaC struct {
	ConfirmedActivity int64                `nbt:"confirmedActivity"`
	Dimensions        map[string]Dimension `nbt:"dimensions"`
	Username          string               `nbt:"username"`
}

// ReadNBT reads and decodes the NBT data from a file
func ReadNBT(filename string) (*ChunkData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var chunkData OpenPaC
	_, err = nbt.NewDecoder(file).Decode(&chunkData)
	if err != nil {
		return nil, err
	}
	return convertOpenPaCToChunkData(&chunkData), nil
}

func convertOpenPaCToChunkData(o *OpenPaC) *ChunkData {
	c := ChunkData{
		MaxClaimChunks:     250,
		MaxForceLoadChunks: 2,
		LastLoginTime:      time.Now().Unix(),
		ChunksOverworld:    make([]Chunk, 0),
	}
	for i, dim := range o.Dimensions {
		for _, cl := range dim.Claims {
			for _, pos := range cl.Claims {
				if i == "minecraft:overworld" {
					c.ChunksOverworld = append(c.ChunksOverworld, Chunk{
						X:    int32(pos.X),
						Z:    int32(pos.Y),
						Time: time.Now().Unix(),
					})
				}
			}
		}
	}
	return &c
}

// ConvertToSNBT converts chunk data to SNBT (Stringified NBT) format
func ConvertToSNBT(data *ChunkData) (string, error) {
	snbtData, err := nbt.Marshal(data)
	if err != nil {
		return "", err
	}
	var s nbt.StringifiedMessage
	err = nbt.Unmarshal(snbtData, &s)
	if err != nil {
		return "", err
	}
	return string(s), nil
}

// WriteSNBT writes the SNBT string to a file
func WriteSNBT(filename string, snbt string) error {
	return ioutil.WriteFile(filename, []byte(snbt), 0644)
}

func main() {
	nbtFile := "openpac-sample.nbt"
	//snbtFile := "ftb-sample.snbt"
	newSnbtFile := "test.snbt"

	// Read the NBT data
	chunkData, err := ReadNBT(nbtFile)
	if err != nil {
		fmt.Println("Error reading NBT:", err)
		return
	}

	// Convert the chunk data to SNBT format
	snbt, err := ConvertToSNBT(chunkData)
	if err != nil {
		fmt.Println("Error converting to SNBT:", err)
		return
	}

	// Write the SNBT to a file
	err = WriteSNBT(newSnbtFile, snbt)
	if err != nil {
		fmt.Println("Error writing SNBT:", err)
		return
	}

	fmt.Println("Successfully converted NBT to SNBT!")
}
