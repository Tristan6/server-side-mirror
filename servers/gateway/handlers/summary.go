package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

//PreviewImage represents a preview image for a page
type PreviewImage struct {
	URL       string `json:"url,omitempty"`
	SecureURL string `json:"secureURL,omitempty"`
	Type      string `json:"type,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Alt       string `json:"alt,omitempty"`
}

//PageSummary represents summary properties for a web page
type PageSummary struct {
	Type        string          `json:"type,omitempty"`
	URL         string          `json:"url,omitempty"`
	Title       string          `json:"title,omitempty"`
	SiteName    string          `json:"siteName,omitempty"`
	Description string          `json:"description,omitempty"`
	Author      string          `json:"author,omitempty"`
	Keywords    []string        `json:"keywords,omitempty"`
	Icon        *PreviewImage   `json:"icon,omitempty"`
	Images      []*PreviewImage `json:"images,omitempty"`
}

//SummaryHandler handles requests for the page summary API.
//This API expects one query string parameter named `url`,
//which should contain a URL to a web page. It responds with
//a JSON-encoded PageSummary struct containing the page summary
//meta-data.
// -------------------------------------------------------------------- REQUIRED FUNCTION
func SummaryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	keys, ok := r.URL.Query()["url"]
	// not considering length of string here
	fmt.Println(r)
	if !ok {
		// case when there are no url parameters present in the requested url
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Oops something looks fishy :("))
	} else {
		// case when there is a url param in the request, and we now process it
		requestURL := keys[0]
		// get the html stream of the url
		resp, err := fetchHTML(requestURL)
		// get summary of the html stream
		pageSummary, err := extractSummary(requestURL, resp)
		fmt.Println("***0*")
		fmt.Println(pageSummary)

		//close the response stream
		defer resp.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("problem fetching data"))
		}
		encodedStruct, err := json.Marshal(pageSummary)
		if err != nil {
			// handle error in json encoding
			fmt.Println(err)
			return
		}
		fmt.Println("Final json: ")
		fmt.Println(encodedStruct)
		w.Write([]byte(encodedStruct))

		fmt.Println(err)
	}

}

//fetchHTML fetches `pageURL` and returns the body stream or an error.
//Errors are returned if the response status code is an error (>=400),
//or if the content type indicates the URL is not an HTML page.
// -------------------------------------------------------------------- REQUIRED FUNCTION
func fetchHTML(pageURL string) (io.ReadCloser, error) {
	// if the url length exists, fetch it
	if len(pageURL) >= 1 {
		// get the url response
		resp, err := http.Get(pageURL)
		// handle error from http get
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Println("Bad status code returned from url fetch")
			return nil, errors.New("Bad status code returned from url fetch")
		}
		ctype := resp.Header.Get("Content-Type")
		if !strings.HasPrefix(ctype, "text/html") {
			fmt.Println("Bad content type")
			return nil, errors.New("Bad content types")
		}
		// reach here when everything looks ok, and we respond with the body of http response
		return resp.Body, nil
	}
	return nil, http.ErrContentLength
}

// -------------------------------------------------------------------- REQUIRED FUNCTION
//extractSummary tokenizes the `htmlStream` and populates a PageSummary
//struct with the page's summary meta-data.
func extractSummary(pageURL string, htmlStream io.ReadCloser) (*PageSummary, error) {
	/*TODO: tokenize the `htmlStream` and extract the page summary meta-data
	according to the assignment description.

	To test your implementation of this function, run the TestExtractSummary
	test in summary_test.go. You can do that directly in Visual Studio Code,
	or at the command line by running:
		go test -run TestExtractSummary

	Helpful Links:
	https://drstearns.github.io/tutorials/tokenizing/
	http://ogp.me/
	https://developers.facebook.com/docs/reference/opengraph/
	https://golang.org/pkg/net/url/#URL.ResolveReference
	*/
	mapOfTags := map[string]string{
		"og:title":       "",
		"og:type":        "",
		"og:url":         "",
		"og:site_name":   "",
		"og:description": "",
		"og:image":       "",
		"author":         "",
		"keywords":       "",
		"icon":           "",
	}

	/*  This function goes through the html stream only once,
	and pulls out all the necessary information by returning
	a map of data (meta tags extracted from the head tag) and an array of images */
	resultMap, resultImages := extractRequiredTokens(mapOfTags, &htmlStream)
	// Result images seem to be coming in empty because the tokenizer is not correctly being passed
	// Between functions.

	// do postprocessing of strings here
	ogtype := resultMap["og:type"]
	url := resultMap["og:url"]
	title := resultMap["og:title"]
	siteName := resultMap["og:site_name"]
	description := resultMap["og:description"]
	author := resultMap["og:author"]
	keywords := resultMap["keywords"]

	keywordsArray, nilKeywords := generateKeywordsArray(keywords)

	icons := resultMap["icon"]

	iconStruct := generateIconsPreviewImage(icons)

	nilPreviewImage := false
	// If there were no values created for the website preview image (aka icon) mark a flag
	if reflect.DeepEqual(&iconStruct, &PreviewImage{"", "", "", 0, 0, ""}) {
		nilPreviewImage = true
	}

	resultImagesStruct := generateResultImagesStruct(resultImages)

	finalPageSummary := &PageSummary{
		ogtype, url, title, siteName, description, author, keywordsArray, &iconStruct, resultImagesStruct,
	}

	if nilPreviewImage && nilKeywords {
		finalPageSummary = &PageSummary{
			ogtype, url, title, siteName, description, author, nil, nil, resultImagesStruct,
		}
	} else if nilKeywords {
		finalPageSummary = &PageSummary{
			ogtype, url, title, siteName, description, author, nil, &iconStruct, resultImagesStruct,
		}
	} else if nilPreviewImage {
		finalPageSummary = &PageSummary{
			ogtype, url, title, siteName, description, author, keywordsArray, nil, resultImagesStruct,
		}
	}
	return finalPageSummary, nil
}

// generateKeywordsArray
func generateKeywordsArray(keywords string) ([]string, bool) {
	var keywordsArray []string
	nilKeywords := false
	if len(keywords) == 0 {
		nilKeywords = true
	} else {
		keywordsArray = strings.Split(keywords, ",")
		for i := range keywordsArray {
			keywordsArray[i] = strings.TrimSpace(keywordsArray[i])
		}
	}
	return keywordsArray, nilKeywords
}

// generateIconsPreviewImage
// Need to understand how commas and 3 stars are being added into the resultMap attributes
func generateIconsPreviewImage(icons string) PreviewImage {
	iconsArray := strings.Split(icons, ",")
	var iconStruct PreviewImage

	for _, attr := range iconsArray {
		attr := strings.Split(attr, ">>>")
		// grabs the first item in the array (the )
		switch attr[0] {
		case "href":
			iconStruct.URL = attr[1]
		case "sizes":
			hW := strings.Split(attr[1], "x")
			h, err := strconv.Atoi(hW[0])
			w, err := strconv.Atoi(hW[1])

			if err == nil {
				iconStruct.Height = h
				iconStruct.Width = w
			}
		case "type":
			iconStruct.Type = attr[1]
		}
	}
	return iconStruct
}

// generateResultImagesStruct combines multiple images to create an array of PreviewImage's (essentially an icon data type)
func generateResultImagesStruct(resultImages []string) []*PreviewImage {
	var resultImagesStruct []*PreviewImage

	// This parsing seems like overkill to me
	// Also, the case statements stop after one is used.
	for _, attr := range resultImages {
		var tempImagesStruct PreviewImage
		allLinks := strings.Split(attr, ",")

		for _, b := range allLinks {
			allSubs := strings.Split(b, ">>>")

			switch allSubs[0] {
			case "url":
				tempImagesStruct.URL = allSubs[1]
			case "og:image:width":
				w, err := strconv.Atoi(allSubs[1])
				if err == nil {
					tempImagesStruct.Width = w
				}
			case "og:image:height":
				h, err := strconv.Atoi(allSubs[1])
				if err == nil {
					tempImagesStruct.Height = h
				}
			case "og:image:type":
				tempImagesStruct.Type = allSubs[1]
			case "og:image:secure_url":
				tempImagesStruct.SecureURL = allSubs[1]
			case "og:image:alt":
				tempImagesStruct.Alt = allSubs[1]

			}
		}
		resultImagesStruct = append(resultImagesStruct, &tempImagesStruct)
	}
	return resultImagesStruct
}

/*  This function goes through the html stream only once,
and pulls out all the necessary information by returning
a map of data and an array of images */
func extractRequiredTokens(mapOfTags map[string]string, htmlStream *io.ReadCloser) (map[string]string, []string) {
	tokenizer := html.NewTokenizer(*htmlStream)
	var PreviewImages = []string{}
	var parsedImageData string

	// If og:image:url add one then append to the end of PreviewImages otherwise append to the current image index
	for {
		// grab next token
		tokenType := tokenizer.Next()
		//if it's an error token, we either reached
		//the end of the file, or the HTML was malformed
		if tokenType == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				//end of the file
				break
			}
			log.Fatalf("error tokenizing HTML: %v", tokenizer.Err())
		}

		token := tokenizer.Token()
		// check if this has not reached the end of head tag
		if tokenType == html.EndTagToken && "head" == token.Data {
			break
		}

		// if its a start token
		if tokenType == html.StartTagToken {
			// a link tag
			if token.Data == "link" {
				mapOfTags = processIcons(mapOfTags, token)
			}
			// a meta tag
			if token.Data == "meta" {
				mapOfTags, PreviewImages, parsedImageData = processMetaTags(mapOfTags, token, PreviewImages, parsedImageData)
			}
		}

	}
	return mapOfTags, PreviewImages
}

// processLinkTags
func processIcons(mapOfTags map[string]string, token html.Token) map[string]string {
	iconExistsFlag := false
	// The following variable stands for Open Graph Property (since we will be capturing a lot of these)
	var ogProp string
	// This detects if there is an icon
	for _, attr := range token.Attr {
		_, exists := mapOfTags[attr.Val]
		if attr.Key == "rel" && exists {
			ogProp = attr.Val
			iconExistsFlag = true
			break
		}
	}
	// we have a link with rel=icon & we want to capture the other properties
	if iconExistsFlag {
		thingsToFetch := []string{"href", "type", "sizes"}
		var finalStringForIcon string
		// fmt.Println("token.Attr is: ", token.Attr)

		// for each attribute of the link
		for _, attr := range token.Attr {
			// check if the attribute is one of our required attributes to fetch
			// pattern is key-value,key-value....
			if contains(thingsToFetch, attr.Key) {
				// add the attribute to the final string
				finalStringForIcon += attr.Key + ">>>" + attr.Val + ","
			}
		}
		// fmt.Println("processLinkTags	", finalStringForIcon)
		mapOfTags[ogProp] = finalStringForIcon
	}
	return mapOfTags
}

// processMetaTags
func processMetaTags(mapOfTags map[string]string,
	token html.Token,
	PreviewImages []string,
	parsedImageData string) (map[string]string, []string, string) {

	isProperty := false
	metaNameExists := false
	var ogProp string

	for _, attr := range token.Attr {
		_, exists := mapOfTags[attr.Val]

		if attr.Key == "property" {
			ogProp = attr.Val
			isProperty = true
			break
		}

		if attr.Key == "name" && exists {
			ogProp = attr.Val
			metaNameExists = true
			break
		}
	}
	// if it is an og:image do this
	if isProperty {
		if strings.HasPrefix(ogProp, "og:image") {
			// This is not able to grab anything after the og:image because the htmlStream gets passed empty
			// And therefore it must process each image property in separate calls
			PreviewImages = processImageElements(ogProp, token, PreviewImages)
		} else {
			mapOfTags = processContent(mapOfTags, token, ogProp)
		}
	}
	if metaNameExists {
		mapOfTags = processContent(mapOfTags, token, ogProp)
	}
	return mapOfTags, PreviewImages, parsedImageData
}

func processImageElements(ogProp string, token html.Token, PreviewImages []string) []string {

	ImageElements := [6]string{
		"og:image",
		"og:image:secure_url",
		"og:image:type",
		"og:image:width",
		"og:image:height",
		"og:image:alt",
	}
	isImgURL := ogProp == ImageElements[0]
	var parsedImageData string

	// This needs to be able to process all image attributes
	for _, attr := range token.Attr {
		exists := contains(ImageElements[0:6], ogProp)

		if attr.Key == "content" && exists {
			if isImgURL {
				parsedImageData = "url>>>" + attr.Val + ","
				PreviewImages = append(PreviewImages, parsedImageData)
			} else {
				PreviewImages[len(PreviewImages)-1] += ogProp + ">>>" + attr.Val + ","
			}
		}
	}
	return PreviewImages
}

// processContent
func processContent(mapOfTags map[string]string, token html.Token, ogProp string) map[string]string {
	for _, attr := range token.Attr {
		if attr.Key == "content" {
			mapOfTags[ogProp] = attr.Val
		}
	}
	return mapOfTags
}

// helper function to check whether an element is present in a string
func contains(s []string, e string) bool {
	for _, attr := range s {
		if attr == e {
			return true
		}
	}
	return false
}
