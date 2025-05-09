package inbox

import (
	"bytes"
	"crypto/tls"
	"fmt"
	james_go "go.opscenter.dev/james-go-client/inbox"
	"net/http"
)

const (
	testServerHostname     = "10.2.0.214"
	testServerJMAPPort     = "80"
	testServerWebAdminPort = 8000
	testuserDomain         = "cloud.appscode.com"
	testuserSender         = "usr&admin$2@cloud.appscode.com"
	testuserRecipient      = "usr&admin$1@cloud.appscode.com"
	testuserPassword       = "password"
	testToken              = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDI5NTUxMzQ2LCJwZXJtaXNzaW9ucyI6eyJwZXJtLmFkZHJlc3MuZ3JvdXBzLioiOlsiR0VUIiwiUE9TVCIsIkRFTEVURSJdLCJwZXJtLmFkZHJlc3MuZ3JvdXBzLiouKiI6WyJQVVQiLCJERUxFVEUiXX0sInN1YiI6InVzciZhZG1pbiQyQGNsb3VkLmFwcHNjb2RlLmNvbSIsInR5cGUiOiJhZG1pbiJ9.LEHdZ07qjsLaPEAbfdpRY1aq6KDQCJ9z6PkY6QxpcWprAKs4jQu365uqHuMYVxmdauVpoFGaAhs8a_WlGbq8voEzRevxdq6imGQZp0I_F8hnJyBigxDEeDJpEUeCbGUjPWK_FVxzlQtnjzxoqaVPt9NA73szStdEXzNeeX7KjVKWPbLh5WN0hUoHmxazYiBPitsu0IUEbcSQs-sN5Yl_hQjhYhCFi7qmcDSv8vwlKoydzMi93PFLCvuH5pJfpAALzzs2bwcHzPDjM-KhfdhV7AvU87dJBCzh8VzivCoNWXjA5P0EpGN7bkpvLmz3kifSEuUPU_SCX5fTNrPUpVqOag"
	jmapSessionEndPoint    = "http://" + testServerHostname + ":" + testServerJMAPPort + "/jmap/session"
	forceBasicAuth         = true
)

func NewHttpClient() *http.Client {

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // temporary workaround for testing cross-cluster functionalities
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{Transport: transport}
}

/*
jmapSessionEndPoint
forceBasicAuth
testuserSender
testuserPassword
testServerHostname
testServerJMAPPort
*/

// defaults to JWT auth unless ForceBasicAuth is set to true
func GetTestJMAPClient() (*james_go.JMAPClient, error) {
	return james_go.NewJMAPClient(&james_go.JMAPConf{
		JMAPSessionEndpoint: jmapSessionEndPoint,
		ForceBasicAuth:      forceBasicAuth,
		BasicAuthCreds: james_go.BasicAuthCredentials{
			Username: testuserSender,
			Password: testuserPassword,
		},
		JMAPServerAddr: testServerHostname,
		JMAPServerPort: testServerJMAPPort,
		TokenGetter: func() (*http.Client, string, error) {
			return NewHttpClient(), testToken, nil
		},
	})
}

func SendMail(testClient *james_go.JMAPClient, receiverGroup string, output string) error {
	emailOptions := []james_go.Option{
		james_go.WithSubject("Test NOW"),
		james_go.WithHTMLBody(output),
		james_go.WithAttachment("one.txt", bytes.NewReader([]byte(output))),
		james_go.WithAttachment("two.txt", bytes.NewReader([]byte(output))),
		james_go.WithAttachment("three.txt", bytes.NewReader([]byte(output))),
		james_go.WithRecipients([]string{receiverGroup}),
	}
	myMail, _ := testClient.NewEmail(emailOptions...)
	err := testClient.SendEmail(myMail)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
