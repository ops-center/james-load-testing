package inbox

import "testing"

func TestSendMail(t *testing.T) {
	err := SendMail(testuserSender)
	if err != nil {
		t.Error(err)
	}
}
