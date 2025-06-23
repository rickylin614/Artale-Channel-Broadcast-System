package devicex

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// 嘗試用 UDP 取得「對外使用的本地 IP」
// 不會真的送出封包，但會讓 OS 選擇用哪個本地位址/介面
func outboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80") // 可換成任意可到達的外網 IP
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}

// 在 pcap.FindAllDevs() 結果裡，根據本地 IP 找到對應的介面名稱
func findDeviceByIP(ip net.IP) (string, error) {
	devs, err := pcap.FindAllDevs()
	if err != nil {
		return "", err
	}
	for _, d := range devs {
		for _, a := range d.Addresses {
			if a.IP != nil && a.IP.Equal(ip) {
				return d.Name, nil
			}
		}
	}
	return "", fmt.Errorf("no device matched ip %s", ip.String())
}

// fallback: 短時間掃描每個介面，看哪個介面能接到含 TCP payload 或含特徵字串的封包
func probeDevicesForTOZ(timeoutPerDevice time.Duration) (string, error) {
	devs, err := pcap.FindAllDevs()
	if err != nil {
		return "", err
	}
	for _, d := range devs {
		// skip obvious virtual adapters that usually have no external traffic
		nameLower := strings.ToLower(d.Name + " " + d.Description)
		if strings.Contains(nameLower, "wan miniport") ||
			strings.Contains(nameLower, "bluetooth") ||
			strings.Contains(nameLower, "virtual") ||
			strings.Contains(nameLower, "tunnel") {
			continue
		}

		handle, err := pcap.OpenLive(d.Name, 1600, true, pcap.BlockForever)
		if err != nil {
			continue
		}
		// 只抓 tcp，減少雜訊
		_ = handle.SetBPFFilter("tcp")
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

		found := make(chan bool, 1)
		go func() {
			for pkt := range packetSource.Packets() {
				if tcpLayer := pkt.Layer(layers.LayerTypeTCP); tcpLayer != nil {
					// 如果 payload 有東西就視為有流量
					if app := pkt.ApplicationLayer(); app != nil && len(app.Payload()) > 0 {
						// 可檢查特徵: 例如是否包含 "TOZ "
						payload := app.Payload()
						if len(payload) >= 4 && (string(payload[:3]) == "TOZ" || true) {
							found <- true
							return
						}
					}
				}
			}
		}()

		select {
		case <-found:
			handle.Close()
			return d.Name, nil
		case <-time.After(timeoutPerDevice):
			handle.Close()
		}
	}
	return "", fmt.Errorf("no active device found by probing")
}

// 綜合策略：先嘗試 outboundIP -> findDeviceByIP，失敗就 probe
func AutoPickDevice() (string, error) {
	// 1) 嘗試用 UDP trick 拿到 outbound IP
	ip, err := outboundIP()
	if err == nil {
		if ip != nil && !ip.IsUnspecified() {
			if dev, err2 := findDeviceByIP(ip); err2 == nil {
				return dev, nil
			}
		}
	} else {
		log.Println("outboundIP err:", err)
	}

	// 2) 若找不到或失敗，短時間 probe 各介面
	dev, err := probeDevicesForTOZ(2 * time.Second) // 每個介面試 2 秒
	if err == nil {
		return dev, nil
	}

	// 3) 最後 fallback：回傳第一個看起來像實體網卡的介面
	devs, err := pcap.FindAllDevs()
	if err != nil || len(devs) == 0 {
		return "", fmt.Errorf("no devices available")
	}
	for _, d := range devs {
		nameLower := strings.ToLower(d.Name + " " + d.Description)
		if strings.Contains(nameLower, "intel") ||
			strings.Contains(nameLower, "realtek") ||
			strings.Contains(nameLower, "wifi") ||
			strings.Contains(nameLower, "ethernet") {
			return d.Name, nil
		}
	}
	return devs[0].Name, nil
}
