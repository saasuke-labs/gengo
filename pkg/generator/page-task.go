package generator

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
)

// TODO - For now we assume all data is HTTP API, with GET method, etc
type ExternalDataTask struct {
	Key string
	Url string
}

type PageData struct {
	Title        string
	Tags         []string
	HTML         template.HTML
	Metadata     map[string]string
	Section      string
	Sections     []string
	ExternalData map[string]interface{}
}

type PageTask struct {
	Title             string
	InputFile         string
	OutputFile        string
	Url               string
	Template          string
	LayoutTemplate    string
	Metadata          map[string]string
	Tags              []string
	Section           string
	Sections          []string
	ExternalDataTasks []ExternalDataTask
}

func fetchData(url string) (interface{}, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request: ", err)
		return nil, err
	}

	//req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error executing request: ", err)
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {

		fmt.Println("Status: ", resp.StatusCode)
		return nil, err
	}

	// Assume data is HTML for now
	// TODO - Parse the data and return it as string

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("Error reading response body: ", err)
		return nil, err
	}

	// TODO - Return the string, expose the safeHTML function in the templates
	//  and make the client decide if the HTML is safe to render or not.
	fmt.Println("Data: ", string(data))
	return template.HTML(string(data)), nil
}

func (t PageTask) Execute() error {

	html := generateMarkdownPage(t.InputFile)

	externalData := make(map[string]interface{})

	for _, task := range t.ExternalDataTasks {
		fmt.Println("Fetching data for key: ", task.Key, " from URL: ", task.Url)
		// Fetch Data from the external API
		data, err := fetchData(task.Url)
		// TODO - See how to handle errors
		if err != nil {
			externalData[task.Key] = nil
		} else {
			externalData[task.Key] = data
		}
	}

	fmt.Println("External Data: ", externalData)

	html = applyTemplate(t.Template, PageData{
		// See how to get the title from the HTML
		Title:        "",
		Tags:         t.Tags,
		Metadata:     t.Metadata,
		HTML:         html,
		ExternalData: externalData,
	})
	html = applyTemplate(t.LayoutTemplate, PageData{
		Title:    t.Title,
		Tags:     t.Tags,
		Metadata: t.Metadata,
		HTML:     html,
		Section:  t.Section,
		Sections: t.Sections,
	})

	savePage(html, t.OutputFile)
	return nil
}

func (t PageTask) Name() string {
	return t.OutputFile
}
