import _ from 'lodash';
import { appEvents, coreModule } from 'app/core/core';

export class AlertNotificationEditCtrl {
  theForm: any;
  navModel: any;
  testSeverity = 'critical';
  notifiers: any;
  notifierTemplateId: string;
  alarmGroups: any;
  alarmURL: string;
  isNew: boolean;
  model: any;
  defaults: any = {
    type: 'email',
    settings: {
      httpMethod: 'POST',
      autoResolve: true,
      uploadImage: true,
    },
    isDefault: false,
  };

  /** @ngInject */
  constructor(private $routeParams, private backendSrv, private $location, private $templateCache, navModelSrv, private $sce) {
    this.navModel = navModelSrv.getNav('alerting', 'channels', 0);
    this.isNew = !this.$routeParams.id;

    this.backendSrv
      .get(`/api/alert-notifiers`)
      .then(notifiers => {
        this.notifiers = notifiers;

        // add option templates
        for (let notifier of this.notifiers) {
          this.$templateCache.put(this.getNotifierTemplateId(notifier.type), notifier.optionsTemplate);
        }

        if (!this.$routeParams.id) {
          this.navModel.breadcrumbs.push({ text: 'New channel' });
          this.navModel.node = { text: 'New channel' };
          return _.defaults(this.model, this.defaults);
        }

        return this.backendSrv.get(`/api/alert-notifications/${this.$routeParams.id}`).then(result => {
          this.navModel.breadcrumbs.push({ text: result.name });
          this.navModel.node = { text: result.name };
          result.settings = _.defaults(result.settings, this.defaults.settings);
          return result;
        });
      })
      .then(model => {
        this.model = model;
        this.notifierTemplateId = this.getNotifierTemplateId(this.model.type);
        this.alarmURL = this.$sce.trustAsResourceUrl(`http://apm.mobike.io/api/apm-argus-dispatcher/api/alarmGroup/allAlarmGroup`);
        this.alarmGroups = this.getAlarmGroups();
      });
  }

  save() {
    if (!this.theForm.$valid) {
      return;
    }

    if (this.model.id) {
      this.backendSrv
        .put(`/api/alert-notifications/${this.model.id}`, this.model)
        .then(res => {
          this.model = res;
          appEvents.emit('alert-success', ['Notification updated', '']);
        })
        .catch(err => {
          if (err.data && err.data.error) {
            appEvents.emit('alert-error', [err.data.error]);
          }
        });
    } else {
      this.backendSrv
        .post(`/api/alert-notifications`, this.model)
        .then(res => {
          appEvents.emit('alert-success', ['Notification created', '']);
          this.$location.path('alerting/notifications');
        })
        .catch(err => {
          if (err.data && err.data.error) {
            appEvents.emit('alert-error', [err.data.error]);
          }
        });
    }
  }

  getAlarmGroups() {
    this.backendSrv.requestAlarmGroup(this.alarmURL).then(data => {
      let result = data.data;
      result.unshift('---请选择---');
      this.alarmGroups = result;
    });
    return this.alarmGroups;
  }

  getNotifierTemplateId(type) {
    return `notifier-options-${type}`;
  }

  typeChanged() {
    this.model.settings = _.defaults({}, this.defaults.settings);
    this.notifierTemplateId = this.getNotifierTemplateId(this.model.type);
  }

  testNotification() {
    if (!this.theForm.$valid) {
      return;
    }

    var payload = {
      name: this.model.name,
      type: this.model.type,
      settings: this.model.settings,
    };

    this.backendSrv.post(`/api/alert-notifications/test`, payload).then(res => {
      appEvents.emit('alert-success', ['Test notification sent', '']);
    });
  }
}

coreModule.controller('AlertNotificationEditCtrl', AlertNotificationEditCtrl);
