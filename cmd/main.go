package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	commonUrl  string = "https://enroll.spbstu.ru/applications-manager/api/v1/admission-list/form-rating?applicationEducationLevel=%s&directioneducationformid=%d&directionId=%d"
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
	fmt.Println(string(body))
	if err != nil {
		return nil, err
	}
	var resultResp Response
	if err = json.Unmarshal(body, &resultResp); err != nil {
		return nil, err
	}
	return resultResp.Users, nil
}
const (
    firstDirID = 0
    lastDirID = 303
    maxWorkers = 8
)


type Result struct {
    Users []User
    Err error
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
type UserDb map[Snils][]*User

func (db UserDb) addUserRow(user *User) {
    db[Snils(user.UserSnils)] = append(db[Snils(user.UserSnils)], user)
}
func (db UserDb) PrinUserRow(user *User) {
    row, ok := db[Snils(user.UserSnils)]
    if !ok {
        fmt.Printf("User not found\n")
        return
    }
    
    fmt.Printf("User: %s\n", user.UserSnils)
    for _, u := range row {
        fmt.Printf("Специальность: %s, сумма баллов: %d, приоритет: %s, позиция в списке: %d, оригинал: %t\n", u.DirectionEducationForm.TitleForeign)
    }
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second * 30)
    reqChan := make(chan string, maxWorkers)
    resultChan := make(chan Result, maxWorkers)
    wg := &sync.WaitGroup{}
    db := make(UserDb, 4096)
    defer cancel()

    
    for range maxWorkers {
        go Worker(ctx, wg, reqChan, resultChan)
    }

    for dirID := firstDirID; dirID <= lastDirID; dirID++ {
       reqChan <- fmt.Sprintf(commonUrl, EducationLevelMaster, EducationFormIdFullTime, dirID)
    } 
    close(reqChan)

    wg.Wait()
    close(resultChan)

    for res := range resultChan {
        if res.Err != nil {
            fmt.Fprintf(os.Stderr, "error occured while making request: %v", res.Err)
        } else {
            for _, u := range res.Users {
                db.addUserRow(&u)
            }    
        }
    }




}


