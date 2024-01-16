package main

import (
	"context"
	"fmt"
	openapi "github.com/searchlight/james-go-client"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"
)

const (
	NoOfUsers       = 1000
	NoOfMailingList = 10000
)

var (
	PrometheusEndpoint          = ""
	ApacheJamesWebAdminEndpoint = "http://james.appscode.ninja"
	ApacheJamesWebAdminPort     = "8000"
	ApacheJamesWebAdminUsername = ""
	ApacheJamesWebAdminPassword = ""
	ApacheJamesJMAPEndpoint     = ""
	RunLoadTestingForMinute     = 60
	ReqPerSecondForLoadTesting  = 10

	TestingDomain           = "load.testing"
	UserEmailPattern        = "user_%v@" + TestingDomain
	UserEmailCommonPassword = "H1sJk6ORaS"
	GroupPattern            = "group_%v@" + TestingDomain
	GroupMembers            = make([][]int, NoOfMailingList+1)
	EmailCountsOfUsers      = make([]int, NoOfUsers+1)
	NumberOfMemberPerGroup  = 20
	mu                      = sync.Mutex{}
)

func GetApacheJamesApiClient() *openapi.APIClient {
	configuration := openapi.NewConfiguration()
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
	Name:   "testing",
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
	},
}

func RunLoadTesting(ctx *cli.Context) {
	// Get the cli options
	RunLoadTestingForMinute = ctx.Int("run_for_minutes")
	ReqPerSecondForLoadTesting = ctx.Int("req_per_second")

	initiate()

	eg := errgroup.Group{}
	eg.SetLimit(1)

	eg.Go(func() error {
		// start the load testing
		startBulkProcess()
		return nil
	})
	_ = eg.Wait()

	log.Printf("No of email sending reports")
	clean()
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
		maxGoRoutineLimit = ReqPerSecondForLoadTesting
		reqInterval       = time.Second / time.Duration(ReqPerSecondForLoadTesting)
		eg                = errgroup.Group{}
		osSignalChan      = make(chan os.Signal, 1)
	)
	// Catch the os signal
	signal.Notify(osSignalChan, os.Interrupt, os.Kill)

	// set the maximum go routine supported
	eg.SetLimit(maxGoRoutineLimit)

	log.Printf("Started load testing at time: %v", time.Now())
	finishingTrigger := time.NewTicker(time.Minute * time.Duration(RunLoadTestingForMinute))
	for {
		select {
		case <-osSignalChan:
			log.Printf("Process cancled by os signal")
			return
		case <-finishingTrigger.C:
			log.Printf("Time has ended, time: %v", time.Now())
			return
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
						log.Printf("failed to send mail, err: %v", err)
					} else if r.StatusCode >= 300 {
						log.Printf("failed to send mail, err: %v", err)
					} else {
						log.Printf("successflly send email, from: %v, to group; %v", userEmailAddr, groupEmailAddr)
					}

					return nil
				})
			}(UserEmailPattern, GroupPattern, sampleMIME)

			time.Sleep(reqInterval)
		}
	}
}
