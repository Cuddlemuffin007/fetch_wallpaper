
package main

import (
    "flag"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "regexp"
    "runtime"

    "../util"
    "../web"
)

const BaseUrl string = "https://www.reddit.com/"
var ImgUrlRE = regexp.MustCompile(`i\.imgur\.com/(?P<name>\w+)\.(?P<ext>jpg|png)$`)


func main() {
    targetPtr := flag.String("target", "r/earthporn", "Subreddit from which to fetch a wallpaper image.")
    outfileNamePtr := flag.String("name", "wallpaper", "Name to give to image file when downloaded.")
    outputDirPtr := flag.String(
        "output",
        "~/Downloads",
        fmt.Sprintf(
            "%s\n%s",
            "Directory to which to download requested image.",
            "If overriding, the directory specified should be absolute or relative to $HOME, e.g., ~/path/to/",
        ),
    )
    flag.Parse()

    match, _ := regexp.MatchString(`^r/\w+$`, *targetPtr)
    if !match {
        fmt.Fprintf(os.Stderr, "%s does not appear to be a valid subreddit. Did you lead with r/?\n", *targetPtr)
        os.Exit(1)
    }

    // configure and get a pointer to the web client
    client := web.Client()
    // initialize the response struct
    res := web.RedditResponse{}
    // format url to fetch JSON data from
    url := fmt.Sprintf("%s%s.json", BaseUrl, *targetPtr)
    // make the request, returning an error if unsuccessful, otherwise read response body into the res struct
    if err := web.FetchJSONResponse(url, client, &res); err != nil {
        util.HandleError(err)
    }

    // filter posts to feature only those hosted by imgur
    var targetImageUrls []string
    for _, child := range res.Data.Children {
        if ImgUrlRE.MatchString(child.Data.Url) {
            targetImageUrls = append(targetImageUrls, child.Data.Url)
        }
    }

    // download to the download directry the first image for which there is a successful request (skip returning > 400)
    var imageUrl string
    var imageRes *http.Response = nil
    for _, url := range targetImageUrls {
        // attempt to get image, if error then proceed to next image, if response status is ok break from the loop
        req, err := web.CreateRequest("GET", url)
        if err != nil {
            util.HandleError(err)
        }
        res, err := client.Do(req)
        if err != nil {
            continue
        } else if res.StatusCode < 400 {
            imageRes = res
            imageUrl = url
            defer imageRes.Body.Close()
            break
        }
    }

    if imageRes != nil {
        // extract file name and extension from url to use when writing to file. (Use file name only when not specified)
        match := ImgUrlRE.FindStringSubmatch(imageUrl)
        result := make(map[string]string)
        for i, name := range ImgUrlRE.SubexpNames() {
            if i != 0 && name != "" {
                result[name] = match[i]
            }
        }

        // determine output directory and filename
        if *outfileNamePtr == "" {
            /*
                Override filename with the value from the URL if not provided in flags. For now there is a default
                value being provided, so this is moot, but leaving it for now because I'm considering generating
                unique filenames instead of using the ID from the url in the future.
            */
            *outfileNamePtr = result["name"]
        }
        outputDir, _ := util.ExpandPath(*outputDirPtr)
        fullFilePath := filepath.Join(outputDir, fmt.Sprintf("%s.%s", *outfileNamePtr, result["ext"]))
        outfile, err := os.Create(fullFilePath)
        if err != nil {
            panic(err)
        }
        defer outfile.Close()

        // write file to disk
        _, err = io.Copy(outfile, imageRes.Body)
        if err != nil {
            panic(err)
        }

        // set downloaded file as current wallpaper
        switch runtime.GOOS {
        case "windows":
            // TODO
            fmt.Println("Image downloaded, but setting as wallpaper is not implemented for Windows systems.")
        case "darwin":
            err := util.SetBackgroundMacOS(fullFilePath)
            if err != nil {
                fmt.Fprintf(os.Stderr, "An error occured while attempting to set %s as wallpaper.\n", fullFilePath)
                os.Exit(6)
            }
        default:
            // TODO
            fmt.Println("Image downloaded, but setting as wallpaper is not implemented for Linux systems.")
        }
    } else {
        fmt.Fprintf(os.Stdout, "Could not find an image. Try again later.\n")
    }
}
