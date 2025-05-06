package inbox

import (
	"fmt"
	"testing"
)

func TestSendMail(t *testing.T) {

	testClient, err := GetTestJMAPClient()
	if err != nil {
		//fmt.Errorf("could not create jmap client: ", err)
		fmt.Printf("Could not create jmap client")
		return
	}

	err = SendMail(testClient, "usr&admin$1@cloud.appscode.com", "output")
	if err != nil {
		t.Error(err)
	}
}
