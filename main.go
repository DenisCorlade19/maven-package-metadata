package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/vifraa/gopom"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"
)

// Different structs that are needed throughout the implementation

type Version struct {
	Number      string    `json:number`
	PublishedAt time.Time `json:published_at`
}

type OutputVersion struct {
	TimeStamp    string            `json:"timestamp"`
	Dependencies map[string]string `json:"dependencies"`
}

type VersionDependencies struct {
	Name     string                   `json:"name"`
	Versions map[string]OutputVersion `json:"versions"`
}

type MyDependency struct {
	GroupId         string
	Artifact        string
	RequiredVersion string
}

func (v *Version) UnmarshalJSON(b []byte) error {
	var dat map[string]interface{}

	if err := json.Unmarshal(b, &dat); err != nil {
		return err
	}
	date_string := "\"" + dat["published_at"].(string) + "\""
	date_json := []byte(date_string)
	var date time.Time

	if err := json.Unmarshal(date_json, &date); err != nil {
		return err
	}

	*v = Version{dat["number"].(string), date}
	return nil
}

func ingestData() {

	// Open the folder where the effective POMs are located
	f, err := os.Open("./effective_pom_folder/")
	if err != nil {
		fmt.Println(err)
		return
	}

	files, err := f.Readdirnames(0)
	if err != nil {
		fmt.Println(err)
		return
	}
	var result []VersionDependencies
	helper_map := make(map[string]int)

	r := strings.NewReplacer(
		":", "/",
		".", "/",
	)
	// RegEx for scraping the timetable
	reg, _ := regexp.Compile("(?P<version>(?:0|[1-9]\\d*)?\\.?(?:0|[1-9]\\d*)\\.?(?:0|[1-9]\\d*)?(?:-|_?(?:(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+(?:[0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?)\\/\\s+(?P<timestamp>\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}).*")
	id := 0
	// Iterate through the list of effective POMs
	for _, v := range files {
		resultTime := make(map[string]string)
		id += 1
		fmt.Println(id)

		namePomFile := "./new_pom/" + v
		//Open the effective POM file
		xmlFile, err := os.Open(namePomFile)
		// if we os.Open returns an error then handle it
		if err != nil {
			fmt.Println(err)
		}
		// defer the closing of our xmlFile so that we can parse it later on
		defer xmlFile.Close()
		// Read the effective POM file
		rawDataAddr, _ := ioutil.ReadAll(xmlFile)

		// Create a variable where we can store the metadata from the xml file (effective POM)
		var parsed gopom.Project
		// Unmarshal the data into the variable we just created
		xml.Unmarshal(rawDataAddr, &parsed)

		// Store the metadata into variables
		nameGroup := parsed.GroupID
		nameArtifact := parsed.ArtifactID
		name := nameGroup + ":" + nameArtifact
		number := parsed.Version
		deps := parsed.Dependencies

		// Get the dependencies of the package
		allDependencies := make([]MyDependency, 0, len(deps))
		for _, k := range deps {
			info := MyDependency{k.GroupID, k.ArtifactID, k.Version}

			// If there are exclusions mentioned in the xml file, do not add them to the list of dependencies
			exclusions := k.Exclusions
			exclusionName := ""
			if len(exclusions) > 0 {
				for _, v := range exclusions {
					exclusionName = v.GroupID + ":" + v.ArtifactID
					if exclusionName == name {
						continue
					} else {
						allDependencies = append(allDependencies, info)
					}
				}
			} else {
				allDependencies = append(allDependencies, info)
			}

		}

		// Scrape the timestamps from the Maven Central repository
		timeNameGroup := r.Replace(nameGroup)
		timeNameFormat := timeNameGroup + "/" + nameArtifact
		c := colly.NewCollector(
			colly.AllowedDomains("repo1.maven.org", "https://repo1.maven.org/"),
		)
		c.OnHTML("pre#contents", func(e *colly.HTMLElement) {
			var version_part string
			var time_part string
			s := e.Text
			match := reg.FindAllStringSubmatch(s, -1)
			for k := range match {
				for i, name := range reg.SubexpNames() {
					if name == "version" {
						version_part = match[k][i]
					}
					if name == "timestamp" {
						time_part = match[k][i]
					}
					if version_part != "" && time_part != "" {
						resultTime[version_part] = time_part
						version_part = ""
						time_part = ""
					}
				}
			}
		})
		currentUrl := fmt.Sprintf("https://repo1.maven.org/maven2/%s", timeNameFormat)
		c.Visit(currentUrl)
		// Format the timestamp to RFC 3339 standard
		timestamp := resultTime[number] + ":00Z"
		dt, _ := time.Parse("2006-01-02 15:04:05", timestamp)
		dtstr2 := dt.Format("2006-01-02T15:04:05Z")

		// Create the dependency map
		dep_map := make(map[string]string)
		for _, k := range allDependencies {
			name_package := k.GroupId + ":" + k.Artifact
			dep_map[name_package] = k.RequiredVersion
		}

		// Fill in the structs with the data we got above
		var output_ver = OutputVersion{dtstr2, dep_map}
		ver_map := make(map[string]OutputVersion)

		if _, ok := helper_map[name]; ok {
			result[helper_map[name]].Versions[number] = output_ver
		} else {
			helper_map[name] = len(result)
			ver_map[number] = output_ver
			versionDeps := VersionDependencies{name, ver_map}
			result = append(result, versionDeps)
		}

	}
	// Write to JSON file the final result
	jsonFile, err := json.MarshalIndent(result, "", "    ")
	_ = ioutil.WriteFile("10kPackages.json", jsonFile, 0644)

}

func main() {
	ingestData()
}
