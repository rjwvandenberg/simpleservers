package secrets

import (
	"encoding/hex"
	"testing"
)

const contents = "{\r\t\"action\": \"opened\",\r\t\"issue\": {\r\t\t\"url\": \"https://api.github.com/repos/octocat/Hello-World/issues/1347\",\r\t\t\"number\": 1347,\r\t},\r\t\"repository\" : {\r\t\t\"id\": 1296269,\r\t\t\"full_name\": \"octocat/Hello-World\",\r\t\t\"owner\": {\r\t\t\"login\": \"octocat\",\r\t\t\"id\": 1,\r\t\t},\r\t},\r\t\"sender\": {\r\t\t\"login\": \"octocat\",\r\t\t\"id\": 1,\r\t}\r}"

func TestHMACSha256(t *testing.T) {
	testparams := []struct {
		secret string
		msg    string
		hash   string
	}{
		{"A", "{", "sha256=15f032205debfab34832b1bfd117fe5316ed2926cadce750cb23094669b9d045"},
		{"A", contents, "sha256=33f00bec779fbd3297481970be3cbd94ea97c41bbcbb2eb3da071c6e43330a38"},
		{"It's a Secret to Everybody", "Hello, World!", "sha256=757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17"},
	}
	for _, v := range testparams {
		secretbytes := []byte(v.secret)
		hashstring := v.hash[7:]
		sh := New("", secretbytes)
		hash, _ := hex.DecodeString(hashstring)
		if !sh.Validate(hash, []byte(v.msg)) {
			t.Fail()
		}
	}
}
