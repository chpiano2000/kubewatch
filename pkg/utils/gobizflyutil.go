package utils

import (
	"context"
	"github.com/bizflycloud/gobizfly"
	"log"
)

type BizflyAuth struct {
	Host          string
	AppCredId     string
	AppCredSecret string
	Region        string
	BasicAuth     string
}

func GetApiClient(config *BizflyAuth) (*gobizfly.Client, context.Context) {
	host := config.Host
	authMethod := "application_credential"
	appCredId := config.AppCredId
	appCredSecret := config.AppCredSecret
	region := config.Region
	basicAuth := config.BasicAuth

	// nolint:staticcheck
	client, err := gobizfly.NewClient(gobizfly.WithAPIUrl(host), gobizfly.WithRegionName(region), gobizfly.WithBasicAuth(basicAuth))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), 0)
	defer cancelFunc()

	tcr := &gobizfly.TokenCreateRequest{
		AuthMethod:    authMethod,
		AppCredID:     appCredId,
		AppCredSecret: appCredSecret,
	}
	tok, err := client.Token.Create(ctx, tcr)
	if err != nil {
		log.Fatal(err)
	}

	client.SetKeystoneToken(tok)

	return client, ctx
}
