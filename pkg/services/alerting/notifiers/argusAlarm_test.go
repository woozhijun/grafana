package notifiers

import (
	"context"
	"testing"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	m "github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/services/alerting"
)

func TestDemo(t *testing.T)  {

	Convey("ArgusAlarm notifier tests", t, func() {

		Convey(" httpGet alarmGroup return..", func() {

			s := httpGetAlarmGroups()
			fmt.Println("================" + s)
		})

		Convey("Get settings ", func() {
			json := `{ "url": "http://172.17.8.127:8080/alarm/grafana/send","httpMethod":"POST","alarmGroup":"apm" }`

			settingsJSON, _ := simplejson.NewJson([]byte(json))
			model := &m.AlertNotification{
				Name:     "ArgusAlarm",
				Type:     "argusAlarm",
				Settings: settingsJSON,
			}

			alarmNot, error := NewArgusAlarmNotifier(model)
			if error != nil {
				fmt.Println(error)
			}

			var rule alerting.Rule
			rule.Id = 1
			rule.NoDataState = "alerting"
			rule.ExecutionErrorState = "alerting"

			eval := alerting.NewEvalContext(context.Background(), &rule)

			argustNotifier := alarmNot.(*ArgusAlarmNotifier)
			argustNotifier.Notify(eval)

			fmt.Println(">>>." + argustNotifier.alarmGroup)

			So(argustNotifier.Name, ShouldEqual, "ArgusAlarm")
		})

	})
}
