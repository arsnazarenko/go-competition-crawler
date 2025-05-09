package main

import (
	"context"
	"encoding/json"
	"fmt"
	worker_pool "go-competiotion-crawler/internal"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	commonUrl string = "https://enroll.spbstu.ru/applications-manager/api/v1/admission-list/form-rating?applicationEducationLevel=%s&directioneducationformid=%d&directionId=%d"
)

type (
	EducationFormId uint8
	EducationLevel  string
)

const (
	EducationLevelBachelor EducationLevel = "BACHELOR" // Бакалавриат
	EducationLevelMaster   EducationLevel = "MASTER"   // Магистратура
	EducationLevelGraduate EducationLevel = "GRADUATE" // Аспирантура

	EducationFormIdCorrespondence EducationFormId = iota + 1 // Заочное
	EducationFormIdFullTime                                  // Очное
	EducationFormIdPartTime                                  // Очно-заочное
)

type (
	Subject struct {
		Title      string `json:"title"`
		ExternalId string `json:"externalId"`
		Score      uint16 `json:"score"`
	}
	DirectionEducationForm struct {
		Id           uint64 `json:"id"`
		Title        string `json:"title"`
		TitleForeign string `json:"titleForeign"`
		ExternalId   string `json:"externalId"`
	}

	DirectionPaymentForm struct {
		Id           uint64 `json:"id"`
		Title        string `json:"title"`
		TitleForeign string `json:"titleForeign"`
		ExternalId   string `json:"externalId"`
	}

	User struct {
		ApplicationEducationLevel string `json:"applicationEducationLevel"`

		DirectionEducationForm DirectionEducationForm `json:"directionEducationForm"`
		DirectionPaymentForm   DirectionPaymentForm   `json:"directionPaymentForm"`

		DirectionId uint64 `json:"directionId"`

		Subjects []Subject `json:"subjects"`

		UserFullName              string  `json:"userFullName"`
		UserSnils                 string  `json:"userSnils"`
		UserUniqueId              string  `json:"userUniqueId"`
		UserExternalId            string  `json:"userExternalId"`
		Priority                  uint16  `json:"priority"`
		WithoutExam               bool    `json:"withoutExam"`
		FullScore                 uint16  `json:"fullScore"`
		SubjectScore              uint16  `json:"subjectScore"`
		HasFeature                bool    `json:"hasFeature"`
		HasAgreement              bool    `json:"hasAgreement"`
		HasOriginalDocuments      bool    `json:"hasOriginalDocuments"`
		AchievementScore          uint16  `json:"achievementScore"`
		AchievementScoreExtra     uint16  `json:"achievementScoreExtra"`
		HasSpecialFeature         bool    `json:"hasSpecialFeature"`
		HasAchievement            bool    `json:"hasAchievement"`
		HasOlympiad               bool    `json:"hasOlympiad"`
		HasOlympiadReset          bool    `json:"hasOlympiadReset"`
		HasGovernmentContract     bool    `json:"hasGovernmentContract"`
		NeedDormitory             bool    `json:"needDormitory"`
		CertificateAverage        float32 `json:"certificateAverage"`
		CertificateProfileAverage float32 `json:"certificateProfileAverage"`
		State                     string  `json:"state"`
	}

	Response struct {
		Users              []User `json:"list"`
		Log                string `json:"-"`
		DirectionCapacity  uint64 `json:"-"`
		Total              uint64 `json:"-"`
		TotalWithOriginals uint64 `json:"-"`
	}
)

var (
	headers = map[string]string{
		"User-Agent":      "Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0",
		"Accept":          "application/json",
		"Accept-Language": "en-US,en;q=0.5",
	}
)

func GetCompetitionList(ctx context.Context, url string) ([]User, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var resultResp Response
	if err = json.Unmarshal(body, &resultResp); err != nil {
		return nil, err
	}
	if len(resultResp.Users) < 1 {
		return nil, fmt.Errorf("have not users for this directionId\n")
	}
	return resultResp.Users, nil
}

const (
	firstDirID = 200
	lastDirID  = 1000
	maxWorkers = 8
	totalJobs  = lastDirID - firstDirID
)

type Result struct {
	Users []User
	Err   error
}

func Worker(ctx context.Context, wg *sync.WaitGroup, inChan <-chan string, outChan chan<- Result) {
	defer wg.Done()
	for url := range inChan {
		users, err := GetCompetitionList(ctx, url)
		outChan <- Result{
			Users: users,
			Err:   err,
		}
	}
}

type Snils string
type UserInfo struct {
	position uint64
	u        *User
}
type UserDb map[Snils][]UserInfo

func (db UserDb) addUserRow(userInfo UserInfo) {
	db[Snils(userInfo.u.UserSnils)] = append(db[Snils(userInfo.u.UserSnils)], userInfo)
}
func (db UserDb) PrinUserRow(snils Snils) {
	row, ok := db[Snils(snils)]
	if !ok {
		fmt.Printf("User not found\n")
		return
	}

	fmt.Printf("User: %s\n", snils)
	for _, info := range row {
		fmt.Printf("Специальность: (%d) %s, сумма баллов: %d, приоритет: %d, позиция в списке: %d, оригинал: %t\n", info.u.DirectionId, info.u.Subjects[0].Title, info.u.FullScore, info.u.Priority, info.position, info.u.HasOriginalDocuments)
	}
}
func main() {
	worker_pool.Example1()
}
func main1() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	reqChan := make(chan string, totalJobs)
	resultChan := make(chan Result, totalJobs)
	wg := &sync.WaitGroup{}
	db := make(UserDb, 4096)
	defer cancel()

	wg.Add(maxWorkers)
	for range maxWorkers {
		go Worker(ctx, wg, reqChan, resultChan)
	}
	for dirID := firstDirID; dirID <= lastDirID; dirID++ {
		reqChan <- fmt.Sprintf(commonUrl, EducationLevelMaster, EducationFormIdFullTime, dirID)
	}
	close(reqChan)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for res := range resultChan {
		if res.Err != nil {
			fmt.Fprintf(os.Stderr, "error occured while making request: %v\n", res.Err)
		} else {
			for position, u := range res.Users {
				db.addUserRow(UserInfo{
					position: uint64(position),
					u:        &u,
				})
			}
		}
	}
	db.PrinUserRow("163-124-528 36")

}
