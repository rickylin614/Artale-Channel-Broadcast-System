package parse

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const MAX_VAL_LEN = 256

var KNOWN = map[string]bool{
	"Nickname":    true,
	"Channel":     true,
	"Text":        true,
	"Type":        true,
	"ProfileCode": true,
	"UserId":      true,
}

// ---------- ChatParser ----------
func parseStruct(data []byte) map[string]interface{} {
	out := map[string]interface{}{}
	colors := []string{}
	L := len(data)
	i := 0

	for i+4 <= L {
		if i+4 > L {
			break
		}
		nameLen := int(binary.LittleEndian.Uint32(data[i : i+4]))
		// basic sanity checks
		if !(nameLen > 0 && nameLen <= 64) || i+4+nameLen+6 > L {
			i++
			continue
		}

		nameBytes := data[i+4 : i+4+nameLen]
		// decode ascii name
		name := ""
		if utf8.Valid(nameBytes) {
			name = string(nameBytes)
		} else {
			// fallback ascii-like
			name = string(bytes.Map(func(r rune) rune {
				if r < 128 {
					return r
				}
				return '?'
			}, nameBytes))
		}

		cur := i + 4 + nameLen
		if cur+6 > L {
			i++
			continue
		}
		typeTag := int(binary.LittleEndian.Uint16(data[cur : cur+2]))
		valLen := int(binary.LittleEndian.Uint32(data[cur+2 : cur+6]))
		vStart := cur + 6
		vEnd := vStart + valLen

		if vEnd > L || valLen > MAX_VAL_LEN {
			i++
			continue
		}

		if name != "Channel" {
			if KNOWN[name] {
				// in original code only type_tag==4 were taken as utf8
				if typeTag == 4 {
					valBytes := data[vStart:vEnd]
					// try utf-8 decode, fallback replacement
					if utf8.Valid(valBytes) {
						out[name] = string(valBytes)
					} else {
						// replace invalids
						out[name] = string(bytes.Runes(bytes.ReplaceAll(valBytes, []byte{0xff, 0xfd}, []byte("?"))))
					}
				}
			} else if strings.HasPrefix(name, "#") && nameLen == 7 {
				colors = append(colors, name)
			}
		}
		i = vEnd
	}

	if len(colors) > 0 {
		out["color1"] = colors[0]
	}
	if len(colors) > 1 {
		out["color2"] = colors[1]
	}

	out["timestamp"] = time.Now().Format("2006-01-02 15:04:05")

	// scan for Channel pattern: 0x02 xx xx xx xx 0x04
	for k := 0; k+6 <= L; k++ {
		if data[k] == 0x02 && data[k+5] == 0x04 {
			val := int(binary.LittleEndian.Uint32(data[k+1 : k+5]))
			if val >= 1 && val <= 9999 {
				out["Channel"] = fmt.Sprintf("CH%d", val)
				break
			}
		}
	}

	return out
}

func ParsePacketBytes(blob []byte) map[string]interface{} {
	// remove leading 'TOZ ' + size (first 8 bytes) if present
	if len(blob) > 8 {
		return parseStruct(blob[8:])
	}
	return map[string]interface{}{}
}

// // ParseMegaphoneData parses the megaphone messages into structured maps
// func ParseMegaphoneData(input string) map[string]string {
// 	// 切割每一筆 MegaphoneData
// 	entries := strings.Split(input, "MegaphoneData\x06\x04")
// 	if len(entries) != 2 {
// 		return map[string]string{} // 解析失敗
// 	}

// 	// 正則用來擷取鍵值
// 	re := regexp.MustCompile(`([A-Za-z]+)\x04([^\x06\x0b]+)`)

// 	e := entries[1]
// 	e = strings.TrimSpace(e)
// 	if e == "" {
// 		return map[string]string{}
// 	}
// 	data := make(map[string]string)
// 	data["Header"] = "MegaphoneData"

// 	matches := re.FindAllStringSubmatch(e, -1)
// 	for _, m := range matches {
// 		if len(m) == 3 {
// 			key := strings.TrimSpace(m[1])
// 			val := strings.TrimSpace(m[2])
// 			data[key] = val
// 		}
// 	}

// 	return data
// }

// ParseMegaphoneData 解析單筆 MegaphoneData 字串（可以包含控制字元），回傳 map
func ParseMegaphoneData(input string) map[string]string {
	data := map[string]string{
		"Header": "MegaphoneData",
	}

	// 欄位名稱（若未來想擴充只要調整這裡）
	keys := []string{"Type", "Nickname", "Text", "UserId", "ProfileCode"}

	// 找到 MegaphoneData 開頭（如果有的話）並從那裡開始處理
	startIdx := strings.Index(input, "MegaphoneData")
	var b []byte
	if startIdx >= 0 {
		b = []byte(input[startIdx+len("MegaphoneData"):])
	} else {
		b = []byte(input)
	}

	// 找出每個 key 在 b 中的所有位置（只取第一次出現）
	type posKv struct {
		pos int
		key string
	}
	var found []posKv
	for _, k := range keys {
		if idx := bytes.Index(b, []byte(k)); idx >= 0 {
			found = append(found, posKv{pos: idx, key: k})
		}
	}

	// 若沒有找到任何 key 就直接回傳空 map（帶 Header）
	if len(found) == 0 {
		return data
	}

	// 依在 stream 中的位置排序
	sort.Slice(found, func(i, j int) bool { return found[i].pos < found[j].pos })

	// 逐一取出 value（從 key 後面開始，直到下一個 key 開頭）
	for i, kv := range found {
		key := kv.key
		valStart := kv.pos + len(key)

		// 跳過緊接的控制字元（例如 \x04 等）
		for valStart < len(b) && b[valStart] < 0x20 {
			valStart++
		}

		valEnd := len(b)
		if i+1 < len(found) {
			valEnd = found[i+1].pos
		}

		// 如果下一 key 之前有多餘控制字元，可以向前 trim
		// 將 valEnd 向左移動直到遇到非控制字元（避免把控制符放到 value 末端）
		for valEnd > valStart && b[valEnd-1] < 0x20 {
			valEnd--
		}

		if valStart >= valEnd {
			data[key] = ""
			continue
		}

		val := string(b[valStart:valEnd])

		whisperIndex := strings.Index(val, "Whisper")
		if whisperIndex > 0 {
			val = val[:whisperIndex-1]
		}

		// 移除詭異參數
		reColor := regexp.MustCompile(`[A-Za-z[:punct:]][\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0A\x0B\x0C\x0D\x0E\x0F]{3}`)
		val = reColor.ReplaceAllString(val, "")

		// 移除色碼資訊
		reColor = regexp.MustCompile(`(?i)#[0-9A-F]{6}`)
		val = reColor.ReplaceAllString(val, "")

		// 通用 trim：去掉頭尾空白與常見控制字元
		val = strings.TrimSpace(val)
		val = strings.Trim(val, "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0A\x0B\x0C\x0D\x0E\x0F\xf4")

		// 清除控制字元與無效 rune
		cleaned := make([]rune, 0, len(val))
		for _, r := range val {
			// 保留常見可見字元與常用符號
			if r >= 0x20 && r != 0x7F && utf8.ValidRune(r) {
				cleaned = append(cleaned, r)
			}
		}
		val = string(cleaned)

		// 特例處理（保留你原本的邏輯）
		switch key {
		// case "Nickname":
		// case "Text":
		// 	if len(val) > 4 && val[1] == '\x00' && val[2] == '\x00' && val[3] == '\x00' {
		// 		val = val[4:]
		// 	}
		// 	val = strings.TrimSuffix(val, "\x06")
		// case "Type":
		// 	val = strings.TrimSuffix(val, "\b")
		case "ProfileCode":
			// 若存在，取前 5 個字元（但先檢查長度避免 panic）
			if len(val) > 5 {
				val = val[:5]
			}
		}

		data[key] = val
	}

	return data
}
