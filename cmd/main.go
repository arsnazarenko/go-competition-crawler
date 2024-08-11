package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	commonUrl string = "https://enroll.spbstu.ru/applications-manager/api/v1/admission-list/form-rating?applicationEducationLevel=%s&directioneducationformid=%d&directionId=%d"
)

type EducationLevel string

const (
	EducationLevelBachelor EducationLevel = "BACHELOR" // Бакалавриат
	EducationLevelMaster   EducationLevel = "MASTER"   // Магистратура
	EducationLevelGraduate EducationLevel = "GRADUATE" // Аспирантура
)

type EducationFormId uint8

const (
	EducationFormIdCorrespondence EducationFormId = iota + 1 // Заочное
	EducationFormIdFullTime                                  // Очное
	EducationFormIdPartTime                                  // Очно-заочное
)

type Subject struct {
	Title      string `json:"title"`
	ExternalId string `json:"externalId"`
	Score      uint16 `json:"score"`
}
type DirectionEducationForm struct {
	Id           uint64 `json:"id"`
	Title        string `json:"title"`
	TitleForeign string `json:"titleForeign"`
	ExternalId   string `json:"externalId"`
}

type DirectionPaymentForm struct {
	Id           uint64 `json:"id"`
	Title        string `json:"title"`
	TitleForeign string `json:"titleForeign"`
	ExternalId   string `json:"externalId"`
}

type User struct {
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

type Response struct {
	Users              []User `json:"list"`
	Log                string `json:"-"`
	DirectionCapacity  uint64 `json:"-"`
	Total              uint64 `json:"-"`
	TotalWithOriginals uint64 `json:"-"`
}

var (
	headers = map[string]string{
		"User-Agent":      "Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0",
		"Accept":          "application/json",
		"Accept-Language": "en-US,en;q=0.5",
	}
)

func GetCompetitionList(eduLevel EducationLevel, eduForm EducationFormId, directionId uint64) ([]User, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(commonUrl, eduLevel, eduForm, directionId), nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := http.DefaultClient
	resp, err := client.Do(req)
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

func main() {
	users, err := GetCompetitionList(EducationLevelMaster, EducationFormIdFullTime, 307)
	if err != nil {
		panic(err)
	}
	fmt.Println(users)

}
