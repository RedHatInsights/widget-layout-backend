package models_test

import (
	"encoding/json"
	"testing"

	"github.com/RedHatInsights/widget-layout-backend/pkg/models"
)

func TestWidgetUnmarshaling(t *testing.T) {
	testLayoutString := `[{"name":"landing-landingPage","displayName":"LandingPage","templateConfig":{"sm":[{"w":1,"h":4,"maxH":10,"minH":1,"cx":0,"cy":0,"i":"rhel#rhel"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":0,"cy":1,"i":"openshift#openshift"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":0,"cy":2,"i":"ansible#ansible"},{"w":1,"h":5,"maxH":10,"minH":1,"cx":0,"cy":3,"i":"recentlyVisited#recentlyVisited"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":0,"cy":4,"i":"favoriteServices#favoriteServices"},{"w":1,"h":6,"maxH":10,"minH":1,"cx":0,"cy":5,"i":"subscriptions#subscriptions"},{"w":1,"h":13,"maxH":13,"minH":1,"cx":0,"cy":6,"i":"exploreCapabilities#exploreCapabilities"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":0,"cy":7,"i":"openshiftAi#openshiftAi"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":0,"cy":8,"i":"imageBuilder#imageBuilder"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":0,"cy":9,"i":"acs#acs"}],"md":[{"w":1,"h":4,"maxH":10,"minH":1,"cx":0,"cy":0,"i":"rhel#rhel"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":0,"cy":1,"i":"openshift#openshift"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":0,"cy":2,"i":"ansible#ansible"},{"w":1,"h":3,"maxH":10,"minH":1,"cx":1,"cy":0,"i":"recentlyVisited#recentlyVisited"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":1,"cy":1,"i":"favoriteServices#favoriteServices"},{"w":1,"h":5,"maxH":10,"minH":1,"cx":1,"cy":2,"i":"subscriptions#subscriptions"},{"w":2,"h":6,"maxH":10,"minH":1,"cx":0,"cy":3,"i":"exploreCapabilities#exploreCapabilities"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":0,"cy":4,"i":"imageBuilder#imageBuilder"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":1,"cy":4,"i":"openshiftAi#openshiftAi"},{"w":2,"h":3,"maxH":10,"minH":1,"cx":1,"cy":3,"i":"acs#acs"}],"lg":[{"w":1,"h":4,"maxH":10,"minH":1,"cx":0,"cy":0,"i":"rhel#rhel"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":1,"cy":0,"i":"openshift#openshift"},{"w":2,"h":4,"maxH":10,"minH":1,"cx":0,"cy":1,"i":"ansible#ansible"},{"w":2,"h":6,"maxH":10,"minH":1,"cx":0,"cy":4,"i":"exploreCapabilities#exploreCapabilities"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":2,"cy":0,"i":"recentlyVisited#recentlyVisited"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":2,"cy":2,"i":"favoriteServices#favoriteServices"},{"w":1,"h":6,"maxH":10,"minH":1,"cx":2,"cy":3,"i":"subscriptions#subscriptions"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":0,"cy":4,"i":"openshiftAi#openshiftAi"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":1,"cy":4,"i":"imageBuilder#imageBuilder"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":2,"cy":4,"i":"acs#acs"}],"xl":[{"w":1,"h":4,"maxH":10,"minH":1,"cx":0,"cy":0,"i":"rhel#rhel"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":1,"cy":0,"i":"openshift#openshift"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":2,"cy":0,"i":"ansible#ansible"},{"w":3,"h":6,"maxH":10,"minH":1,"cx":0,"cy":2,"i":"exploreCapabilities#exploreCapabilities"},{"w":1,"h":3,"maxH":10,"minH":1,"cx":3,"cy":0,"i":"recentlyVisited#recentlyVisited"},{"w":1,"h":5,"maxH":10,"minH":1,"cx":3,"cy":1,"i":"favoriteServices#favoriteServices"},{"w":1,"h":6,"maxH":10,"minH":1,"cx":3,"cy":2,"i":"subscriptions#subscriptions"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":0,"cy":3,"i":"openshiftAi#openshiftAi"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":1,"cy":3,"i":"imageBuilder#imageBuilder"},{"w":1,"h":4,"maxH":10,"minH":1,"cx":2,"cy":3,"i":"acs#acs"}]},"frontendRef":"landing"}]`
	t.Run("should unmarshal widget items correctly", func(t *testing.T) {
		var bdt []models.BaseWidgetDashboardTemplate
		err := json.Unmarshal([]byte(testLayoutString), &bdt)
		if err != nil {
			t.Fatalf("Failed to unmarshal widget items: %v", err)
		}

		items := bdt[0].TemplateConfig.Sm.Data()

		if len(items) == 0 {
			t.Fatal("No widget items were unmarshaled")
		}

		for _, item := range items {
			if item.X == nil || item.Y == nil {
				t.Error("WidgetItem must have either x/y or cx/cy attributes")
			}
		}
	})
}
