package main

import (
	"bytes"
	"context"
	"fmt"
	goenv "github.com/joho/godotenv"
	"github.com/searchlight/james-load-testing/inbox"
	openapi "go.opscenter.dev/james-go-client"
	"html/template"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"sync"
	"time"
	//openapi "github.com/searchlight/james-go-client"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
)

const (
	NoOfUsers       = 1000
	NoOfMailingList = 10000
)

var (
	PrometheusEndpoint          = ""
	ApacheJamesWebAdminEndpoint = ""
	ApacheJamesWebAdminPort     = ""
	ApacheJamesWebAdminToken    = ""
	ApacheJamesJMAPEndpoint     = ""

	RunLoadTestingForMinute    = 60
	ReqPerSecondForLoadTesting = 10
	NumberOfMemberPerGroup     = 20

	TestingDomain           = "load.testing"
	UserEmailPattern        = "user_%v@" + TestingDomain
	UserEmailCommonPassword = "H1sJk6ORaS"
	GroupPattern            = "group_%v@" + TestingDomain
	GroupMembers            = make([][]int, NoOfMailingList+1)
	EmailCountsOfUsers      = make([]int, NoOfUsers+1)
	mu                      = sync.Mutex{}
)

func GetApacheJamesApiClient() *openapi.APIClient {
	configuration := openapi.NewConfiguration().WithAccessToken(context.Background(), ApacheJamesWebAdminToken)
	mu.Lock()
	configuration.Servers[0] = openapi.ServerConfiguration{
		URL: fmt.Sprintf("%v:%v", ApacheJamesWebAdminEndpoint, ApacheJamesWebAdminPort),
	}
	mu.Unlock()
	apiClient := openapi.NewAPIClient(configuration)

	return apiClient
}

func main() {
	app := cli.NewApp()
	app.Name = "James load testing"
	app.Usage = "To load test the apache james server"
	app.Commands = []cli.Command{
		CmdLoadTesting,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("failed to run app with %s: %v", os.Args, err)
	}
	//james.SendMailAPIService{}
}

var CmdLoadTesting = cli.Command{
	Name:   "loadtest",
	Action: RunLoadTesting,
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "req_per_second",
			Value: 10,
		},
		cli.IntFlag{
			Name:  "run_for_minutes",
			Value: 60,
		},
		cli.IntFlag{
			Name:  "member_per_group",
			Value: 20,
		},
	},
}

func RunLoadTesting(ctx *cli.Context) {
	// Load environment variables from .env file
	loadEnv()
	log.Printf("=====================new================")
	// Get the cli options
	RunLoadTestingForMinute = ctx.Int("run_for_minutes")
	ReqPerSecondForLoadTesting = ctx.Int("req_per_second")
	NumberOfMemberPerGroup = ctx.Int("member_per_group")

	// Get options from env file
	ApacheJamesWebAdminEndpoint = os.Getenv("URL")
	ApacheJamesWebAdminPort = os.Getenv("HTTP_PORT")
	ApacheJamesWebAdminToken = os.Getenv("ADMIN_TOKEN")

	log.Printf("RunLoadTestingForMinute: %v", RunLoadTestingForMinute)
	log.Printf("ReqPerSecondForLoadTesting: %v", ReqPerSecondForLoadTesting)
	log.Printf("NumberOfMemberPerGroup: %v", NumberOfMemberPerGroup)

	log.Printf("ApacheJamesWebAdminEndpoint: %v", ApacheJamesWebAdminEndpoint)
	log.Printf("ApacheJamesWebAdminPort: %v", ApacheJamesWebAdminPort)
	log.Printf("ApacheJamesWebAdminToken: %v", ApacheJamesWebAdminToken)

	if err := testServerConnectivity(); err != nil {
		log.Fatalf("can't connect with the admin service, err: %v", err)
	}

	//initiate server
	initiate()

	// start load testing process
	startBulkProcess()

	// clean the server
	clean()
}

func loadEnv() {
	err := goenv.Load(".env")
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}
}

func testServerConnectivity() error {
	client := GetApacheJamesApiClient()
	sts, r, err := client.HealthcheckAPI.CheckAllComponents(context.TODO()).Execute()
	if err != nil {
		return err
	}

	if r.StatusCode != 200 || *sts.Status != "healthy" {
		return fmt.Errorf("health check failed: statusCode: %v", err)
	}

	return nil
}

func initiate() {
	/*
		- Create the domain
		- Create the users account
		- Create the groups/mailing list
		- Assign users among the groups randomly
	*/
	if err := createDomain(); err != nil {
		log.Fatal(err.Error())
	}

	if err := createUsers(); err != nil {
		log.Fatal(err.Error())
	}

	if err := assignUsersToGroups(); err != nil {
		log.Printf(err.Error())
	}
}

func clean() {
	/*
		- Delete the domain
		- Delete the users
		- Delete the groups
	*/
	_ = deleteUsers()
	_ = deleteDomain()
}

func deleteDomain() error {
	apiClient := GetApacheJamesApiClient()
	r, err := apiClient.DomainsAPI.DeleteDomain(context.Background(), TestingDomain).Execute()
	if err != nil {
		return err
	}
	if r.StatusCode >= 300 {
		return fmt.Errorf("status: %v, statusCode: %v", r.Status, r.StatusCode)
	}

	return nil
}

func deleteUsers() error {
	var (
		maxGoRoutineLimit = runtime.NumCPU() * 10
		eg                = errgroup.Group{}
	)
	// set the maximum go routine supported
	eg.SetLimit(maxGoRoutineLimit)

	for i := 1; i <= NoOfUsers; i++ {
		iCopy := i
		func(userNo int, commonEmailPattern string) {
			eg.Go(func() error {
				apiClient := GetApacheJamesApiClient()
				userEmail := fmt.Sprintf(commonEmailPattern, userNo)

				var (
					r   *http.Response
					err error
				)
				for try := 0; try <= 3; try++ {
					r, err = apiClient.UsersAPI.DeleteUser(context.Background(), userEmail).Execute()
					if r.StatusCode == 409 {
						err = nil
						break
					}
					if err != nil {
						time.Sleep(time.Millisecond * 10)
						continue
					}
				}

				if err != nil || r.StatusCode >= 300 {
					log.Printf("failed deletion of user: %v, status: %v, statusCode: %v", userEmail, r.Status, r.StatusCode)
				} else {
					log.Printf("successfully deleted user: %v", userEmail)
				}

				return nil
			})
		}(iCopy, UserEmailPattern)
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func createDomain() error {
	apiClient := GetApacheJamesApiClient()
	r, err := apiClient.DomainsAPI.CreateDomain(context.Background(), TestingDomain).Execute()
	if err != nil {
		return err
	}
	if r.StatusCode >= 300 {
		return fmt.Errorf("status: %v, statusCode: %v", r.Status, r.StatusCode)
	}

	return nil
}

func createUsers() error {
	var (
		maxGoRoutineLimit = runtime.NumCPU() * 10
		eg                = errgroup.Group{}
	)
	// set the maximum go routine supported
	eg.SetLimit(maxGoRoutineLimit)

	for i := 1; i <= NoOfUsers; i++ {
		iCopy := i
		func(userNo int, commonEmailPattern, commonPass string) {
			eg.Go(func() error {
				apiClient := GetApacheJamesApiClient()

				userEmail := fmt.Sprintf(commonEmailPattern, userNo)
				body := openapi.UpsertUserRequest{
					Password: commonPass,
				}

				var (
					r   *http.Response
					err error
				)
				for try := 0; try <= 3; try++ {
					r, err = apiClient.UsersAPI.UpsertUser(context.Background(), userEmail).UpsertUserRequest(body).Execute()
					if r.StatusCode == 409 {
						err = nil
						break
					}
					if err != nil {
						time.Sleep(time.Millisecond * 10)
						continue
					}
				}

				if err != nil {
					return err
				}

				if r.StatusCode >= 300 && r.StatusCode != 409 {
					return fmt.Errorf("failed creation of user: %v, status: %v, statusCode: %v", userEmail, r.Status, r.StatusCode)
				}

				log.Printf("successfully create user: %v", userEmail)
				return nil
			})
		}(iCopy, UserEmailPattern, UserEmailCommonPassword)
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func assignUsersToGroups() error {
	log.Printf("assinging users to groups started")
	var (
		maxGoRoutineLimit = runtime.NumCPU() * 20
		eg                = errgroup.Group{}
	)
	// set the maximum go routine supported
	eg.SetLimit(maxGoRoutineLimit)

	for groupNo := 1; groupNo <= NoOfMailingList; groupNo++ {

		groupNoCopy := groupNo
		func(groupNo int, groupPattern, userPattern string) {
			eg.Go(func() error {

				localErrGroup := errgroup.Group{}
				localErrGroup.SetLimit(10)
				localGroupNoCopy := groupNo

				for rn := 1; rn <= NumberOfMemberPerGroup; rn++ {
					func(groupNo int, groupPattern, userPattern string) {
						localErrGroup.Go(func() error {
							apiClient := GetApacheJamesApiClient()

							userNo := rand.Intn(NoOfUsers) + 1
							memberAddr := fmt.Sprintf(userPattern, userNo)
							groupAddr := fmt.Sprintf(GroupPattern, groupNo)

							var (
								r   *http.Response
								err error
							)
							for try := 0; try < 3; try++ {
								r, err = apiClient.AddressGroupAPI.AddMember(context.Background(), groupAddr).MemberAddress(memberAddr).Execute()
								if err != nil || r.StatusCode >= 300 {
									time.Sleep(time.Millisecond * 10)
									continue
								}
							}

							if err != nil {
								log.Printf("failed assinging in group, groupAddr: %v, memberAddr: %v, err: %v", err, groupPattern, memberAddr)
							} else if r.StatusCode >= 300 {
								log.Printf("failed assinging in group, groupAddr: %v, memberAddr: %v, status: %v, statusCode: %v", groupAddr, memberAddr, r.Status, r.StatusCode)
							} else {
								log.Printf("sucessfully aassigned in group: groupAddr: %v, memberAddr: %v", groupAddr, memberAddr)
							}

							// Add to group member
							GroupMembers[groupNo] = append(GroupMembers[groupNo], userNo)

							return nil
						})
					}(localGroupNoCopy, groupPattern, userPattern)
				}

				_ = localErrGroup.Wait()

				return nil
			})
		}(groupNoCopy, GroupPattern, UserEmailPattern)
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

type DiagnosticResult struct {
	CheckType string
	Timestamp metav1.Time
	Outputs   []DiagnosticOutput
}

type DiagnosticOutput struct {
	Description string
	Content     []byte
}

func GetSampleOutput() (string, error) {
	mp := make(map[string]string)
	mp["hello"] = "world"
	mp["hala"] = "madrid"

	var data []byte
	for i := 0; i < 1; i++ {
		data = append([]byte("Lorem Ipsum is simply dummy text of the printing and typesetting industry."), data...)
	}
	var data1 []byte
	for i := 0; i < 2; i++ {
		data1 = append([]byte("Lorem Ipsum is simply dummy text of the printing and typesetting industry."), data1...)
	}
	var data2 []byte
	for i := 0; i < 3; i++ {
		data2 = append([]byte("Lorem Ipsum is simply dummy text of the printing and typesetting industry."), data2...)
	}
	var data3 []byte
	for i := 0; i < 4; i++ {
		data3 = append([]byte("Lorem Ipsum is simply dummy text of the printing and typesetting industry."), data3...)
	}
	result := []DiagnosticResult{
		{
			CheckType: "InspectLogs",
			Timestamp: metav1.Now(),
			Outputs: []DiagnosticOutput{
				{Description: "Log 1", Content: data},
				{Description: "Log 2", Content: data1},
				{Description: "Log 3", Content: data2},
				{Description: "Log 4", Content: data3},
			},
		},
		{
			CheckType: "InspectConditions",
			Timestamp: metav1.Date(2010, 11, 1, 1, 1, 1, 1, time.Local),
			Outputs: []DiagnosticOutput{
				{Description: "Log 1", Content: data1},
				{Description: "Log 2", Content: data2},
				{Description: "Log 3", Content: data3},
				{Description: "Log 4", Content: data1},
			},
		},
	}

	out, err := GetOutput(mp, result)
	if err != nil {
		return "", err
	}
	return out, nil
}
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

func startBulkProcess() {
	var (
		maxGoRoutineLimit         = ReqPerSecondForLoadTesting
		reqInterval               = time.Second / time.Duration(ReqPerSecondForLoadTesting)
		eg                        = errgroup.Group{}
		osSignalChan              = make(chan os.Signal, 1)
		finishingTrigger          = time.NewTicker(time.Minute * time.Duration(RunLoadTestingForMinute))
		logPrintingTrigger        = time.NewTicker(time.Second * 10)
		numberOfSuccessfulReqSent = 0
		numberOfFailedReq         = 0
		mu                        = sync.Mutex{}
	)

	//client : = james.SendMailAPIService{}.SendEmail(context.Background()).Execute()

	// Catch the os signal
	signal.Notify(osSignalChan, os.Interrupt, os.Kill)

	// set the maximum go routine supported
	eg.SetLimit(maxGoRoutineLimit)

	log.Printf("Started load testing at time: %v", time.Now())

	testClient, err := inbox.GetTestJMAPClient()
	if err != nil {
		fmt.Errorf("could not create jmap client: ", err)
		return
	}
	output, err := GetSampleOutput()
	if err != nil {
		fmt.Errorf("could not get outputs: ", err)
		return
	}

	for {
		select {
		case <-osSignalChan:
			log.Printf("Process cancled by os signal")
			return
		case <-finishingTrigger.C:
			log.Printf("Time has ended, time: %v", time.Now())
			mu.Lock()
			log.Printf("Stats: successful req: %v, failed req: %v", numberOfSuccessfulReqSent, numberOfFailedReq)
			mu.Unlock()
			return
		case <-logPrintingTrigger.C:
			mu.Lock()
			log.Printf("Stats: successful req: %v, failed req: %v", numberOfSuccessfulReqSent, numberOfFailedReq)
			mu.Unlock()
		default:
			func(groupAddrPattern string) {
				eg.Go(func() error {
					randomGroupNo := rand.Intn(NoOfMailingList) + 1

					groupEmailAddr := fmt.Sprintf(groupAddrPattern, randomGroupNo)

					err = inbox.SendMail(testClient, groupEmailAddr, output)

					if err != nil {
						mu.Lock()
						numberOfFailedReq++
						mu.Unlock()
						log.Printf("failed to send mail, err: %v", err)
					} else {
						mu.Lock()
						numberOfSuccessfulReqSent++
						mu.Unlock()
						log.Printf("successflly send email, from: TEST CLIENT, to group; %v", groupEmailAddr)
					}

					return nil
				})
			}(GroupPattern)

			time.Sleep(reqInterval)
		}
	}
}
