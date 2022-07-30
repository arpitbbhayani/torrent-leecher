package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
)

// https://wiki.theory.org/BitTorrentSpecification
type FileInfo struct {
	PieceLength int64
	Pieces      [][]byte
	Length      int64
	Name        string
}

type Torrent struct {
	Info     FileInfo
	Announce string
}

func BDecode(reader *bufio.Reader) (interface{}, error) {
	ch, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch ch {

	// integer
	case 'i':
		var buffer []byte
		for {
			ch, err = reader.ReadByte()
			if err != nil {
				return nil, err
			}

			// if we stumble upon `e`, dict complete, and we return
			if ch == 'e' {
				value, err := strconv.ParseInt(string(buffer), 10, 64)
				if err != nil {
					panic(err)
				}
				return value, nil
			}
			buffer = append(buffer, ch)
		}

	// list
	case 'l':
		var listHolder []interface{}
		for {
			ch, err = reader.ReadByte()
			if err != nil {
				return nil, err
			}

			// if we stumble upon `e`, list complete, and we return
			if ch == 'e' {
				return listHolder, nil
			}

			// read the key
			reader.UnreadByte()
			data, err := BDecode(reader)
			if err != nil {
				return nil, err
			}

			// put key and value in dictionary
			listHolder = append(listHolder, data)
		}

	// dictioanry
	case 'd':
		dictHolder := map[string]interface{}{}
		for {
			ch, err = reader.ReadByte()
			if err != nil {
				return nil, err
			}

			// if we stumble upon `e`, dict complete, and we return
			if ch == 'e' {
				return dictHolder, nil
			}

			// read the key
			reader.UnreadByte()
			data, err := BDecode(reader)
			if err != nil {
				return nil, err
			}

			// key has to be a string, if not then error
			key, ok := data.(string)
			if !ok {
				return nil, errors.New("key of the dictionary is not string")
			}

			// read the value
			value, err := BDecode(reader)
			if err != nil {
				return nil, err
			}

			// if key == "announce" || key == "announce-list" || key == "comment" || key == "created by" || key == "creation date" || key == "length" || key == "name" || key == "piece length" {
			// 	fmt.Println(value)
			// }

			// put key and value in dictionary
			dictHolder[key] = value
		}

	// string
	default:
		reader.UnreadByte()

		var lengthBuf []byte
		for {
			ch, err := reader.ReadByte()
			if err != nil {
				panic(err)
			}
			if ch == ':' {
				break
			}
			lengthBuf = append(lengthBuf, ch)
		}

		length, err := strconv.Atoi(string(lengthBuf))
		if err != nil {
			panic(err)
		}

		var strBuf []byte
		for i := 0; i < length; i++ {
			ch, err := reader.ReadByte()
			if err != nil {
				panic(err)
			}
			strBuf = append(strBuf, ch)
		}

		return string(strBuf), nil
	}
}

func batch(data []byte, batch int) [][]byte {
	var result [][]byte
	for i := 0; i < len(data); i += batch {
		end := i + batch
		if end > len(data) {
			end = len(data)
		}
		result = append(result, data[i:end])
	}
	return result
}

func ParseTorrent(reader *bufio.Reader) Torrent {
	data, err := BDecode(reader)
	if err != nil {
		panic(err)
	}

	tData, ok := data.(map[string]interface{})
	if !ok {
		panic("not a valid torrent file")
	}
	tInfoData, ok := tData["info"].(map[string]interface{})
	if !ok {
		panic("not a valid torrent file")
	}

	var torrentData Torrent
	torrentData.Announce = tData["announce"].(string)
	torrentData.Info = FileInfo{
		PieceLength: tInfoData["piece length"].(int64),
		Length:      tInfoData["length"].(int64),
		Name:        tInfoData["name"].(string),
		Pieces:      batch([]byte(tInfoData["pieces"].(string)), 20),
	}
	return torrentData
}

func main() {
	fp, err := os.Open("ubuntu-22.04-desktop-amd64.iso.torrent")
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	breader := bufio.NewReader(fp)
	torrentData := ParseTorrent(breader)
	fmt.Println(torrentData.Info.Length)
	fmt.Println(torrentData.Info.PieceLength)
	fmt.Println(torrentData.Info.Length / torrentData.Info.PieceLength)
	fmt.Println(len(torrentData.Info.Pieces))
}
