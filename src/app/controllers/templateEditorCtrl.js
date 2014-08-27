define([
  'angular',
  'lodash',
],
function (angular, _) {
  'use strict';

  var module = angular.module('grafana.controllers');

  module.controller('TemplateEditorCtrl', function($scope, datasourceSrv) {

    var replacementDefaults = {
      type: 'metric query',
      datasource: null,
      refresh_on_load: false,
      name: '',
      options: [],
    };

    $scope.init = function() {
      $scope.editor = { index: 0 };
      $scope.datasources = datasourceSrv.getMetricSources();
      $scope.currentDatasource = _.findWhere($scope.datasources, { default: true });
      $scope.templateParameters = $scope.filter.templateParameters;
      $scope.reset();

      $scope.$watch('editor.index', function(newVal) {
        if (newVal !== 2) { $scope.reset(); }
      });
    };

    $scope.add = function() {
      $scope.current.datasource = $scope.currentDatasource.name;
      $scope.templateParameters.push($scope.current);
      $scope.reset();
    };

    $scope.edit = function(param) {
      $scope.current = param;
      $scope.currentIsNew = false;
      $scope.currentDatasource = _.findWhere($scope.datasources, { name: param.datasource });

      if (!$scope.currentDatasource) {
        $scope.currentDatasource = $scope.datasources[0];
      }

      $scope.editor.index = 2;
    };

    $scope.reset = function() {
      $scope.currentIsNew = true;
      $scope.current = angular.copy(replacementDefaults);
      $scope.editor.index = 0;
    };

    $scope.removeTemplateParam = function(templateParam) {
      var index = _.indexOf($scope.templateParameters, templateParam);
      $scope.templateParameters.splice(index, 1);
    };

  });

});
