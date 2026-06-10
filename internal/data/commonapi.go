package data

import (
	"charm.land/log/v2"
	graphql "github.com/cli/shurcooL-graphql"
)

type VersionResponse struct {
	Repository struct {
		LatestRelease struct {
			TagName string
		}
	} `graphql:"repository(owner: $owner, name: $name)"`
}

func FetchLatestVersion() (VersionResponse, error) {
	var queryResult VersionResponse
	client, err := getGraphQLClient()
	if err != nil {
		return VersionResponse{}, err
	}

	variables := map[string]any{
		"owner": graphql.String("dlvhdr"),
		"name":  graphql.String("dehub"),
	}

	log.Debug("Fetching latest version")
	err = client.Query("LatestVersion", &queryResult, variables)
	if err != nil {
		return VersionResponse{}, err
	}
	log.Info("Successfully fetched latest version", "version",
		queryResult.Repository.LatestRelease.TagName)

	return queryResult, nil
}
