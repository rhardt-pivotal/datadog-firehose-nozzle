package util

import (
	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/cloudfoundry/gosteno"

	"fmt"
	"sync/atomic"
)

var Durations = []int{1,60,300,3600,86400}

type AppDataLookup struct {

	cfClient			*cfclient.Client
	log					*gosteno.Logger
	appMetadataMap		map[string]*AppMetadata

}

type AppMetadata struct {
	OrgName		string
	SpaceName	string
	AppName		string
	requestCount int64
}

func (amd *AppMetadata) Incr() {
	atomic.AddInt64(&(amd.requestCount), 1)
}

func (amd *AppMetadata) GetCount() int64 {
	return atomic.LoadInt64(&(amd.requestCount))
}

const _unknown  = "<unknown>"
const not_an_app  = "00000000-0000-0000-0000-000000000000"

var  unknown = &AppMetadata{
	OrgName:_unknown,
	SpaceName:_unknown,
	AppName:_unknown,
}

var nonAppApiHttp = &AppMetadata{
	OrgName:"NA",
	SpaceName:"NA",
	AppName:"non-app-api-http-call",
}


func NewAppDataLookup(
	apiURL string,
	clientId string,
	clientSecret string,
    skipSslValidation bool,
	log *gosteno.Logger,
) *AppDataLookup {

	c := &cfclient.Config{
		ApiAddress: apiURL,
		ClientID: clientId,
		ClientSecret: clientSecret,
		SkipSslValidation: skipSslValidation,
	}

	cfc,err := cfclient.NewClient(c)

	if (err != nil) {
		panic(err.Error())
	}

	log.Info("Fetching apps")
	allApps, err := cfc.ListApps()
	log.Info(fmt.Sprintf("retrieved %d apps", len(allApps)))

	if (err != nil) {
		panic(err.Error())
	}

	log.Info("Fetching orgs")
	allOrgs, err := cfc.ListOrgs()
	log.Info(fmt.Sprintf("retrieved %d orgs", len(allApps)))

	if (err != nil) {
		panic(err.Error())
	}

	orgNames := make(map[string]string)

	for _, org := range allOrgs {
		orgNames[org.Guid] = org.Name
	}


	log.Info("Building internal map")
	appMetaDataMap := make(map[string]*AppMetadata)
	for _, app := range allApps {
		amd := &AppMetadata{
			OrgName: orgNames[app.SpaceData.Entity.OrganizationGuid],
			SpaceName: app.SpaceData.Entity.Name,
			AppName: app.Name,
		}
		appMetaDataMap[app.Guid] = amd
	}
	appMetaDataMap[not_an_app] = nonAppApiHttp
	log.Info("Done")


	return &AppDataLookup{
		cfClient: cfc,
		log: 	  log,
		appMetadataMap: appMetaDataMap,
	}
}



//func (c *RequestRegistry) RegisterMessage(envelope *events.Envelope) {
//	appId := util.Stringify(envelope.HttpStartStop.ApplicationId)
//	//c.log.Info("ENRICH MESSAGE: "+appId+"-----"+envelope.String())
//	if (appId != not_an_app) {
//		appStr := c.appMap[appId]
//		if appStr != "" {
//			if c.ratecountersmap[appStr] == nil {
//				c.ratecountersmap[appStr] = util.NewRateCounters(Durations)
//			}
//			c.ratecountersmap[appStr].Incr()
//		}
//		c.appsThisCycle[appId] = true
//	}
//}


func (c *AppDataLookup) LookupAppMetadata(appId string) *AppMetadata {
	if c.appMetadataMap[appId] == nil {
		c.updateMetadata(appId)
	}
	return c.appMetadataMap[appId]
}

func (c *AppDataLookup) Incr(appId string) {
	c.LookupAppMetadata(appId).Incr()
}

func (c *AppDataLookup) GetValue(appId string) int64 {
	return c.LookupAppMetadata(appId).GetCount()
}

func (c *AppDataLookup) updateMetadata(appId string) {
	app, err := c.cfClient.GetAppByGuid(appId)
	updated := &AppMetadata{}
	if err != nil {
		updated = unknown
	} else {
		org, err := c.cfClient.GetOrgByGuid(app.SpaceData.Entity.OrganizationGuid)
		if err != nil {
			updated.OrgName = unknown.OrgName
		} else {
			updated.OrgName = org.Name
		}
		updated.AppName = app.Name
		updated.SpaceName = app.SpaceData.Entity.Name
	}
	c.appMetadataMap[appId] = updated
}


