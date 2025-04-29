package inbox

import (
	"bytes"
	"fmt"
	"git.sr.ht/~rockorager/go-jmap"
	_ "git.sr.ht/~rockorager/go-jmap/mail"
	james_go "go.opscenter.dev/james-go-client/inbox"
	"html/template"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"sort"
	"sync"
	"time"
)

const (
	testServerHostname     = "192.168.0.246"
	testServerJMAPPort     = "80"
	testServerWebAdminPort = 8000
	testuserDomain         = "cloud.appscode.com"
	testuserSender         = "testuser.acc@cloud.appscode.com"
	//testuserRecipient      = "recipient.acc@cloud.appscode.com"
	testuserPassword    = "password"
	testToken           = ""
	jmapSessionEndPoint = "http://" + testServerHostname + "/jmap/session"

	testRunbookPath = "../../hack/samples/mongodb-down-runbook.yaml"
	testAlertPath   = "../../hack/samples/webhookalert_mongodb_sample.yaml"
)

type DiagnosticResult struct {
	CheckType string
	Timestamp metav1.Time
	Outputs   []DiagnosticOutput
}

type DiagnosticOutput struct {
	Description string
	Content     []byte
}

type JMAPClient struct {
	jmap.Client
	//tokenGetter                 TokenGetterFunc
	mu                          sync.RWMutex
	userId                      jmap.ID
	userEmail                   string
	mailboxIds                  map[string]jmap.ID
	lastComputedCacheAtUnixTime int64
	cachedRenewalErr            error
}

// GetTestJMAPClient defaults to JWT auth unless ForceBasicAuth is set to true
func GetTestJMAPClient(
	jmapSessionEndpoint,
	userEmail,
	userPassword string,
	ForceBasicAuth bool,
) (*james_go.JMAPClient, error) {
	return james_go.NewJMAPClient(&james_go.JMAPConf{
		JMAPSessionEndpoint: jmapSessionEndpoint,
		ForceBasicAuth:      ForceBasicAuth,
		BasicAuthCreds: james_go.BasicAuthCredentials{
			Username: userEmail,
			Password: userPassword,
		},
		JMAPServerAddr: testServerHostname,
		JMAPServerPort: testServerJMAPPort,
		TokenGetter: func() (*http.Client, string, error) {
			return &http.Client{
				Transport: &http.Transport{},
			}, "", nil
		},
	})
}

/*
type JMAPConf struct {
	JMAPServerAddr      string
	JMAPServerPort      string
	JMAPSessionEndpoint string
	ForceBasicAuth      bool
	BasicAuthCreds      BasicAuthCredentials
	TokenGetter         TokenGetterFunc
}
*/

//func SendMail(testUserRecipient string) error {
//	testClient, err := GetTestJMAPClient(jmapSessionEndPoint, testuserSender, testuserPassword, true)
//	if err != nil {
//		//t.Error("could not create jmap client: ", err)
//		return err
//	}
//	mp := make(map[string]string)
//	mp["hello"] = "world"
//	mp["hala"] = "madrid"
//
//	var data []byte
//	for i := 0; i < 1; i++ {
//		data = append([]byte("Lorem Ipsum is simply dummy text of the printing and typesetting industry."), data...)
//	}
//	var data1 []byte
//	for i := 0; i < 2; i++ {
//		data1 = append([]byte("Lorem Ipsum is simply dummy text of the printing and typesetting industry."), data1...)
//	}
//	var data2 []byte
//	for i := 0; i < 3; i++ {
//		data2 = append([]byte("Lorem Ipsum is simply dummy text of the printing and typesetting industry."), data2...)
//	}
//	var data3 []byte
//	for i := 0; i < 4; i++ {
//		data3 = append([]byte("Lorem Ipsum is simply dummy text of the printing and typesetting industry."), data3...)
//	}
//	result := []DiagnosticResult{
//		{
//			CheckType: "InspectLogs",
//			Timestamp: metav1.Now(),
//			Outputs: []DiagnosticOutput{
//				{Description: "Log 1", Content: data},
//				{Description: "Log 2", Content: data1},
//				{Description: "Log 3", Content: data2},
//				{Description: "Log 4", Content: data3},
//			},
//		},
//		{
//			CheckType: "InspectConditions",
//			Timestamp: metav1.Date(2010, 11, 1, 1, 1, 1, 1, time.Local),
//			Outputs: []DiagnosticOutput{
//				{Description: "Log 1", Content: data1},
//				{Description: "Log 2", Content: data2},
//				{Description: "Log 3", Content: data3},
//				{Description: "Log 4", Content: data1},
//			},
//		},
//	}
//
//	out, err := GetOutput(mp, result)
//	if err != nil {
//		return err
//	}
//	emailOptions := []james_go.Option{
//		james_go.WithSubject("Test"),
//		james_go.WithHTMLBody(out),
//		james_go.WithRecipients([]string{testUserRecipient}),
//	}
//	myMail, _ := testClient.NewEmail(emailOptions...)
//	err = testClient.SendEmail(myMail)
//	if err != nil {
//		fmt.Println(err)
//		return err
//	}
//	return nil
//}

func GetOutput(vars map[string]string, results []DiagnosticResult) (string, error) {
	sortedKeys := make([]string, 0, len(vars))
	for key := range vars {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)

	const htmlTemplate = `
<!DOCTYPE html>
<html>
	<head>
		<style>
			table {
				width: 100%;
				border-collapse: collapse;
				table-layout: fixed;
				overflow: auto;
			}

			th,td {
				text-align: center;
				vertical-align: top;
				border: 1px solid black;
				padding: 4px;
				overflow: auto;
				white-space: pre-wrap;
				word-wrap: break-word;
				overflow-wrap: break-word;
			}

			pre {
				text-align: left;
				vertical-align: top;
				word-wrap: break-word;
				overflow-wrap: break-word;
				overflow: auto;
			}

		</style>

		<title>Diagnostic Results</title>
	</head>
	<body>
		<h3>Diagnostic Results</h3>
		<div>
			<ul>
				{{range $key, $value := .ArgKeys}}
					<li><strong>{{$key}}: {{$value}}</strong></li>
				{{end}}
			</ul>
		</div>
		<div>
			<table>
				<colgroup>
					<col style="width: 9.5%;">
					<col style="width: 7.5%;">
					<col style="width: 25%;">
					<col style="width: 53%;">
				</colgroup>
				<tr>
					<th>Check Type</th>
					<th>Timestamp</th>
					<th>Description</th>
					<th>Content</th>
				</tr>
				{{range $result := .DiagnosticResults}}
					{{range $output := $result.Outputs}}
						<tr>
							<td>{{$result.CheckType}}</td>
							<td>{{$result.Timestamp | formatTime}}</td>
							<td>{{$output.Description}}</td>
							<td><pre>{{$output.Content | html}}</pre></td>
						</tr>
					{{end}}
				{{end}}
			</table>
		</div>
	</body>

</html>
`
	tmpl, err := template.New("html").Funcs(template.FuncMap{
		"html": func(b []byte) string {
			return string(b) // In real scenarios, you might want to escape HTML, this is simplified.
		},
		"formatTime": func(t metav1.Time) string {
			return t.Time.Format(time.DateTime)
		},
	}).Parse(htmlTemplate)
	if err != nil {
		return "", fmt.Errorf("error creating template: %w", err)
	}

	var htmlBuf bytes.Buffer
	err = tmpl.Execute(&htmlBuf, struct {
		DiagnosticResults []DiagnosticResult
		ArgKeys           map[string]string
	}{
		DiagnosticResults: results,
		ArgKeys:           vars,
	})
	if err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}
	return htmlBuf.String(), nil
}
