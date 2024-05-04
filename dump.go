package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

func getAssignedTests(c *http.Client) ([]AssignedTest, error) {
	req, _ := http.NewRequest(
		"GET",
		"https://pts.measureup.com/web/phpfiles/tests/getAssignedTestUsers.php",
		nil,
	)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf(resp.Status)
	}

	var tests []AssignedTest
	return tests, json.Unmarshal(body, &tests)
}

func getSkillGroups(c *http.Client, dest string, test AssignedTest) ([]SkillGroup, error) {
	params := make(url.Values)
	params.Set("directory", "../../instances/MUP/")
	params.Set("test", test.Test)
	params.Set("key", test.KeyID)
	params.Set("role", "1")
	params.Set("license", strconv.Itoa(test.License))

	req, _ := http.NewRequest(
		"POST",
		"https://pts.measureup.com/web/PBS/LMS/phpFilesLMS/getTestSkillgroups.php",
		strings.NewReader(params.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf(resp.Status)
	}

	err = os.WriteFile(filepath.Join(dest, "skillGroups.json"), body, 0o644)
	if err != nil {
		return nil, err
	}

	var groups []SkillGroup
	return groups, json.Unmarshal(body, &groups)
}

func getTextDB(c *http.Client, dest string, test AssignedTest) (TextDB, error) {
	params := make(url.Values)
	params.Set("test", test.Test)
	params.Set("shortname", test.VendorTest)

	req, _ := http.NewRequest(
		"POST",
		"https://pts.measureup.com/web/phpfiles/obtainQuestions.php",
		strings.NewReader(params.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(filepath.Join(dest, "textdb.json"), body, 0o644)
	if err != nil {
		return nil, err
	}

	var texts TextDB
	return texts, json.Unmarshal(body, &texts)
}

func getQuestion(c *http.Client, dest string, questionName string) (*Question, error) {
	req, _ := http.NewRequest(
		"GET",
		"https://pts.measureup.com/web/instances/MUP/model/questions/"+questionName+".json",
		nil,
	)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf(resp.Status)
	}

	err = os.WriteFile(dest, body, 0o644)
	if err != nil {
		return nil, err
	}

	var question Question
	return &question, json.Unmarshal(body, &question)
}

func getSlide(c *http.Client, dest string, slideName string) (*QuestionSlide, error) {
	req, _ := http.NewRequest(
		"GET",
		"https://pts.measureup.com/web/instances/MUP/views/"+slideName+".json",
		nil,
	)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf(resp.Status)
	}

	err = os.WriteFile(dest, body, 0o644)
	if err != nil {
		return nil, err
	}

	var slide QuestionSlide
	return &slide, json.Unmarshal(body, &slide)
}

func getImage(c *http.Client, dest string, imageName string) error {
	req, _ := http.NewRequest(
		"GET",
		"https://pts.measureup.com/web/instances/MUP/"+imageName,
		nil,
	)
	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	} else if resp.StatusCode != 200 {
		return fmt.Errorf(resp.Status)
	}

	return os.WriteFile(dest, body, 0o644)
}

type transport struct {
	http.Transport
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 OPR/107.0.0.0")
	return t.Transport.RoundTrip(req)
}

func sessionCookie(session string) (*url.URL, []*http.Cookie) {
	u, _ := url.Parse("https://measureup.com")
	return u, []*http.Cookie{{
		Name:   "PHPSESSID",
		Value:  session,
		Domain: u.Hostname(),
	}}
}

func dump(session string, testName string) ([]AssignedTest, error) {
	cookies, _ := cookiejar.New(nil)
	cookies.SetCookies(sessionCookie(session))

	c := &http.Client{
		Jar:       cookies,
		Transport: &transport{},
	}

	tests, err := getAssignedTests(c)
	if err != nil {
		return nil, err
	}
	tests = slices.DeleteFunc(tests, func(test AssignedTest) bool {
		return test.ProductID == 0
	})

	if testName == "" {
		return tests, nil
	}

	idx := slices.IndexFunc(tests, func(test AssignedTest) bool {
		return strings.EqualFold(test.VendorTest, testName)
	})
	if idx < 0 {
		return tests, fmt.Errorf("no test named '%s' found", testName)
	}

	path := filepath.Join("out", "dump", strings.ToLower(testName))
	os.MkdirAll(filepath.Join(path, "questions"), 0o755)
	os.MkdirAll(filepath.Join(path, "images"), 0o755)
	os.MkdirAll(filepath.Join(path, "slides"), 0o755)

	test := tests[idx]
	log.Printf("%s (%s) - %s\n", test.VendorTest, test.Test, test.VendorName)
	groups, err := getSkillGroups(c, path, test)
	if err != nil {
		return tests, err
	}
	_, err = getTextDB(c, path, test)
	if err != nil {
		return tests, err
	}

	for _, group := range groups {
		for i := 0; i < len(group.Questions); i++ {
			groupQuestion := group.Questions[i]
			log.Printf("%s\n", groupQuestion.Name)

			_, qfn, _ := strings.Cut(groupQuestion.Name, "/")
			question, err := getQuestion(
				c,
				filepath.Join(path, "questions", qfn+".json"),
				groupQuestion.Name,
			)
			if err != nil {
				return tests, err
			}

			parts := strings.Split(question.StartSlide.Value, "/")
			slide, err := getSlide(
				c,
				filepath.Join(path, "slides", parts[1]+".json"),
				question.StartSlide.Value,
			)
			if err != nil {
				return tests, err
			}

			images := question.Images()
			if slide.View.Image != "" {
				images = append(images, QuestionImage{
					Name: slide.View.Image, Alt: slide.View.Alt,
				})
			}
			for _, image := range slide.Images {
				images = append(images, QuestionImage{
					Name: image.Image, Alt: image.Alt,
				})
			}

			for _, image := range images {
				log.Println("  ", image.Name)

				parts := strings.Split(image.Name, "/")
				ifn := qfn + "-" + parts[len(parts)-1]
				err := getImage(
					c,
					filepath.Join(path, "images", ifn),
					image.Name,
				)
				if err != nil {
					return tests, err
				}
			}

			if groupQuestion.Type == "caseStudy" {
				for _, opt := range slide.CaseStudy[0].Options {
					if opt.CSContext != "" {
						group.Questions = append(group.Questions, SkillGroupQuestion{
							Name: groupQuestion.Name + "_" + opt.CSContext,
							Type: "caseStudyQuestion",
						})
					}
				}
			}
		}
	}

	return tests, nil
}
