package txlive

import (
	"crypto/md5"
	"fmt"
)

func BuildRTMPUrl(domain, streamName, key string, time string) (addrstr string) {
	var ext_str string
	if key != "" && time != "" {
		txSecret := md5.Sum([]byte(key + streamName + time))
		txSecretStr := fmt.Sprintf("%x", txSecret)
		ext_str = "?txSecret=" + txSecretStr + "&txTime=" + time
	}
	addrstr = "rtmp://" + domain + "/live/" + streamName + ext_str
	return
}
