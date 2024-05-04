package main

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

const MaxOptions = 10

func CSVColumns() []string {
	var cols []string

	cols = append(cols, "ID", "Text", "Explanation", "Exhibits")
	for i := 1; i <= MaxOptions; i++ {
		cols = append(cols, "Option-"+strconv.Itoa(i))
	}
	cols = append(cols, "Type", "Image", "Answer")
	return cols
}

type AssignedTest struct {
	ID          int
	Test        string
	TestName    string
	KeyID       string
	ProductID   int
	ProductType string
	VendorName  string
	VendorTest  string
	License     int
	Paused      bool `json:"testPaused"`
}

type SkillGroup struct {
	ID        int
	Name      string
	Questions []SkillGroupQuestion
}

type SkillGroupQuestion struct {
	Name string
	Type string `json:"question_type"`
	Stem string
}

type TextDB map[string]string

func (db TextDB) Get(key string) string {
	return db[strings.TrimPrefix(key, "$$")]
}

type Question struct {
	Stem struct {
		Value string
	}
	Type struct {
		Value string
	}
	Explanation struct {
		Value string
	}
	StartSlide struct {
		Value string
	}
	Exhibit struct {
		Content interface{}
	}
	Models []struct {
		Model     string
		Options   interface{}
		Correct   interface{}
		ByDefault []struct {
			Label string
		}
	}
}

func (q Question) Images() QuestionImages {
	var res QuestionImages

	switch content := q.Exhibit.Content.(type) {
	case []interface{}:
		for _, c := range content {
			c := c.(map[string]interface{})
			image := c["image"].(string)
			alt := c["alt"].(string)
			res = append(res, QuestionImage{
				Name: image,
				Alt:  alt,
			})
		}
	}
	return res
}

func (q Question) Correct() [][]string {
	var res [][]string

	for _, m := range q.Models {
		var a []string

		switch correct := m.Correct.(type) {
		case string:
			a = append(a, correct)
		case []interface{}:
			for _, c := range correct {
				a = append(a, c.(string))
			}
		}

		switch options := m.Options.(type) {
		case []interface{}:
			for _, opt := range options {
				switch opt := opt.(type) {
				case map[string]interface{}:
					correct := opt["correct"].(string)
					if strings.EqualFold(correct, "yes") {
						a = append(a, opt["row"].(string))
					}
				}
			}
		}

		res = append(res, a)
	}
	return res
}

func (q Question) Statements() [][]string {
	var res [][]string

	for _, m := range q.Models {
		var a []string

		switch options := m.Options.(type) {
		case []interface{}:
			for _, opt := range options {
				switch opt := opt.(type) {
				case map[string]interface{}:
					statement := opt["row"].(string)
					a = append(a, statement)
				}
			}
		}

		res = append(res, a)
	}
	return res
}

type QuestionImage struct {
	Name string
	Alt  string
}

type QuestionImages []QuestionImage

func (qi QuestionImages) HTML() string {
	var images []string
	for _, image := range qi {
		images = append(images, fmt.Sprintf(
			`<img src="%s" alt="%s" class="exhibit">`,
			image.Name,
			strings.ReplaceAll(image.Alt, `"`, `\"`),
		))
	}
	return strings.Join(images, "\n<br>\n")
}

type QuestionSlide struct {
	View struct {
		ID    string
		Image string
		Alt   string
	}
	RadioButtons []struct {
		ID    string
		Value string
	}
	CheckBoxes []struct {
		ID    string
		Value string
	}
	Selects []struct {
		ID    string
		Model string
	}
	SelectPlaceMup []struct {
		ID      string
		Options []struct {
			Alt string
		}
	}
	Images []struct {
		Image string
		Alt   string
	}
	CaseStudy []struct {
		Options []struct {
			Label     string
			CSContext string
		}
	}
}

type Record interface {
	Record() []string
}

type SingleChoice struct {
	ID          string
	Group       SkillGroup
	Text        string
	Explanation string
	Exhibits    QuestionImages
	Options     []string
	Answer      int
}

func NewSingleChoice(
	id string,
	textDB TextDB,
	group SkillGroup,
	question Question,
	images QuestionImages,
	slide QuestionSlide,
) *SingleChoice {
	var options []string
	for _, btn := range slide.RadioButtons {
		options = append(options, textDB.Get(btn.Value))
	}

	var answer int
	correctBtn := question.Correct()[0][0]
	for _, btn := range slide.RadioButtons {
		if btn.ID == correctBtn {
			answer = slices.Index(options, textDB.Get(btn.Value)) + 1
			break
		}
	}

	return &SingleChoice{
		ID:          id,
		Group:       group,
		Text:        textDB.Get(question.Stem.Value),
		Explanation: textDB.Get(question.Explanation.Value),
		Exhibits:    images,
		Options:     options,
		Answer:      answer,
	}
}

func (sc *SingleChoice) Record() []string {
	record := append([]string{},
		sc.ID, sc.Text, sc.Explanation, sc.Exhibits.HTML())

	options := append([]string{}, sc.Options...)
	if len(options) > MaxOptions {
		panic("MaxOptions is too low")
	}
	for i := len(options); i < MaxOptions; i++ {
		options = append(options, "")
	}

	record = append(record, options...)
	record = append(record, "singleChoice", "", strconv.Itoa(sc.Answer))

	return record
}

type MultipleChoice struct {
	ID          string
	Group       SkillGroup
	Text        string
	Explanation string
	Exhibits    QuestionImages
	Options     []string
	Answers     []int
}

func NewMultipleChoice(
	id string,
	textDB TextDB,
	group SkillGroup,
	question Question,
	images QuestionImages,
	slide QuestionSlide,
) *MultipleChoice {
	var options []string
	for _, btn := range slide.CheckBoxes {
		options = append(options, textDB.Get(btn.Value))
	}

	var answers []int
	correctBtns := question.Correct()[0]
	for _, btn := range slide.CheckBoxes {
		if slices.Index(correctBtns, btn.ID) > -1 {
			idx := slices.Index(options, textDB.Get(btn.Value))
			answers = append(answers, idx+1)
		}
	}
	slices.Sort(answers)

	return &MultipleChoice{
		ID:          id,
		Group:       group,
		Text:        textDB.Get(question.Stem.Value),
		Explanation: textDB.Get(question.Explanation.Value),
		Exhibits:    images,
		Options:     options,
		Answers:     answers,
	}
}

func (mc *MultipleChoice) Record() []string {
	record := append([]string{},
		mc.ID, mc.Text, mc.Explanation, mc.Exhibits.HTML())

	options := append([]string{}, mc.Options...)
	if len(options) > MaxOptions {
		panic("MaxOptions is too low")
	}
	for i := len(options); i < MaxOptions; i++ {
		options = append(options, "")
	}

	record = append(record, options...)
	answers, _ := json.Marshal(mc.Answers)
	record = append(record, "multipleChoice", "", string(answers))

	return record
}

type LiveScreen struct {
	ID          string
	Group       SkillGroup
	Text        string
	Explanation string
	Exhibits    QuestionImages
	ImageName   string
	ImageAlt    string
	Options     [][]string
	Answers     []int
}

func NewLiveScreen(
	id string,
	textDB TextDB,
	group SkillGroup,
	question Question,
	images QuestionImages,
	slide QuestionSlide,
) *LiveScreen {
	var options [][]string
	for _, sel := range slide.Selects {
		for _, m := range question.Models {
			if m.Model == sel.Model {
				var opts []string
				for _, opt := range m.Options.([]interface{}) {
					opts = append(opts, textDB.Get(opt.(string)))
				}
				options = append(options, opts)
			}
		}
	}

	var answers []int
	for i, correct := range question.Correct() {
		answers = append(answers, slices.Index(options[i], textDB.Get(correct[0])))
	}

	return &LiveScreen{
		ID:          id,
		Group:       group,
		Text:        textDB.Get(question.Stem.Value),
		Explanation: textDB.Get(question.Explanation.Value),
		Exhibits:    images,
		ImageName:   slide.View.Image,
		ImageAlt:    slide.View.Alt,
		Options:     options,
		Answers:     answers,
	}
}

func (ls *LiveScreen) ImageHTML() string {
	return fmt.Sprintf(
		`<img src="%s" alt="%s" class="image">`,
		ls.ImageName,
		strings.ReplaceAll(ls.ImageAlt, `"`, `\"`),
	)
}

func (sc *LiveScreen) Record() []string {
	record := append([]string{},
		sc.ID, sc.Text, sc.Explanation, sc.Exhibits.HTML())

	var options []string
	for _, opts := range sc.Options {
		options = append(options, strings.Join(opts, " ╱ "))
	}
	if len(options) > MaxOptions {
		panic("MaxOptions is too low")
	}
	for i := len(options); i < MaxOptions; i++ {
		options = append(options, "")
	}

	record = append(record, options...)
	answers, _ := json.Marshal(sc.Answers)
	record = append(record, "liveScreen", sc.ImageHTML(), string(answers))

	return record
}

type ContentTable struct {
	ID          string
	Group       SkillGroup
	Text        string
	Explanation string
	Exhibits    QuestionImages
	Options     []string
	Answers     []int
}

func NewContentTable(
	id string,
	textDB TextDB,
	group SkillGroup,
	question Question,
	images QuestionImages,
	slide QuestionSlide,
) *ContentTable {
	var options []string
	for _, m := range question.Models {
		for _, opt := range m.Options.([]interface{}) {
			opt := opt.(map[string]interface{})
			options = append(options, textDB.Get(opt["row"].(string)))
		}
	}

	var answers []int
	for _, correct := range question.Correct()[0] {
		answers = append(answers, slices.Index(options, textDB.Get(correct)))
	}
	slices.Sort(answers)

	return &ContentTable{
		ID:          id,
		Group:       group,
		Text:        textDB.Get(question.Stem.Value),
		Explanation: textDB.Get(question.Explanation.Value),
		Exhibits:    images,
		Options:     options,
		Answers:     answers,
	}
}

func (ct *ContentTable) Record() []string {
	record := append([]string{},
		ct.ID, ct.Text, ct.Explanation, ct.Exhibits.HTML())

	options := append([]string{}, ct.Options...)
	if len(options) > MaxOptions {
		panic("MaxOptions is too low")
	}
	for i := len(options); i < MaxOptions; i++ {
		options = append(options, "")
	}

	record = append(record, options...)
	answers, _ := json.Marshal(ct.Answers)
	record = append(record, "contentTable", "", string(answers))

	return record
}

type BuildList struct {
	ID          string
	Group       SkillGroup
	Text        string
	Explanation string
	Exhibits    QuestionImages
	Options     []string
	Answers     []int
}

func NewBuildList(
	id string,
	textDB TextDB,
	group SkillGroup,
	question Question,
	images QuestionImages,
	slide QuestionSlide,
) *BuildList {
	var options []string
	for _, m := range question.Models {
		for _, opt := range m.ByDefault {
			options = append(options, textDB.Get(opt.Label))
		}
	}

	var answers []int
	for _, c := range question.Correct()[0] {
		n, _ := strconv.Atoi(c)
		answers = append(answers, n+1)
	}

	return &BuildList{
		ID:          id,
		Group:       group,
		Text:        textDB.Get(question.Stem.Value),
		Explanation: textDB.Get(question.Explanation.Value),
		Exhibits:    images,
		Options:     options,
		Answers:     answers,
	}
}

func (bl *BuildList) Record() []string {
	record := append([]string{},
		bl.ID, bl.Text, bl.Explanation, bl.Exhibits.HTML())

	options := append([]string{}, bl.Options...)
	if len(options) > MaxOptions {
		panic("MaxOptions is too low")
	}
	for i := len(options); i < MaxOptions; i++ {
		options = append(options, "")
	}

	record = append(record, options...)
	answers, _ := json.Marshal(bl.Answers)
	record = append(record, "buildList", "", string(answers))

	return record
}

type SelectPlaceMup struct {
	ID          string
	Group       SkillGroup
	Text        string
	Explanation string
	Exhibits    QuestionImages
	ImageName   string
	ImageAlt    string
	Options     []string
	Answers     []int
}

func NewSelectPlaceMup(
	id string,
	textDB TextDB,
	group SkillGroup,
	question Question,
	images QuestionImages,
	slide QuestionSlide,
) *SelectPlaceMup {
	var options []string
	for _, sel := range slide.SelectPlaceMup {
		for _, opt := range sel.Options {
			options = append(options, opt.Alt)
		}
	}

	answers := make([]int, 0)

	if len(slide.Images) < 1 {
		panic("No image for SelectPlaceMup")
	} else if len(slide.Images) > 1 {
		panic("Too many images for SelectPlaceMup")
	}

	return &SelectPlaceMup{
		ID:          id,
		Group:       group,
		Text:        textDB.Get(question.Stem.Value),
		Explanation: textDB.Get(question.Explanation.Value),
		Exhibits:    images,
		ImageName:   slide.Images[0].Image,
		ImageAlt:    slide.Images[0].Alt,
		Options:     options,
		Answers:     answers,
	}
}

func (sp *SelectPlaceMup) ImageHTML() string {
	return fmt.Sprintf(
		`<img src="%s" alt="%s" class="image">`,
		sp.ImageName,
		strings.ReplaceAll(sp.ImageAlt, `"`, `\"`),
	)
}

func (sp *SelectPlaceMup) Record() []string {
	record := append([]string{},
		sp.ID, sp.Text, sp.Explanation, sp.Exhibits.HTML())

	options := append([]string{}, sp.Options...)
	if len(options) > MaxOptions {
		panic("MaxOptions is too low")
	}
	for i := len(options); i < MaxOptions; i++ {
		options = append(options, "")
	}

	record = append(record, options...)
	answers, _ := json.Marshal(sp.Answers)
	record = append(record, "selectPlaceMup", sp.ImageHTML(), string(answers))

	return record
}
