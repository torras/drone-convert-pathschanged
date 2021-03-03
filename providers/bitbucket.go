package providers

import (
	"context"
	"net/http"

	"github.com/drone/drone-go/drone"
	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/driver/bitbucket"
	"github.com/drone/go-scm/scm/transport/oauth2"
	"github.com/prometheus/client_golang/prometheus"
)

func GetBitbucketFilesChanged(repo drone.Repo, build drone.Build, clientId string, clientSecret string) ([]string, error) {
	newctx := context.Background()
	client := bitbucket.NewDefault()
	client.Client = &http.Client{
		Transport: &oauth2.Transport{
			Source: &oauth2.Refresher{
				ClientID:     clientId,
				ClientSecret: clientSecret,
				Endpoint:     "https://bitbucket.org/site/oauth2/access_token",
				Source:       oauth2.ContextTokenSource(),
			},
		},
	}

	var changes []*scm.Change
	var result *scm.Response
	var err error

	if build.Before == "" || build.Before == scm.EmptyCommit {
		changes, result, err = client.Git.ListChanges(newctx, repo.Slug, build.After, scm.ListOptions{})
		if err != nil {
			return nil, err
		}
	} else {
		changes, result, err = client.Git.CompareChanges(newctx, repo.Slug, build.Before, build.After, scm.ListOptions{})
		if err != nil {
			return nil, err
		}
	}

	BitbucketApiCount.Set(float64(result.Rate.Remaining))

	var files []string
	for _, c := range changes {
		files = append(files, c.Path)
	}

	return files, nil
}

var (
	BitbucketApiCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "bitbucket_api_calls_remaining",
			Help: "Total number of github api calls per hour remaining",
		})
)

func init() {
	prometheus.MustRegister(BitbucketApiCount)
}
