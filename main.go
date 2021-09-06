package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Projects struct {
	Projects []string `json:"projects"`
}

type Groups struct {
	ProjectID     string   `json:"project_id"`
	ProjectName   string   `json:"project_name"`
	VersionGroups []string `json:"version_groups"`
	Versions      []string `json:"versions"`
}

type Versions struct {
	ProjectID    string   `json:"project_id"`
	ProjectName  string   `json:"project_name"`
	VersionGroup string   `json:"version_group"`
	Versions     []string `json:"versions"`
}

type Builds struct {
	ProjectID   string `json:"project_id"`
	ProjectName string `json:"project_name"`
	Version     string `json:"version"`
	Builds      []int  `json:"builds"`
}

type Build struct {
	ProjectID   string    `json:"project_id"`
	ProjectName string    `json:"project_name"`
	Version     string    `json:"version"`
	Build       int       `json:"build"`
	Time        time.Time `json:"time"`
	Changes     []struct {
		Commit  string `json:"commit"`
		Summary string `json:"summary"`
		Message string `json:"message"`
	} `json:"changes"`
	Downloads struct {
		Application struct {
			Name   string `json:"name"`
			Sha256 string `json:"sha256"`
		} `json:"application"`
	} `json:"downloads"`
}

type Downloads struct {
	ProjectID   string    `json:"project_id"`
	ProjectName string    `json:"project_name"`
	Version     string    `json:"version"`
	Build       int       `json:"build"`
	Time        time.Time `json:"time"`
	Changes     []struct {
		Commit  string `json:"commit"`
		Summary string `json:"summary"`
		Message string `json:"message"`
	} `json:"changes"`
	Downloads struct {
		Application struct {
			Name   string `json:"name"`
			Sha256 string `json:"sha256"`
		} `json:"application"`
	} `json:"downloads"`
}

func main() {
	var project = flag.String("project", "paper", "Ex. paper, travertine, waterfall")
	var group = flag.String("group", "latest", "Version group X.XX, latest to use latest")
	var version = flag.String("version", "latest", "Subversion X.XX.X, latest to use latest version")
	var build = flag.String("build", "latest", "Build of version, latest to use latest build")
	var output = flag.String("output", "source", "Output file name, source to use source name")
	var verbose = flag.Bool("verbose", false, "Verbose output")
	flag.Parse()

	if *verbose {
		fmt.Printf("Project: %v\n", *project)
		fmt.Printf("Group: %v\n", *group)
		fmt.Printf("Version: %v\n", *version)
		fmt.Printf("Build: %v\n", *build)
		fmt.Printf("Output: %v\n", *output)
	}

	if *group == "latest" {
		groups, err := GetGroups(*project)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		*group = groups[len(groups)-1]
	}

	if *version == "latest" {
		versions, err := GetVersions(*project, *group)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		*version = versions[len(versions)-1]
	}

	if *build == "latest" {
		builds, err := GetBuilds(*project, *version)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		*build = builds[len(builds)-1]
	}

	name, err := GetName(*project, *version, *build)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *output == "source" {
		*output = name
	}

	remote := fmt.Sprintf("https://papermc.io/api/v2/projects/%s/versions/%s/builds/%s/downloads/%s", *project, *version, *build, name)
	DownloadFile(*output, remote)
}

func GetProjects() ([]string, error) {
	resp, err := http.Get("https://papermc.io/api/v2/projects")
	if err != nil {
		return []string{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []string{}, err
	}

	var data Projects
	err = json.Unmarshal(body, &data)
	if err != nil {
		return []string{}, err
	}

	return data.Projects, nil
}

func GetGroups(project string) ([]string, error) {
	resp, err := http.Get("https://papermc.io/api/v2/projects/" + project)
	if err != nil {
		return []string{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []string{}, err
	}

	var data Groups
	err = json.Unmarshal(body, &data)
	if err != nil {
		return []string{}, err
	}

	return data.VersionGroups, nil
}

func GetVersions(project string, group string) ([]string, error) {
	resp, err := http.Get("https://papermc.io/api/v2/projects/" + project + "/version_group/" + group)
	if err != nil {
		return []string{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []string{}, err
	}

	var data Versions
	err = json.Unmarshal(body, &data)
	if err != nil {
		return []string{}, err
	}

	return data.Versions, nil
}

func GetBuilds(project string, version string) ([]string, error) {
	resp, err := http.Get("https://papermc.io/api/v2/projects/" + project + "/versions/" + version)
	if err != nil {
		return []string{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []string{}, err
	}

	var data Builds
	err = json.Unmarshal(body, &data)
	if err != nil {
		return []string{}, err
	}

	var builds []string
	for _, build := range data.Builds {
		builds = append(builds, strconv.Itoa(build))
	}

	return builds, nil
}

func GetName(project string, version string, build string) (string, error) {
	resp, err := http.Get("https://papermc.io/api/v2/projects/" + project + "/versions/" + version + "/builds/" + build)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var data Downloads
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", err
	}

	return data.Downloads.Application.Name, nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
