package main // import "github.com/JassoftLtd/DashVid-Uploader"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/urfave/cli" // imports as package "cli"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const baseUrl = "https://api.dashvid.io/v1/"

type LoginResponse struct {
	Login      bool
	IdentityID string
	Token      string
}

type CreateVideoResponse struct {
	Id  string
	Url string
}

type MyProvider struct {
	cameraKey string
	expirationTime time.Time
}

func (m *MyProvider) Retrieve() (credentials.Value, error) {

	url := baseUrl + "auth/login/cameraKey"

	var jsonString string = `{"cameraKey":"` + m.cameraKey + `"}`
	var jsonB = []byte(jsonString)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonB))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	loginResponse := &LoginResponse{}

	err = json.Unmarshal(body, loginResponse)

	if err != nil {
		panic(err)
	}

	sess := session.New(&aws.Config{Region: aws.String("eu-west-1")})
	svc := cognitoidentity.New(sess)

	credential, err := svc.GetCredentialsForIdentity(&cognitoidentity.GetCredentialsForIdentityInput{
		IdentityId: &loginResponse.IdentityID,
		Logins: map[string]*string{
			"cognito-identity.amazonaws.com": &loginResponse.Token,
		},
	})

	if err != nil {
		panic(err)
	}

	m.expirationTime = *credential.Credentials.Expiration

	return credentials.Value{
		AccessKeyID:     *credential.Credentials.AccessKeyId,
		SecretAccessKey: *credential.Credentials.SecretKey,
		SessionToken:    *credential.Credentials.SessionToken,
		ProviderName:    *credential.IdentityId,
	}, nil
}
func (m *MyProvider) IsExpired() bool {
	if m.expirationTime.Before(time.Now()) {
		return true
	}

	return false
}

func uploadFile(directory string, file os.FileInfo, url string) bool {
	fileToUpload, err := os.Open(directory + file.Name())
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("PUT", url, fileToUpload)
	req.ContentLength = file.Size()
	req.Header.Add("Content-Type", "text/plain;charset=UTF-8")

	if err != nil {
		fmt.Println("error creating request", url)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("failed uploading file:", file.Name(), "Error:", err)
	}

	return resp.StatusCode == 200
}

func main() {
	app := cli.NewApp()
	app.Name = "dashvid"
	app.Usage = "DashVid.io Client"

	app.Commands = []cli.Command{
		{
			Name:  "upload",
			Usage: "upload video files using CameraKey. eg. dashvid upload [CameraKey] [VideoFolder]",
			Action: func(c *cli.Context) error {
				cameraKey := c.Args().Get(0)
				directory := c.Args().Get(1)
				fmt.Println("Uploading videos for CameraKey", cameraKey)
				fmt.Println("Uploading videos in folder", directory)
				deleteFiles := c.String("delete") == "yes"

				if deleteFiles {
					fmt.Println("Deleting files after uplaod")
				}

				creds := credentials.NewCredentials(&MyProvider{cameraKey: cameraKey})

				files, _ := ioutil.ReadDir(directory)
				for _, f := range files {

					if f.IsDir() {
						continue
					}

					fmt.Println("Requesting Upload for:", f.Name(), "Size:", f.Size()/1024/1024, "Mb")

					createVideoUrl := baseUrl + "video"

					var jsonString string = `{"fileName":"` + f.Name() + `", "fileType":"` + f.Name() + `", "cameraKey":"` + cameraKey + `"}`
					var jsonB = []byte(jsonString)

					req, err := http.NewRequest("POST", createVideoUrl, bytes.NewBuffer(jsonB))
					req.Header.Set("Content-Type", "application/json")

					signer := v4.NewSigner(creds)

					signer.Sign(req, bytes.NewReader(jsonB), "execute-api", "eu-west-1", time.Now())

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						panic(err)
					}
					defer resp.Body.Close()

					body, _ := ioutil.ReadAll(resp.Body)

					createVideoResponse := &CreateVideoResponse{}

					err = json.Unmarshal(body, createVideoResponse)
					if err != nil {
						panic(err)
					}

					uploaded := uploadFile(directory, f, createVideoResponse.Url)

					if uploaded {
						fmt.Println("Upload Successful for: ", f.Name())

						if deleteFiles {
							os.Remove(directory + f.Name())
							fmt.Println("Deleted: ", f.Name())
						}
					}
				}

				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "delete",
					Value: "yes",
					Usage: "delete files after upload",
				},
			},
		},
	}

	app.Run(os.Args)
}
