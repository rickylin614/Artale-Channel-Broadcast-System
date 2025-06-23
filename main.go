package main

import (
	"artale-broadcast/config"
	"artale-broadcast/devicex"
	"artale-broadcast/discordsender"
	"artale-broadcast/parse"
	"fmt"
	"log"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

const (
	PORT           = 32800
	BPF_FILTER     = "tcp port 32800"
	WEBSOCKET_PORT = 8765
)

var analyzeStrSli = make([]string, 0)

func startSniffer(sender *discordsender.Sender) {
	device, err := devicex.AutoPickDevice()
	if err != nil {
		log.Fatal("pick device:", err)
	}
	log.Println("Using device:", device)

	// snapshot len, promiscuous, timeout
	handle, err := pcap.OpenLive(device, 65536, true, pcap.BlockForever)
	if err != nil {
		log.Fatal("pcap openlive:", err)
	}
	if err := handle.SetBPFFilter(BPF_FILTER); err != nil {
		log.Fatalf("BPF filter error (%s): %v", BPF_FILTER, err)
	}
	log.Println("Started capture with filter:", BPF_FILTER)
	source := gopacket.NewPacketSource(handle, handle.LinkType())

	for dataPacket := range source.Packets() {
		// get TCP payload
		if tcpLayer := dataPacket.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			tcp, _ := tcpLayer.(*layers.TCP)
			payload := tcp.Payload
			if len(payload) == 0 {
				continue
			}
			payloadStr := string(payload)
			index := strings.Index(payloadStr, "MegaphoneData")
			if index < 0 {
				continue
			}
			// TODO check the prefix data mean?
			// analyzeStrSli = append(analyzeStrSli, payloadStr[:index])
			// if len(analyzeStrSli)%10 == 0 {
			// 	packet.AnalyzePackets(analyzeStrSli)
			// }
			msgInput := payloadStr[index:]
			data := parse.ParseMegaphoneData(msgInput)
			fmt.Println(data)
			sender.InputChan <- discordsender.Message{
				Nickname:    data["Nickname"],
				ProfileCode: data["ProfileCode"],
				Text:        data["Text"],
			}
		}
	}
	analyzeStrSli = make([]string, 0)
}

// ---------- main ----------
func main() {
	conf, err := config.LoadConfig("conf.toml")
	if err != nil {
		log.Fatalf("載入設定失敗: %v", err)
	}
	log.Println("讀取設定檔案 web_hook:", conf.WebHook)
	log.Println("服務啟動")

	sender := discordsender.NewSender(conf.WebHook)

	// start sniffer
	startSniffer(sender)
}
