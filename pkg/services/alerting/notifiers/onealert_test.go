package notifiers

import (
	"testing"

	"github.com/grafana/grafana/pkg/components/simplejson"
	m "github.com/grafana/grafana/pkg/models"
	. "github.com/smartystreets/goconvey/convey"
  "fmt"
  "github.com/grafana/grafana/pkg/services/alerting"
)

func TestOneAlertNotifier(t *testing.T) {
	Convey("OneAlert notifier tests", t, func() {

		Convey("Parsing alert notification from settings", func() {
			Convey("empty settings should return error", func() {
				json := `{ }`

				settingsJSON, _ := simplejson.NewJson([]byte(json))
				model := &m.AlertNotification{
					Name:     "OneAlert",
					Type:     "onealert",
					Settings: settingsJSON,
				}

				_, err := NewOneAlertNotifier(model)
				So(err, ShouldNotBeNil)
			})

			Convey("from settings", func() {
				json := `
				{
          "url": "http://google.com"
				}`

				settingsJSON, _ := simplejson.NewJson([]byte(json))
				model := &m.AlertNotification{
					Name:     "OneAlert",
					Type:     "onealert",
					Settings: settingsJSON,
				}

				not, err := NewOneAlertNotifier(model)
				oneAlertNotifier := not.(*OneAlertNotifier)

        //eval := alerting.EvalContext{}
        //oneAlertNotifier.Notify(&eval)

				So(err, ShouldBeNil)
				So(oneAlertNotifier.Name, ShouldEqual, "OneAlert")
				So(oneAlertNotifier.Type, ShouldEqual, "onealert")

				fmt.Println(alerting.EvalContext{})
			})
		})
	})
}
