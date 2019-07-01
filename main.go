package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var clientID string
var clientSecret string
var serviceAccount string

var httpClient = &http.Client{Timeout: 200 * time.Second}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func moraviaBaseURL() string {
	useProd := getenv("moravia_production", "false")
	if useProd == "true" {
		return "https://projects.moravia.com/Api/V4"
	} else {
		return "https://test-projects.moravia.com/Api/V4"
	}
}

func moraviaLoginURL() string {
	useProd := getenv("moravia_production", "false")
	if useProd == "true" {
		return "https://login.moravia.com/connect/token"
	} else {
		return "https://test-login.moravia.com/connect/token"
	}
}

func moraviaJobsURL() string {
	return moraviaBaseURL() + "/Jobs"
}

func moraviaJobAttachmentsURL() string {
	return moraviaBaseURL() + "/jobattachments"
}

func moraviaProjectsURL() string {
	return moraviaBaseURL() + "/Projects"
}

type MoraviaProjectConfiguration struct {
	Id int `yaml:"id"`
}

type MoraviaJobTemplateConfiguration struct {
	Name             string   `yaml:"name"`
	Source           string   `yaml:"source"`
	Source_language  string   `yaml:"source_language"`
	Target_languages []string `yaml:"target_languages"`
}

type MoraviaConfiguration struct {
	Project      MoraviaProjectConfiguration     `yaml:"project"`
	Job_template MoraviaJobTemplateConfiguration `yaml:"job_template"`
}

func (config *MoraviaConfiguration) readFromFile(filepath string) *MoraviaConfiguration {
	yamlFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return config
}

type AuthenticateResponse struct {
	Access_token string `json:"access_token"`
	Expires_in   int    `json:"expires_in"`
	Token_type   string `json:"token_type"`
}

func authenticate(clientID string, clientSecret string, serviceAccount string, target interface{}) error {
	var bodyString = "grant_type=service"
	bodyString += "&client_id=" + clientID
	bodyString += "&client_secret=" + clientSecret
	bodyString += "&scope=symfonie2-api&service_account=" + serviceAccount

	body := strings.NewReader(bodyString)
	req, err := http.NewRequest("POST", moraviaLoginURL(), body)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// responseData, sErr := ioutil.ReadAll(resp.Body)
	// if sErr != nil {
	//
	// }
	// fmt.Println(string(responseData))

	return json.NewDecoder(resp.Body).Decode(target)
}

type Job struct {
	Id                  int
	Name                string `yaml:"name"`
	ProjectId           int
	Description         string   `yaml:"description"`
	SourceLanguageCode  string   `yaml:"source_language"`
	TargetLanguageCodes []string `yaml:"target_languages"`
}

type Jobs struct {
	Value []Job `json:"value"`
}

func moraviaPortalJobDetailsURL(job Job) string {
	useProd := getenv("moravia_production", "false")
	if useProd == "true" {
		return "https://projects.moravia.com/jobs/" + string(job.Id) + "/detail"
	} else {
		return "https://test-projects.moravia.com/jobs/" + string(job.Id) + "/detail"
	}
}

func jobFromTemplate(filepath string, target interface{}) {
	// TODO: Alex - fill in template
}

func findJob(name string, auth AuthenticateResponse, target interface{}) error {
	var bodyString = ""

	body := strings.NewReader(bodyString)

	// https://projects.moravia.com/api/V3/Jobs?$filter=State eq Moravia.Symfonie.Data.JobState'Order'

	projectSearchURL := moraviaJobsURL() + "?$filter=State eq " + "Moravia.Symfonie.Data.JobState'" + name + "'"
	// projectSearchURL := moraviaProjectsURL + "?$filter=Id eq 111111"
	fmt.Println(projectSearchURL)

	req, err := http.NewRequest("GET", projectSearchURL, body)
	if err != nil {
		log.Fatal(err)
	}
	// req.Header.Set("Content-Type", "application/json")
	authorization_value := "Bearer " + auth.Access_token
	req.Header.Set("Authorization", authorization_value)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	responseData, sErr := ioutil.ReadAll(resp.Body)
	if sErr != nil {
		log.Fatal(sErr)
	}
	fmt.Println(string(responseData))

	return nil

	// return json.NewDecoder(resp.Body).Decode(target)
}

func listJobs(auth AuthenticateResponse, target interface{}) error {
	var bodyString = ""

	body := strings.NewReader(bodyString)
	req, err := http.NewRequest("GET", moraviaJobsURL(), body)
	if err != nil {
		log.Fatal(err)
	}
	// req.Header.Set("Content-Type", "application/json")
	authorization_value := "Bearer " + auth.Access_token
	req.Header.Set("Authorization", authorization_value)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}

func createJob(job Job, auth AuthenticateResponse, target interface{}) error {
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(job)

	req, err := http.NewRequest("POST", moraviaJobsURL(), body)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	authorization_value := "Bearer " + auth.Access_token
	req.Header.Set("Authorization", authorization_value)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 {
		fmt.Println("Created job " + job.Name)
	}

	// responseData, sErr := ioutil.ReadAll(resp.Body)
	// if sErr != nil {
	// 	log.Fatal(sErr)
	// }
	// fmt.Println(string(responseData))

	return json.NewDecoder(resp.Body).Decode(target)
}

type Attachment struct {
	JobId              int
	Name               string
	FileType           string // Values - "Other", "Reference", "Source", "Target", "Analysis"
	AttachmentFilePath string `json:"-"`
}

// https://stackoverflow.com/questions/20205796/post-data-using-the-content-type-multipart-form-data
func mustOpen(filePath string) *os.File {
	fileReader, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	return fileReader
}

func upload(client *http.Client, url string, auth AuthenticateResponse, values map[string]io.Reader) (err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return err
		}

	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	authorization_value := "Bearer " + auth.Access_token
	req.Header.Set("Authorization", authorization_value)

	// Submit the request
	res, err := client.Do(req)
	if err != nil {
		return
	}

	// Check the response
	if res.StatusCode != http.StatusCreated {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	return
}

func uploadAttachment(attachment Attachment, auth AuthenticateResponse) {
	//
	// { JobId: 37, Name: "TestData.txt", FileType: "Other"}

	jsonData := new(bytes.Buffer)
	json.NewEncoder(jsonData).Encode(attachment)

	values := map[string]io.Reader{
		"file": mustOpen(attachment.AttachmentFilePath),
		"json": jsonData,
	}
	err := upload(httpClient, moraviaJobAttachmentsURL(), auth, values)
	if err != nil {
		log.Fatal(err)
	}
}

func listJobAttachments(auth AuthenticateResponse) {
	var bodyString = ""

	body := strings.NewReader(bodyString)
	req, err := http.NewRequest("GET", moraviaJobAttachmentsURL(), body)
	if err != nil {
		log.Fatal(err)
	}
	// req.Header.Set("Content-Type", "application/json")
	authorization_value := "Bearer " + auth.Access_token
	req.Header.Set("Authorization", authorization_value)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	responseData, sErr := ioutil.ReadAll(resp.Body)
	if sErr != nil {

	}
	fmt.Println(string(responseData))

	// return json.NewDecoder(resp.Body).Decode(target)
}

type Project struct {
	Id           int `yaml:"id"`
	Name         string
	Code         string
	ProjectState string
}

type Projects struct {
	Value []Project `json:"value"`
}

func findProject(name string, auth AuthenticateResponse, target interface{}) error {
	var bodyString = ""

	body := strings.NewReader(bodyString)

	// https://projects.moravia.com/api/V3/Jobs?$filter=State eq Moravia.Symfonie.Data.JobState'Order'

	projectSearchURL := moraviaProjectsURL() + "?$filter=contains(Name, '" + name + "')"
	// projectSearchURL := moraviaProjectsURL + "?$filter=Id eq 439741"
	fmt.Println(projectSearchURL)

	req, err := http.NewRequest("GET", projectSearchURL, body)
	if err != nil {
		log.Fatal(err)
	}
	// req.Header.Set("Content-Type", "application/json")
	authorization_value := "Bearer " + auth.Access_token
	req.Header.Set("Authorization", authorization_value)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	responseData, sErr := ioutil.ReadAll(resp.Body)
	if sErr != nil {
		log.Fatal(sErr)
	}
	fmt.Println(string(responseData))

	return nil

	// return json.NewDecoder(resp.Body).Decode(target)
}

func listProjects(auth AuthenticateResponse, target interface{}) error {
	var bodyString = ""

	body := strings.NewReader(bodyString)
	req, err := http.NewRequest("GET", moraviaProjectsURL(), body)
	if err != nil {
		log.Fatal(err)
	}
	// req.Header.Set("Content-Type", "application/json")
	authorization_value := "Bearer " + auth.Access_token
	req.Header.Set("Authorization", authorization_value)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}

////

func exampleListProjectsJobs(auth AuthenticateResponse) {
	projects := Projects{}
	listProjects(auth, &projects)

	fmt.Println(projects)

	jobs := Jobs{}
	listJobs(auth, &jobs)

	fmt.Println(jobs)
}

func exampleCreateJob(auth AuthenticateResponse) {
	job := Job{}
	job.Name = "Automation job"
	job.ProjectId = 1
	job.SourceLanguageCode = "en"
	job.TargetLanguageCodes = []string{"de", "nl"}
	createJob(job, auth, &job)
}

func exampleUploadAttachment(source *string, auth AuthenticateResponse) {
	attachment := Attachment{}
	attachment.JobId = 1
	attachment.Name = "en.xliff"
	attachment.FileType = "Source"
	attachment.AttachmentFilePath = *source

	uploadAttachment(attachment, auth)
}

////

func main() {
	moraviaConfigFilepath := getenv("moravia_config", "moravia.yml")

	var configuration MoraviaConfiguration
	configuration.readFromFile(moraviaConfigFilepath)

	fmt.Println(configuration)

	clientID := getenv("moravia_client_id", "")
	clientSecret := getenv("moravia_client_secret", "")
	serviceAccount := getenv("moravia_service_account", "")

	if clientID == "" {
		fmt.Println("Client ID is required\n")
		os.Exit(1)
	}

	if clientSecret == "" {
		fmt.Println("Client secret is required\n")
		os.Exit(1)
	}

	if serviceAccount == "" {
		fmt.Println("Service account is required\n")
		os.Exit(1)
	}

	if configuration.Project.Id == 0 {
		fmt.Println("Project ID is required\n")
		os.Exit(1)
	}

	if configuration.Job_template.Source == "" {
		fmt.Println("Source is required\n")
		os.Exit(1)
	}
	// Test opening the source
	mustOpen(configuration.Job_template.Source)

	if configuration.Job_template.Source_language == "" {
		fmt.Println("Source language is required\n")
		os.Exit(1)
	}

	// TODO: Alex - need a check against target languages

	auth := AuthenticateResponse{}
	authenticate(clientID, clientSecret, serviceAccount, &auth)

	currentTime := time.Now()
	// Golang wat - https://gobyexample.com/time-formatting-parsing
	dateString := currentTime.Format("20060102")

	job := Job{}
	job.Name = dateString + " - " + configuration.Job_template.Name
	job.ProjectId = configuration.Project.Id
	job.SourceLanguageCode = configuration.Job_template.Source_language
	job.TargetLanguageCodes = configuration.Job_template.Target_languages
	err := createJob(job, auth, &job)
	if err != nil {
		log.Fatal(err)
	}

	_, filename := filepath.Split(configuration.Job_template.Source)

	attachment := Attachment{}
	attachment.JobId = job.Id
	attachment.Name = filename
	attachment.FileType = "Source"
	attachment.AttachmentFilePath = configuration.Job_template.Source

	uploadAttachment(attachment, auth)

	portalURL := moraviaPortalJobDetailsURL(job)

	//
	// --- Step Outputs: Export Environment Variables for other Steps:
	// You can export Environment Variables for other Steps with
	//  envman, which is automatically installed by `bitrise setup`.
	// A very simple example:
	cmdLog, err := exec.Command("bitrise", "envman", "add", "--key", "MORAVIA_JOB_DETAIL_URL", "--value", portalURL).CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to expose output with envman, error: %#v | output: %s", err, cmdLog)
	}
	// You can find more usage examples on envman's GitHub page
	//  at: https://github.com/bitrise-io/envman

	os.Exit(0)
}
