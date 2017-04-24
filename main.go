package main

import (
    "fmt"
    "github.com/google/go-github/github"
    "io"
    "log"
    "math"
    "os"
    "time"
    "context"
)

const remaingThreshold = 1

func main() {

    t := &github.UnauthenticatedRateLimitedTransport{
        ClientID:     "b75562c8336156e4da9a",
        ClientSecret: "177e7474bb712d4bf7213dc99e086d3dffce9e11",
    }
    client := github.NewClient(t.Client())
    ctx := context.Background()

    fmt.Println("Repos that contain coala and python code.")

    // create a file to be used for geocoder
    filename := "locations.txt"

    f, err := os.Create(filename)
    if err != nil {
        fmt.Println(err)
        //this is a fatal error, quit
        return
    }
    defer f.Close()

    // slice the queries into batches to get around the API limit of 1000

    queries := []string{
        "\"2010-09-30 .. 2017-09-30\""}

    for _, q := range queries {

        query := fmt.Sprintf("coala created:" + q)

        page := 1
        maxPage := math.MaxInt32

        opts := &github.SearchOptions{
            Sort:  "updated",
            Order: "desc",
            ListOptions: github.ListOptions{
                PerPage: 100,
            },
        }

        for ; page <= maxPage; page++ {
            opts.Page = page
            result, response, err := client.Search.Repositories(ctx, query, opts)
            wait(response)

            if err != nil {
                log.Fatal("FindRepos:", err)
                break
            }

            maxPage = response.LastPage

            msg := fmt.Sprintf("page: %v/%v, size: %v, total: %v",
                page, maxPage, len(result.Repositories), *result.Total)
            log.Println(msg)

            for _, repo := range result.Repositories {

                repoName := *repo.FullName
                username := *repo.Owner.Login
                createdAt := repo.CreatedAt.String()

                fmt.Println("repo: ", repoName)
                fmt.Println("owner: ", username)
                fmt.Println("created at: ", createdAt)

                user, response, err := client.Users.Get(ctx, username)
                wait(response)

                if err != nil {
                    fmt.Println("error getting userinfo for:", username, err)
                    continue
                }

                userLocation := ""
                if user.Location == nil {
                    userLocation = "not found"
                } else {
                    userLocation = *user.Location
                }

                n, err := io.WriteString(f, "\""+username+"\",\""+userLocation+"\",\""+repoName+"\",\""+createdAt+"\"\n")
                if err != nil {
                    fmt.Println(n, err)
                }

                time.Sleep(time.Millisecond * 500)
            }
        }
    }

}

func wait(response *github.Response) {
    if response != nil && response.Remaining <= remaingThreshold {
        gap := time.Duration(response.Reset.Local().Unix() - time.Now().Unix())
        sleep := gap * time.Second
        if sleep < 0 {
            sleep = -sleep
        }

        time.Sleep(sleep)
    }
}
