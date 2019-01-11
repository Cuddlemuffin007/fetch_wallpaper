
package web


import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "time"
)


// Response from reddit containing the main data object with children slice representing the posts.
type RedditResponse struct {
    Data struct {
        Children []Child `json:"children"`
    } `json:"data"`
}

// Inner child object containing a data struct for the post which contains the url which is all we really care about now.
type Child struct {
    Data struct {
        Url string
    } `json:"data"`
}


// returns a pointer to a new http client with overrides to a few default values via custom Transport struct
func Client() *http.Client {
    tr := &http.Transport{
        IdleConnTimeout: 30 * time.Second,
        TLSHandshakeTimeout: 10 * time.Second,
        DisableCompression: true,
    }
    return &http.Client{Transport: tr}
}


func createRequest(method string, url string) (*http.Request, error) {
    req, err := http.NewRequest(method, url, nil)
    if err != nil {
        return req, &RequestError{
            fmt.Sprintf("%s\nThis error occurred while generating a request for %s\n", err, url),
            2,
        }
    }
    req.Header.Set("User-Agent", "Golang_Wallpaper_Bot/1.0")

    return req, nil
}


func CreateRequest(method string, url string) (*http.Request, error) {
    return createRequest(method, url)
}


func FetchJSONResponse(url string, c *http.Client, res *RedditResponse) error {
    // configure request with custom User-Agent to guard against 429 Too Many Requests response from server
    req, err := createRequest("GET", url)
    if err != nil {
        return err
    }
    // fetch data from server
    rawRes, err := c.Do(req)
    if err != nil {
        return &RequestError{
            fmt.Sprintf("%s\nThis error occurred while attempting to fetch a response from %s\n", err, url),
            3,
        }
    } else if rawRes.StatusCode >= 400 {
        return &RequestError{
            fmt.Sprintf("Request to %s failed with status %d\n", url, rawRes.StatusCode),
            3,
        }
    }
    defer rawRes.Body.Close()

    // read body into memory
    body, err := ioutil.ReadAll(rawRes.Body)
    if err != nil {
        return &RequestError{
            fmt.Sprintf("%s\nThis error occurred while attempting to read the response body from %s\n", err, url),
            4,
        }
    }
    // read raw response body into reddit response struct
    if err := json.Unmarshal(body, res); err != nil {
        return &RequestError{
            fmt.Sprintf("%s\nThis error occurred while attempting to parse json response data from %s\n", err, url),
            5,
        }
    }
    return nil
}
