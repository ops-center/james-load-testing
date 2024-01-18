package main

import (
	"context"
	"fmt"
	goenv "github.com/joho/godotenv"
	openapi "github.com/searchlight/james-go-client"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"time"
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
	configuration := openapi.NewConfiguration().WithAccessToken(ApacheJamesWebAdminToken)
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

	// initiate server
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

var sampleMIME = `From: %v
To: %v
MIME-Version: 1.0
Content-Type: multipart/mixed;
        boundary="XXXXboundary text"

This is a multipart message in MIME format.

--XXXXboundary text
Content-Type: text/plain

Sample Body

--XXXXboundary text
Content-Type: text/plain;
Content-Disposition: attachment;
        filename="test.txt"

this is the attachment text

--XXXXboundary text--`

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
	// Catch the os signal
	signal.Notify(osSignalChan, os.Interrupt, os.Kill)

	// set the maximum go routine supported
	eg.SetLimit(maxGoRoutineLimit)

	log.Printf("Started load testing at time: %v", time.Now())
	for {
		select {
		case <-osSignalChan:
			log.Printf("Process cancled by os signal")
			return
		case <-finishingTrigger.C:
			log.Printf("Time has ended, time: %v", time.Now())
			return
		case <-logPrintingTrigger.C:
			mu.Lock()
			log.Printf("Stats: No of successfull req: %v, No of failed req: %v", numberOfSuccessfulReqSent, numberOfFailedReq)
			mu.Unlock()
		default:
			func(userAddrPattern, groupAddrPattern, mimeBody string) {
				eg.Go(func() error {
					randomUserNo := rand.Intn(NoOfUsers) + 1
					randomGroupNo := rand.Intn(NoOfMailingList) + 1

					userEmailAddr := fmt.Sprintf(userAddrPattern, randomUserNo)
					groupEmailAddr := fmt.Sprintf(groupAddrPattern, randomGroupNo)

					emailMimeBody := fmt.Sprintf(mimeBody, userEmailAddr, groupEmailAddr)
					apiClient := GetApacheJamesApiClient()
					r, err := apiClient.SendMailAPI.SendEmail(context.Background()).Body(emailMimeBody).Execute()
					if err != nil {
						mu.Lock()
						numberOfFailedReq++
						mu.Unlock()
						log.Printf("failed to send mail, err: %v", err)
					} else if r.StatusCode >= 300 {
						mu.Lock()
						numberOfFailedReq++
						mu.Unlock()
						log.Printf("failed to send mail, err: %v", err)
					} else {
						mu.Lock()
						numberOfSuccessfulReqSent++
						mu.Unlock()
						log.Printf("successflly send email, from: %v, to group; %v", userEmailAddr, groupEmailAddr)
					}

					return nil
				})
			}(UserEmailPattern, GroupPattern, sampleMIME)

			time.Sleep(reqInterval)
		}
	}
}

// IsEmptyString checks if the provided string is empty
func IsEmptyString(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}
