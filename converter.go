package main

import (
	"fmt"
	"github.com/Tnze/go-mc/nbt"
	"os"
	"path"
	"strings"
	"time"
)

const outputFolder = "./output"

// Chunk represents a Minecraft chunk with x, z coordinates and time
type Chunk struct {
	X           int32  `nbt:"x"`
	Z           int32  `nbt:"z"`
	Time        int64  `nbt:"time"`
	Forceloaded *int64 `nbt:"forceloaded"`
}

type MemberData struct {
	// TODO implement this
}

// ChunkData represents the structure of NBT data to hold chunk coordinates
type ChunkData struct {
	MaxClaimChunks     int32              `nbt:"max_claim_chunks"`
	MaxForceLoadChunks int32              `nbt:"max_force_load_chunks"`
	LastLoginTime      int64              `nbt:"last_login_time"`
	Chunks             map[string][]Chunk `nbt:"chunks"`
	MemberData         MemberData         `nbt:"member_data"`
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
		Chunks:             make(map[string][]Chunk, len(o.Dimensions)),
	}
	for i, dim := range o.Dimensions {
		c.Chunks[i] = make([]Chunk, 0)
		for _, cl := range dim.Claims {
			for _, pos := range cl.Claims {
				c.Chunks[i] = append(c.Chunks[i], Chunk{
					X:    pos.X,
					Z:    pos.Y,
					Time: time.Now().Unix(),
				})
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
	return os.WriteFile(filename, []byte(snbt), 0644)
}

func handleFile(filename string) {
	// Read the NBT data
	chunkData, err := ReadNBT(filename)
	if err != nil {
		fmt.Println("Error reading NBT:", err)
		return
	}

	if len(chunkData.Chunks) == 0 {
		fmt.Println("Invalid NBT file")
		return
	}

	// Convert the chunk data to SNBT format
	snbt, err := ConvertToSNBT(chunkData)
	if err != nil {
		fmt.Println("Error converting to SNBT:", err)
		return
	}

	fileNameClean := strings.Replace(strings.Replace(filename, ".nbt", ".snbt", 1), "./player-claims", "", 1)

	if strings.Contains(fileNameClean, "00000000-0000-0000-0000-") {
		// Server file, hardcoded
		fileNameClean = "30be7d9a-1adb-4a32-b0f0-50fdde3c0dc6.snbt"

		basePath := path.Join(outputFolder, "ftbteams/server")
		err = os.MkdirAll(basePath, os.ModePerm)

		if err != nil {
			fmt.Println("Error making server directories:", err)
			return
		}

		serverContent := `{
	id: "30be7d9a-1adb-4a32-b0f0-50fdde3c0dc6"
	type: "server"
	ranks: { }
	properties: {
		"ftbchunks:allow_explosions": 0b
		"ftbchunks:allow_mob_griefing": 0b
		"ftbchunks:allow_fake_players": 0b
		"ftbchunks:allow_named_fake_players": [ ]
		"ftbchunks:allow_fake_players_by_id": 1b
		"ftbchunks:allow_pvp": 1b
		"ftbchunks:block_edit_and_interact_mode": "allies"
		"ftbchunks:entity_interact_mode": "allies"
		"ftbchunks:nonliving_entity_attack_mode": "allies"
		"ftbchunks:claim_visibility": "public"
		"ftbchunks:location_mode": "allies"
		"ftbteams:display_name": "server"
		"ftbteams:description": ""
		"ftbteams:color": "#59C6FF"
		"ftbteams:free_to_join": 0b
		"ftbteams:max_msg_history_size": 1000
	}
	message_history: [ ]
	extra: { }
}`

		err = WriteSNBT(fmt.Sprintf("%s/%s", basePath, fileNameClean), serverContent)
		if err != nil {
			fmt.Println("Error writing server SNBT:", err)
			return
		}
	}

	basePath := path.Join(outputFolder, "ftbchunks")
	err = os.MkdirAll(basePath, os.ModePerm)

	if err != nil {
		fmt.Println("Error making output directories:", err)
		return
	}

	// Write the SNBT to a file
	err = WriteSNBT(fmt.Sprintf("%s/%s", basePath, fileNameClean), snbt)
	if err != nil {
		fmt.Println("Error writing SNBT:", err)
		return
	}

	fmt.Println("Successfully converted NBT to SNBT!")
}

func main() {
	err := os.RemoveAll(path.Join(outputFolder, "ftbchunks"))
	if err != nil {
		fmt.Println("Failed to remove ftbchunks")
		return
	}
	err = os.RemoveAll(path.Join(outputFolder, "ftbteams"))
	if err != nil {
		fmt.Println("Failed to remove ftbteams")
		return
	}
	// The player-claims directory contains OpenPac claim data
	items, _ := os.ReadDir("./player-claims")
	for _, item := range items {
		if item.IsDir() {
			fmt.Println("Unexpected directory", item.Name())
		} else {
			// handle file there
			handleFile(fmt.Sprintf("%s/%s", "./player-claims", item.Name()))
		}
	}
}
