package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var skillGroup2questionType = map[string]string{
	"singleChoice":   "singleChoice",
	"multipleChoice": "multipleChoice",
	"simulation":     "liveScreen",
}

func removeEscapes(b []byte) []byte {
	s := string(b)
	s = strings.ReplaceAll(s, "\\\\", "")
	s = strings.ReplaceAll(s, "<br />", "<br>")
	return []byte(s)
}

func readJSON(path string, out interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(removeEscapes(data), out)
}

func copyMedia(questionName string, imageName string, from string, to string) string {
	parts := strings.Split(imageName, "/")
	ifn := questionName + "-" + parts[len(parts)-1]
	ifop := filepath.Join(from, ifn)
	ifnp := filepath.Join(to, ifn)

	buf, _ := os.ReadFile(ifop)
	os.WriteFile(ifnp, buf, 0o644)

	return ifn
}

func produce(testName string) error {
	src := filepath.Join("out", "dump", testName)

	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("'%s' was not found", testName)
	} else if err != nil {
		return fmt.Errorf("accessing '%s': %v", testName, err)
	}

	media := filepath.Join("out", "collection.media")
	os.MkdirAll(media, 0o755)

	var groups []SkillGroup
	if err := readJSON(filepath.Join(src, "skillGroups.json"), &groups); err != nil {
		return err
	}

	var textDB TextDB
	if err := readJSON(filepath.Join(src, "textdb.json"), &textDB); err != nil {
		return err
	}

	var records []Record

	for _, group := range groups {
		for i := 0; i < len(group.Questions); i++ {
			groupQuestion := group.Questions[i]
			_, qfn, _ := strings.Cut(groupQuestion.Name, "/")
			qfp := filepath.Join(src, "questions", qfn+".json")

			log.Printf("Got %s (%s)\n", qfn, groupQuestion.Type)

			var question Question
			if err := readJSON(qfp, &question); err != nil {
				return err
			}

			images := question.Images()
			for i := range images {
				images[i].Name = copyMedia(
					qfn,
					images[i].Name,
					filepath.Join(src, "images"),
					media,
				)
			}

			parts := strings.Split(question.StartSlide.Value, "/")
			sfp := filepath.Join(src, "slides", parts[1]+".json")

			var slide QuestionSlide
			if err := readJSON(sfp, &slide); err != nil {
				return err
			} else if slide.View.Image != "" {
				slide.View.Image = copyMedia(
					qfn,
					slide.View.Image,
					filepath.Join(src, "images"),
					media,
				)
			}
			for i := range slide.Images {
				slide.Images[i].Image = copyMedia(
					qfn,
					slide.Images[i].Image,
					filepath.Join(src, "images"),
					media,
				)
			}

			_, id, _ := strings.Cut(groupQuestion.Name, "_")

			if groupQuestion.Type == "caseStudyQuestion" {
				groupQuestion.Type = skillGroup2questionType[question.Type.Value]
			}

			switch groupQuestion.Type {
			case "caseStudy":
				for _, opt := range slide.CaseStudy[0].Options {
					if opt.CSContext != "" {
						group.Questions = append(group.Questions, SkillGroupQuestion{
							Name: groupQuestion.Name + "_" + opt.CSContext,
							Type: "caseStudyQuestion",
						})
					}
				}
			case "singleChoice":
				records = append(records, NewSingleChoice(
					id,
					textDB,
					group,
					question,
					images,
					slide,
				))
			case "multipleChoice":
				records = append(records, NewMultipleChoice(
					id,
					textDB,
					group,
					question,
					images,
					slide,
				))
			case "liveScreen":
				records = append(records, NewLiveScreen(
					id,
					textDB,
					group,
					question,
					images,
					slide,
				))
			case "contentTable":
				records = append(records, NewContentTable(
					id,
					textDB,
					group,
					question,
					images,
					slide,
				))
			case "buildList":
				fallthrough
			case "buildListReorder":
				records = append(records, NewBuildList(
					id,
					textDB,
					group,
					question,
					images,
					slide,
				))
			case "selectPlaceMup":
				records = append(records, NewSelectPlaceMup(
					id,
					textDB,
					group,
					question,
					images,
					slide,
				))
			default:
				log.Println("Skipping...")
			}
		}
	}

	f, _ := os.Create(filepath.Join("out", strings.ToLower(testName)+".csv"))
	defer f.Close()

	fmt.Fprintln(f, "#separator:,")
	fmt.Fprintln(f, "#notetype:MeasureUpCard")
	fmt.Fprintln(f, "#columns:"+strings.Join(CSVColumns(), ","))

	w := csv.NewWriter(f)
	for _, record := range records {
		w.Write(record.Record())
	}
	w.Flush()

	return nil
}
