package parse_test

import (
	"artale-broadcast/parse"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMegaphoneData(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]string
	}{
		{
			name: "t1",
			input: `MegaphoneDataType 5120011Nickname紙槍神人 Whisper Textm72收耳敏60/110雪乾淨火尖槍上衣力量30趴劍士8/9力上衣7力褲裙槍攻10/3雪5力以上手套UserId20372100006036330ProfileCodeM0JXRH #5F0738 #edb0ce`, want: map[string]string{
				"Header":      "MegaphoneData",
				"Type":        "5120011",
				"Nickname":    "紙槍神人",
				"Text":        "m72收耳敏60/110雪乾淨火尖槍上衣力量30趴劍士8/9力上衣7力褲裙槍攻10/3雪5力以上手套",
				"UserId":      "20372100006036330",
				"ProfileCode": "M0JXR",
			},
		},
		{
			name: "t2",
			input: `MegaphoneDataType 5120011Nickname澄澄把拔 Whisper TextN收乾淨香菇112-115ap 120.150.200.250雪/收58-78等法師套11屬+15雪起UserId20372100005723305ProfileCode4kNJRr #5F0738 #edb0ce`, want: map[string]string{
				"Header":      "MegaphoneData",
				"Type":        "5120011",
				"Nickname":    "澄澄把拔",
				"Text":        "N收乾淨香菇112-115ap 120.150.200.250雪/收58-78等法師套11屬+15雪起",
				"UserId":      "20372100005723305",
				"ProfileCode": "4kNJR",
			},
		},
		{
			name:  "t3",
			input: "MegaphoneData\x06\x04Type\x04 5120011\x08Nickname\x04\x0c澄澄把拔 Whisper \x01\x04Text\x04N收乾淨香菇112-115ap 120.150.200.250雪/收58-78等法師套11屬+15雪起\x06UserId\x04\x1120372100005723305\x0bProfileCode\x04\x054kNJR\x02r\x04 #5F0738\x04 #edb0ce", want: map[string]string{
				"Header":      "MegaphoneData",
				"Type":        "5120011",
				"Nickname":    "澄澄把拔",
				"Text":        "N收乾淨香菇112-115ap 120.150.200.250雪/收58-78等法師套11屬+15雪起",
				"UserId":      "20372100005723305",
				"ProfileCode": "4kNJR",
			},
		},
		{
			name:  "t4",
			input: `MegaphoneDataTextM●收弓手套5-8/1/5/15/25●手攻10/7.5/收眼力100%私/克洛斧97/5/10UserId20372100006213311Nickname帥展EXProfileCode4xBuFType5120010WhisperI#c597d4`, want: map[string]string{
				"Header":      "MegaphoneData",
				"Type":        "5120010",
				"Nickname":    "帥展EX",
				"Text":        "M●收弓手套5-8/1/5/15/25●手攻10/7.5/收眼力100%私/克洛斧97/5/10",
				"UserId":      "20372100006213311",
				"ProfileCode": "4xBuF",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parse.ParseMegaphoneData(tt.input)
			assert.Equal(t, tt.want, got, "ParseMegaphoneData() output mismatch")
		})
	}
}
