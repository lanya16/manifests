package tests_test

import (
  "sigs.k8s.io/kustomize/v3/k8sdeps/kunstruct"
  "sigs.k8s.io/kustomize/v3/k8sdeps/transformer"
  "sigs.k8s.io/kustomize/v3/pkg/fs"
  "sigs.k8s.io/kustomize/v3/pkg/loader"
  "sigs.k8s.io/kustomize/v3/pkg/plugins"
  "sigs.k8s.io/kustomize/v3/pkg/resmap"
  "sigs.k8s.io/kustomize/v3/pkg/resource"
  "sigs.k8s.io/kustomize/v3/pkg/target"
  "sigs.k8s.io/kustomize/v3/pkg/validators"
  "testing"
)

func writeIstioInstallBase(th *KustTestHarness) {
  th.writeF("/manifests/istio/istio-install/base/istio-noauth.yaml", `
apiVersion: v1
kind: Namespace
metadata:
  name: istio-system
  labels:
    istio-injection: disabled
---
# Source: istio/charts/kiali/templates/demosecret.yaml

apiVersion: v1
kind: Secret
metadata:
  name: kiali
  labels:
    app: kiali
    chart: kiali
    heritage: Tiller
    release: istio
type: Opaque
data:
  username: YWRtaW4=   # admin
  passphrase: YWRtaW4= # admin

---
# Source: istio/charts/galley/templates/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-galley-configuration
  labels:
    app: galley
    chart: galley
    heritage: Tiller
    release: istio
    istio: galley
data:
  validatingwebhookconfiguration.yaml: |-    
    apiVersion: admissionregistration.k8s.io/v1beta1
    kind: ValidatingWebhookConfiguration
    metadata:
      name: istio-galley
      namespace: istio-system
      labels:
        app: galley
        chart: galley
        heritage: Tiller
        release: istio
        istio: galley
    webhooks:
      - name: pilot.validation.istio.io
        clientConfig:
          service:
            name: istio-galley
            namespace: istio-system
            path: "/admitpilot"
          caBundle: ""
        rules:
          - operations:
            - CREATE
            - UPDATE
            apiGroups:
            - config.istio.io
            apiVersions:
            - v1alpha2
            resources:
            - httpapispecs
            - httpapispecbindings
            - quotaspecs
            - quotaspecbindings
          - operations:
            - CREATE
            - UPDATE
            apiGroups:
            - rbac.istio.io
            apiVersions:
            - "*"
            resources:
            - "*"
          - operations:
            - CREATE
            - UPDATE
            apiGroups:
            - authentication.istio.io
            apiVersions:
            - "*"
            resources:
            - "*"
          - operations:
            - CREATE
            - UPDATE
            apiGroups:
            - networking.istio.io
            apiVersions:
            - "*"
            resources:
            - destinationrules
            - envoyfilters
            - gateways
            - serviceentries
            - sidecars
            - virtualservices
        failurePolicy: Fail
      - name: mixer.validation.istio.io
        clientConfig:
          service:
            name: istio-galley
            namespace: istio-system
            path: "/admitmixer"
          caBundle: ""
        rules:
          - operations:
            - CREATE
            - UPDATE
            apiGroups:
            - config.istio.io
            apiVersions:
            - v1alpha2
            resources:
            - rules
            - attributemanifests
            - circonuses
            - deniers
            - fluentds
            - kubernetesenvs
            - listcheckers
            - memquotas
            - noops
            - opas
            - prometheuses
            - rbacs
            - solarwindses
            - stackdrivers
            - cloudwatches
            - dogstatsds
            - statsds
            - stdios
            - apikeys
            - authorizations
            - checknothings
            # - kuberneteses
            - listentries
            - logentries
            - metrics
            - quotas
            - reportnothings
            - tracespans
        failurePolicy: Fail
---
# Source: istio/charts/grafana/templates/configmap-custom-resources.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-grafana-custom-resources
  labels:
    app: grafana
    chart: grafana
    heritage: Tiller
    release: istio
    istio: grafana
data:
  custom-resources.yaml: |-    
    apiVersion: authentication.istio.io/v1alpha1
    kind: Policy
    metadata:
      name: grafana-ports-mtls-disabled
      namespace: istio-system
      labels:
        app: grafana
        chart: grafana
        heritage: Tiller
        release: istio
    spec:
      targets:
      - name: grafana
        ports:
        - number: 3000
  run.sh: |-    
    #!/bin/sh
    
    set -x
    
    if [ "$#" -ne "1" ]; then
        echo "first argument should be path to custom resource yaml"
        exit 1
    fi
    
    pathToResourceYAML=${1}
    
    kubectl get validatingwebhookconfiguration istio-galley 2>/dev/null
    if [ "$?" -eq 0 ]; then
        echo "istio-galley validatingwebhookconfiguration found - waiting for istio-galley deployment to be ready"
        while true; do
            kubectl -n istio-system get deployment istio-galley 2>/dev/null
            if [ "$?" -eq 0 ]; then
                break
            fi
            sleep 1
        done
        kubectl -n istio-system rollout status deployment istio-galley
        if [ "$?" -ne 0 ]; then
            echo "istio-galley deployment rollout status check failed"
            exit 1
        fi
        echo "istio-galley deployment ready for configuration validation"
    fi
    sleep 5
    kubectl apply -f ${pathToResourceYAML}
    

---
# Source: istio/charts/grafana/templates/configmap-dashboards.yaml

apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-grafana-configuration-dashboards-galley-dashboard
  labels:
    app: grafana
    chart: grafana
    heritage: Tiller
    release: istio
    istio: grafana
data:
  galley-dashboard.json: '{
  "__inputs": [
    {
      "name": "DS_PROMETHEUS",
      "label": "Prometheus",
      "description": "",
      "type": "datasource",
      "pluginId": "prometheus",
      "pluginName": "Prometheus"
    }
  ],
  "annotations": {
    "list": [
      {
        "builtIn": 1,
        "datasource": "-- Grafana --",
        "enable": true,
        "hide": true,
        "iconColor": "rgba(0, 211, 255, 1)",
        "name": "Annotations & Alerts",
        "type": "dashboard"
      }
    ]
  },
  "editable": false,
  "gnetId": null,
  "graphTooltip": 0,
  "links": [],
  "panels": [
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 5,
        "w": 24,
        "x": 0,
        "y": 0
      },
      "id": 46,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(istio_build{component=\"galley\"}) by (tag)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ tag }}",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Galley Versions",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "collapsed": false,
      "gridPos": {
        "h": 1,
        "w": 24,
        "x": 0,
        "y": 5
      },
      "id": 40,
      "panels": [],
      "title": "Resource Usage",
      "type": "row"
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 8,
        "w": 6,
        "x": 0,
        "y": 6
      },
      "id": 36,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "process_virtual_memory_bytes{job=\"galley\"}",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "Virtual Memory",
          "refId": "A"
        },
        {
          "expr": "process_resident_memory_bytes{job=\"galley\"}",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "Resident Memory",
          "refId": "B"
        },
        {
          "expr": "go_memstats_heap_sys_bytes{job=\"galley\"}",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "heap sys",
          "refId": "C"
        },
        {
          "expr": "go_memstats_heap_alloc_bytes{job=\"galley\"}",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "heap alloc",
          "refId": "D"
        },
        {
          "expr": "go_memstats_alloc_bytes{job=\"galley\"}",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "Alloc",
          "refId": "F"
        },
        {
          "expr": "go_memstats_heap_inuse_bytes{job=\"galley\"}",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "Heap in-use",
          "refId": "G"
        },
        {
          "expr": "go_memstats_stack_inuse_bytes{job=\"galley\"}",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "Stack in-use",
          "refId": "H"
        },
        {
          "expr": "sum(container_memory_usage_bytes{container_name=~\"galley\", pod_name=~\"istio-galley-.*\"})",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Total (kis)",
          "refId": "E"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Memory",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 8,
        "w": 6,
        "x": 6,
        "y": 6
      },
      "id": 38,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(rate(container_cpu_usage_seconds_total{container_name=~\"galley\", pod_name=~\"istio-galley-.*\"}[1m]))",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "Total (k8s)",
          "refId": "A"
        },
        {
          "expr": "sum(rate(container_cpu_usage_seconds_total{container_name=~\"galley\", pod_name=~\"istio-galley-.*\"}[1m])) by (container_name)",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "{{ container_name }} (k8s)",
          "refId": "B"
        },
        {
          "expr": "irate(process_cpu_seconds_total{job=\"galley\"}[1m])",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "galley (self-reported)",
          "refId": "C"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "CPU",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 8,
        "w": 6,
        "x": 12,
        "y": 6
      },
      "id": 42,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "process_open_fds{job=\"galley\"}",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "Open FDs (galley)",
          "refId": "A"
        },
        {
          "expr": "container_fs_usage_bytes{container_name=~\"galley\", pod_name=~\"istio-galley-.*\"}",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "{{ container_name }} ",
          "refId": "B"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Disk",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 8,
        "w": 6,
        "x": 18,
        "y": 6
      },
      "id": 44,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "go_goroutines{job=\"galley\"}",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "goroutines_total",
          "refId": "A"
        },
        {
          "expr": "galley_mcp_source_clients_total",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "clients_total",
          "refId": "B"
        },
        {
          "expr": "go_goroutines{job=\"galley\"}/galley_mcp_source_clients_total",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "avg_goroutines_per_client",
          "refId": "C"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Goroutines",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "collapsed": false,
      "gridPos": {
        "h": 1,
        "w": 24,
        "x": 0,
        "y": 14
      },
      "id": 10,
      "panels": [],
      "title": "Runtime",
      "type": "row"
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 0,
        "y": 15
      },
      "id": 2,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(rate(galley_runtime_strategy_on_change_total[1m])) * 60",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Strategy Change Events",
          "refId": "A"
        },
        {
          "expr": "sum(rate(galley_runtime_processor_events_processed_total[1m])) * 60",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Processed Events",
          "refId": "B"
        },
        {
          "expr": "sum(rate(galley_runtime_processor_snapshots_published_total[1m])) * 60",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Snapshot Published",
          "refId": "C"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Event Rates",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": "Events/min",
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": "",
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 8,
        "y": 15
      },
      "id": 4,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(rate(galley_runtime_strategy_timer_max_time_reached_total[1m])) * 60",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Max Time Reached",
          "refId": "A"
        },
        {
          "expr": "sum(rate(galley_runtime_strategy_timer_quiesce_reached_total[1m])) * 60",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Quiesce Reached",
          "refId": "B"
        },
        {
          "expr": "sum(rate(galley_runtime_strategy_timer_resets_total[1m])) * 60",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Timer Resets",
          "refId": "C"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Timer Rates",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": "Events/min",
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 16,
        "y": 15
      },
      "id": 8,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 3,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": true,
      "steppedLine": false,
      "targets": [
        {
          "expr": "histogram_quantile(0.50, sum by (le) (galley_runtime_processor_snapshot_events_total_bucket))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "P50",
          "refId": "A"
        },
        {
          "expr": "histogram_quantile(0.90, sum by (le) (galley_runtime_processor_snapshot_events_total_bucket))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "P90",
          "refId": "B"
        },
        {
          "expr": "histogram_quantile(0.95, sum by (le) (galley_runtime_processor_snapshot_events_total_bucket))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "P95",
          "refId": "C"
        },
        {
          "expr": "histogram_quantile(0.99, sum by (le) (galley_runtime_processor_snapshot_events_total_bucket))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "P99",
          "refId": "D"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Events Per Snapshot",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 8,
        "y": 21
      },
      "id": 6,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum by (typeURL) (galley_runtime_state_type_instances_total)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ typeURL }}",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "State Type Instances",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": "Count",
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "collapsed": false,
      "gridPos": {
        "h": 1,
        "w": 24,
        "x": 0,
        "y": 27
      },
      "id": 34,
      "panels": [],
      "title": "Validation",
      "type": "row"
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 0,
        "y": 28
      },
      "id": 28,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "galley_validation_cert_key_updates{job=\"galley\"}",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Key Updates",
          "refId": "A"
        },
        {
          "expr": "galley_validation_cert_key_update_errors{job=\"galley\"}",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Key Update Errors: {{ error }}",
          "refId": "B"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Validation Webhook Certificate",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 8,
        "y": 28
      },
      "id": 30,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(galley_validation_passed{job=\"galley\"}) by (group, version, resource)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Passed: {{ group }}/{{ version }}/{{resource}}",
          "refId": "A"
        },
        {
          "expr": "sum(galley_validation_failed{job=\"galley\"}) by (group, version, resource, reason)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Failed: {{ group }}/{{ version }}/{{resource}} ({{ reason}})",
          "refId": "B"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Resource Validation",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 16,
        "y": 28
      },
      "id": 32,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(galley_validation_http_error{job=\"galley\"}) by (status)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ status }}",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Validation HTTP Errors",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "collapsed": false,
      "gridPos": {
        "h": 1,
        "w": 24,
        "x": 0,
        "y": 34
      },
      "id": 12,
      "panels": [],
      "title": "Kubernetes Source",
      "type": "row"
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 0,
        "y": 35
      },
      "id": 14,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "rate(galley_source_kube_event_success_total[1m]) * 60",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Success",
          "refId": "A"
        },
        {
          "expr": "rate(galley_source_kube_event_error_total[1m]) * 60",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Error",
          "refId": "B"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Source Event Rate",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": "Events/min",
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 8,
        "y": 35
      },
      "id": 16,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "rate(galley_source_kube_dynamic_converter_success_total[1m]) * 60",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{apiVersion=\"{{apiVersion}}\",group=\"{{group}}\",kind=\"{{kind}}\"}",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Kubernetes Object Conversion Successes",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": "Conversions/min",
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 16,
        "y": 35
      },
      "id": 24,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "rate(galley_source_kube_dynamic_converter_failure_total[1m]) * 60",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Error",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Kubernetes Object Conversion Failures",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": "Failures/min",
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "collapsed": false,
      "gridPos": {
        "h": 1,
        "w": 24,
        "x": 0,
        "y": 41
      },
      "id": 18,
      "panels": [],
      "title": "Mesh Configuration Protocol",
      "type": "row"
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 0,
        "y": 42
      },
      "id": 20,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(galley_mcp_source_clients_total)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Clients",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Connected Clients",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 8,
        "y": 42
      },
      "id": 22,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum by(collection)(irate(galley_mcp_source_request_acks_total[1m]) * 60)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Request ACKs",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": "ACKs/min",
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 16,
        "y": 42
      },
      "id": 26,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "rate(galley_mcp_source_request_nacks_total[1m]) * 60",
          "format": "time_series",
          "intervalFactor": 1,
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeRegions": [],
      "timeShift": null,
      "title": "Request NACKs",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": "NACKs/min",
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    }
  ],
  "refresh": "5s",
  "schemaVersion": 16,
  "style": "dark",
  "tags": [],
  "templating": {
    "list": []
  },
  "time": {
    "from": "now-5m",
    "to": "now"
  },
  "timepicker": {
    "refresh_intervals": [
      "5s",
      "10s",
      "30s",
      "1m",
      "5m",
      "15m",
      "30m",
      "1h",
      "2h",
      "1d"
    ],
    "time_options": [
      "5m",
      "15m",
      "1h",
      "6h",
      "12h",
      "24h",
      "2d",
      "7d",
      "30d"
    ]
  },
  "timezone": "",
  "title": "Istio Galley Dashboard",
  "uid": "TSEY6jLmk",
  "version": 1
}
'
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-grafana-configuration-dashboards-istio-mesh-dashboard
  labels:
    app: grafana
    chart: grafana
    heritage: Tiller
    release: istio
    istio: grafana
data:
  istio-mesh-dashboard.json: '{
  "__inputs": [
    {
      "name": "DS_PROMETHEUS",
      "label": "Prometheus",
      "description": "",
      "type": "datasource",
      "pluginId": "prometheus",
      "pluginName": "Prometheus"
    }
  ],
  "__requires": [
    {
      "type": "grafana",
      "id": "grafana",
      "name": "Grafana",
      "version": "5.2.3"
    },
    {
      "type": "panel",
      "id": "graph",
      "name": "Graph",
      "version": "5.0.0"
    },
    {
      "type": "datasource",
      "id": "prometheus",
      "name": "Prometheus",
      "version": "5.0.0"
    },
    {
      "type": "panel",
      "id": "singlestat",
      "name": "Singlestat",
      "version": "5.0.0"
    },
    {
      "type": "panel",
      "id": "table",
      "name": "Table",
      "version": "5.0.0"
    },
    {
      "type": "panel",
      "id": "text",
      "name": "Text",
      "version": "5.0.0"
    }
  ],
  "annotations": {
    "list": [
      {
        "builtIn": 1,
        "datasource": "-- Grafana --",
        "enable": true,
        "hide": true,
        "iconColor": "rgba(0, 211, 255, 1)",
        "name": "Annotations & Alerts",
        "type": "dashboard"
      }
    ]
  },
  "editable": false,
  "gnetId": null,
  "graphTooltip": 0,
  "id": null,
  "links": [],
  "panels": [
    {
      "content": "<div>\n  <div style=\"position: absolute; bottom: 0\">\n    <a href=\"https://istio.io\" target=\"_blank\" style=\"font-size: 30px; text-decoration: none; color: inherit\"><img src=\"https://istio.io/img/istio-logo.svg\" style=\"height: 50px\"> Istio</a>\n  </div>\n  <div style=\"position: absolute; bottom: 0; right: 0; font-size: 15px\">\n    Istio is an <a href=\"https://github.com/istio/istio\" target=\"_blank\">open platform</a> that provides a uniform way to connect,\n    <a href=\"https://istio.io/docs/concepts/traffic-management/overview.html\" target=\"_blank\">manage</a>, and \n    <a href=\"https://istio.io/docs/concepts/network-and-auth/auth.html\" target=\"_blank\">secure</a> microservices.\n    <br>\n    Need help? Join the <a href=\"https://istio.io/community/\" target=\"_blank\">Istio community</a>.\n  </div>\n</div>",
      "gridPos": {
        "h": 3,
        "w": 24,
        "x": 0,
        "y": 0
      },
      "height": "50px",
      "id": 13,
      "links": [],
      "mode": "html",
      "style": {
        "font-size": "18pt"
      },
      "title": "",
      "transparent": true,
      "type": "text"
    },
    {
      "cacheTimeout": null,
      "colorBackground": false,
      "colorValue": false,
      "colors": [
        "rgba(245, 54, 54, 0.9)",
        "rgba(237, 129, 40, 0.89)",
        "rgba(50, 172, 45, 0.97)"
      ],
      "datasource": "Prometheus",
      "format": "ops",
      "gauge": {
        "maxValue": 100,
        "minValue": 0,
        "show": false,
        "thresholdLabels": false,
        "thresholdMarkers": true
      },
      "gridPos": {
        "h": 3,
        "w": 6,
        "x": 0,
        "y": 3
      },
      "id": 20,
      "interval": null,
      "links": [],
      "mappingType": 1,
      "mappingTypes": [
        {
          "name": "value to text",
          "value": 1
        },
        {
          "name": "range to text",
          "value": 2
        }
      ],
      "maxDataPoints": 100,
      "nullPointMode": "connected",
      "nullText": null,
      "postfix": "",
      "postfixFontSize": "50%",
      "prefix": "",
      "prefixFontSize": "50%",
      "rangeMaps": [
        {
          "from": "null",
          "text": "N/A",
          "to": "null"
        }
      ],
      "sparkline": {
        "fillColor": "rgba(31, 118, 189, 0.18)",
        "full": true,
        "lineColor": "rgb(31, 120, 193)",
        "show": true
      },
      "tableColumn": "",
      "targets": [
        {
          "expr": "round(sum(irate(istio_requests_total{reporter=\"destination\"}[1m])), 0.001)",
          "intervalFactor": 1,
          "refId": "A",
          "step": 4
        }
      ],
      "thresholds": "",
      "title": "Global Request Volume",
      "transparent": false,
      "type": "singlestat",
      "valueFontSize": "80%",
      "valueMaps": [
        {
          "op": "=",
          "text": "N/A",
          "value": "null"
        }
      ],
      "valueName": "avg"
    },
    {
      "cacheTimeout": null,
      "colorBackground": false,
      "colorValue": false,
      "colors": [
        "rgba(245, 54, 54, 0.9)",
        "rgba(237, 129, 40, 0.89)",
        "rgba(50, 172, 45, 0.97)"
      ],
      "datasource": "Prometheus",
      "format": "percentunit",
      "gauge": {
        "maxValue": 100,
        "minValue": 80,
        "show": false,
        "thresholdLabels": false,
        "thresholdMarkers": false
      },
      "gridPos": {
        "h": 3,
        "w": 6,
        "x": 6,
        "y": 3
      },
      "id": 21,
      "interval": null,
      "links": [],
      "mappingType": 1,
      "mappingTypes": [
        {
          "name": "value to text",
          "value": 1
        },
        {
          "name": "range to text",
          "value": 2
        }
      ],
      "maxDataPoints": 100,
      "nullPointMode": "connected",
      "nullText": null,
      "postfix": "",
      "postfixFontSize": "50%",
      "prefix": "",
      "prefixFontSize": "50%",
      "rangeMaps": [
        {
          "from": "null",
          "text": "N/A",
          "to": "null"
        }
      ],
      "sparkline": {
        "fillColor": "rgba(31, 118, 189, 0.18)",
        "full": true,
        "lineColor": "rgb(31, 120, 193)",
        "show": true
      },
      "tableColumn": "",
      "targets": [
        {
          "expr": "sum(rate(istio_requests_total{reporter=\"destination\", response_code!~\"5.*\"}[1m])) / sum(rate(istio_requests_total{reporter=\"destination\"}[1m]))",
          "format": "time_series",
          "intervalFactor": 1,
          "refId": "A",
          "step": 4
        }
      ],
      "thresholds": "95, 99, 99.5",
      "title": "Global Success Rate (non-5xx responses)",
      "transparent": false,
      "type": "singlestat",
      "valueFontSize": "80%",
      "valueMaps": [
        {
          "op": "=",
          "text": "N/A",
          "value": "null"
        }
      ],
      "valueName": "avg"
    },
    {
      "cacheTimeout": null,
      "colorBackground": false,
      "colorValue": false,
      "colors": [
        "rgba(245, 54, 54, 0.9)",
        "rgba(237, 129, 40, 0.89)",
        "rgba(50, 172, 45, 0.97)"
      ],
      "datasource": "Prometheus",
      "format": "ops",
      "gauge": {
        "maxValue": 100,
        "minValue": 0,
        "show": false,
        "thresholdLabels": false,
        "thresholdMarkers": true
      },
      "gridPos": {
        "h": 3,
        "w": 6,
        "x": 12,
        "y": 3
      },
      "id": 22,
      "interval": null,
      "links": [],
      "mappingType": 1,
      "mappingTypes": [
        {
          "name": "value to text",
          "value": 1
        },
        {
          "name": "range to text",
          "value": 2
        }
      ],
      "maxDataPoints": 100,
      "nullPointMode": "connected",
      "nullText": null,
      "postfix": "",
      "postfixFontSize": "50%",
      "prefix": "",
      "prefixFontSize": "50%",
      "rangeMaps": [
        {
          "from": "null",
          "text": "N/A",
          "to": "null"
        }
      ],
      "sparkline": {
        "fillColor": "rgba(31, 118, 189, 0.18)",
        "full": true,
        "lineColor": "rgb(31, 120, 193)",
        "show": true
      },
      "tableColumn": "",
      "targets": [
        {
          "expr": "sum(irate(istio_requests_total{reporter=\"destination\", response_code=~\"4.*\"}[1m])) ",
          "format": "time_series",
          "intervalFactor": 1,
          "refId": "A",
          "step": 4
        }
      ],
      "thresholds": "",
      "title": "4xxs",
      "transparent": false,
      "type": "singlestat",
      "valueFontSize": "80%",
      "valueMaps": [
        {
          "op": "=",
          "text": "N/A",
          "value": "null"
        }
      ],
      "valueName": "avg"
    },
    {
      "cacheTimeout": null,
      "colorBackground": false,
      "colorValue": false,
      "colors": [
        "rgba(245, 54, 54, 0.9)",
        "rgba(237, 129, 40, 0.89)",
        "rgba(50, 172, 45, 0.97)"
      ],
      "datasource": "Prometheus",
      "format": "ops",
      "gauge": {
        "maxValue": 100,
        "minValue": 0,
        "show": false,
        "thresholdLabels": false,
        "thresholdMarkers": true
      },
      "gridPos": {
        "h": 3,
        "w": 6,
        "x": 18,
        "y": 3
      },
      "id": 23,
      "interval": null,
      "links": [],
      "mappingType": 1,
      "mappingTypes": [
        {
          "name": "value to text",
          "value": 1
        },
        {
          "name": "range to text",
          "value": 2
        }
      ],
      "maxDataPoints": 100,
      "nullPointMode": "connected",
      "nullText": null,
      "postfix": "",
      "postfixFontSize": "50%",
      "prefix": "",
      "prefixFontSize": "50%",
      "rangeMaps": [
        {
          "from": "null",
          "text": "N/A",
          "to": "null"
        }
      ],
      "sparkline": {
        "fillColor": "rgba(31, 118, 189, 0.18)",
        "full": true,
        "lineColor": "rgb(31, 120, 193)",
        "show": true
      },
      "tableColumn": "",
      "targets": [
        {
          "expr": "sum(irate(istio_requests_total{reporter=\"destination\", response_code=~\"5.*\"}[1m])) ",
          "format": "time_series",
          "intervalFactor": 1,
          "refId": "A",
          "step": 4
        }
      ],
      "thresholds": "",
      "title": "5xxs",
      "transparent": false,
      "type": "singlestat",
      "valueFontSize": "80%",
      "valueMaps": [
        {
          "op": "=",
          "text": "N/A",
          "value": "null"
        }
      ],
      "valueName": "avg"
    },
    {
      "columns": [],
      "datasource": "Prometheus",
      "fontSize": "100%",
      "gridPos": {
        "h": 21,
        "w": 24,
        "x": 0,
        "y": 6
      },
      "hideTimeOverride": false,
      "id": 73,
      "links": [],
      "pageSize": null,
      "repeat": null,
      "repeatDirection": "v",
      "scroll": true,
      "showHeader": true,
      "sort": {
        "col": 4,
        "desc": true
      },
      "styles": [
        {
          "alias": "Workload",
          "colorMode": null,
          "colors": [
            "rgba(245, 54, 54, 0.9)",
            "rgba(237, 129, 40, 0.89)",
            "rgba(50, 172, 45, 0.97)"
          ],
          "dateFormat": "YYYY-MM-DD HH:mm:ss",
          "decimals": 2,
          "link": false,
          "linkTargetBlank": false,
          "linkTooltip": "Workload dashboard",
          "linkUrl": "/dashboard/db/istio-workload-dashboard?var-namespace=$__cell_2&var-workload=$__cell_",
          "pattern": "destination_workload",
          "preserveFormat": false,
          "sanitize": false,
          "thresholds": [],
          "type": "hidden",
          "unit": "short"
        },
        {
          "alias": "",
          "colorMode": null,
          "colors": [
            "rgba(245, 54, 54, 0.9)",
            "rgba(237, 129, 40, 0.89)",
            "rgba(50, 172, 45, 0.97)"
          ],
          "dateFormat": "YYYY-MM-DD HH:mm:ss",
          "decimals": 2,
          "pattern": "Time",
          "thresholds": [],
          "type": "hidden",
          "unit": "short"
        },
        {
          "alias": "Requests",
          "colorMode": null,
          "colors": [
            "rgba(245, 54, 54, 0.9)",
            "rgba(237, 129, 40, 0.89)",
            "rgba(50, 172, 45, 0.97)"
          ],
          "dateFormat": "YYYY-MM-DD HH:mm:ss",
          "decimals": 2,
          "pattern": "Value #A",
          "thresholds": [],
          "type": "number",
          "unit": "ops"
        },
        {
          "alias": "P50 Latency",
          "colorMode": null,
          "colors": [
            "rgba(245, 54, 54, 0.9)",
            "rgba(237, 129, 40, 0.89)",
            "rgba(50, 172, 45, 0.97)"
          ],
          "dateFormat": "YYYY-MM-DD HH:mm:ss",
          "decimals": 2,
          "pattern": "Value #B",
          "thresholds": [],
          "type": "number",
          "unit": "s"
        },
        {
          "alias": "P90 Latency",
          "colorMode": null,
          "colors": [
            "rgba(245, 54, 54, 0.9)",
            "rgba(237, 129, 40, 0.89)",
            "rgba(50, 172, 45, 0.97)"
          ],
          "dateFormat": "YYYY-MM-DD HH:mm:ss",
          "decimals": 2,
          "pattern": "Value #D",
          "thresholds": [],
          "type": "number",
          "unit": "s"
        },
        {
          "alias": "P99 Latency",
          "colorMode": null,
          "colors": [
            "rgba(245, 54, 54, 0.9)",
            "rgba(237, 129, 40, 0.89)",
            "rgba(50, 172, 45, 0.97)"
          ],
          "dateFormat": "YYYY-MM-DD HH:mm:ss",
          "decimals": 2,
          "pattern": "Value #E",
          "thresholds": [],
          "type": "number",
          "unit": "s"
        },
        {
          "alias": "Success Rate",
          "colorMode": "cell",
          "colors": [
            "rgba(245, 54, 54, 0.9)",
            "rgba(237, 129, 40, 0.89)",
            "rgba(50, 172, 45, 0.97)"
          ],
          "dateFormat": "YYYY-MM-DD HH:mm:ss",
          "decimals": 2,
          "pattern": "Value #F",
          "thresholds": [
            ".95",
            " 1.00"
          ],
          "type": "number",
          "unit": "percentunit"
        },
        {
          "alias": "Workload",
          "colorMode": null,
          "colors": [
            "rgba(245, 54, 54, 0.9)",
            "rgba(237, 129, 40, 0.89)",
            "rgba(50, 172, 45, 0.97)"
          ],
          "dateFormat": "YYYY-MM-DD HH:mm:ss",
          "decimals": 2,
          "link": true,
          "linkTooltip": "$__cell dashboard",
          "linkUrl": "/dashboard/db/istio-workload-dashboard?var-workload=$__cell_2&var-namespace=$__cell_3",
          "pattern": "destination_workload_var",
          "thresholds": [],
          "type": "number",
          "unit": "short"
        },
        {
          "alias": "Service",
          "colorMode": null,
          "colors": [
            "rgba(245, 54, 54, 0.9)",
            "rgba(237, 129, 40, 0.89)",
            "rgba(50, 172, 45, 0.97)"
          ],
          "dateFormat": "YYYY-MM-DD HH:mm:ss",
          "decimals": 2,
          "link": true,
          "linkTooltip": "$__cell dashboard",
          "linkUrl": "/dashboard/db/istio-service-dashboard?var-service=$__cell",
          "pattern": "destination_service",
          "thresholds": [],
          "type": "string",
          "unit": "short"
        },
        {
          "alias": "",
          "colorMode": null,
          "colors": [
            "rgba(245, 54, 54, 0.9)",
            "rgba(237, 129, 40, 0.89)",
            "rgba(50, 172, 45, 0.97)"
          ],
          "dateFormat": "YYYY-MM-DD HH:mm:ss",
          "decimals": 2,
          "pattern": "destination_workload_namespace",
          "thresholds": [],
          "type": "hidden",
          "unit": "short"
        }
      ],
      "targets": [
        {
          "expr": "label_join(sum(rate(istio_requests_total{reporter=\"destination\", response_code=\"200\"}[1m])) by (destination_workload, destination_workload_namespace, destination_service), \"destination_workload_var\", \".\", \"destination_workload\", \"destination_workload_namespace\")",
          "format": "table",
          "hide": false,
          "instant": true,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload}}.{{ destination_workload_namespace }}",
          "refId": "A"
        },
        {
          "expr": "label_join(histogram_quantile(0.50, sum(rate(istio_request_duration_seconds_bucket{reporter=\"destination\"}[1m])) by (le, destination_workload, destination_workload_namespace)), \"destination_workload_var\", \".\", \"destination_workload\", \"destination_workload_namespace\")",
          "format": "table",
          "hide": false,
          "instant": true,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload}}.{{ destination_workload_namespace }}",
          "refId": "B"
        },
        {
          "expr": "label_join(histogram_quantile(0.90, sum(rate(istio_request_duration_seconds_bucket{reporter=\"destination\"}[1m])) by (le, destination_workload, destination_workload_namespace)), \"destination_workload_var\", \".\", \"destination_workload\", \"destination_workload_namespace\")",
          "format": "table",
          "hide": false,
          "instant": true,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }}",
          "refId": "D"
        },
        {
          "expr": "label_join(histogram_quantile(0.99, sum(rate(istio_request_duration_seconds_bucket{reporter=\"destination\"}[1m])) by (le, destination_workload, destination_workload_namespace)), \"destination_workload_var\", \".\", \"destination_workload\", \"destination_workload_namespace\")",
          "format": "table",
          "hide": false,
          "instant": true,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }}",
          "refId": "E"
        },
        {
          "expr": "label_join((sum(rate(istio_requests_total{reporter=\"destination\", response_code!~\"5.*\"}[1m])) by (destination_workload, destination_workload_namespace) / sum(rate(istio_requests_total{reporter=\"destination\"}[1m])) by (destination_workload, destination_workload_namespace)), \"destination_workload_var\", \".\", \"destination_workload\", \"destination_workload_namespace\")",
          "format": "table",
          "hide": false,
          "instant": true,
          "interval": "",
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }}",
          "refId": "F"
        }
      ],
      "timeFrom": null,
      "title": "HTTP/GRPC Workloads",
      "transform": "table",
      "transparent": false,
      "type": "table"
    },
    {
      "columns": [],
      "datasource": "Prometheus",
      "fontSize": "100%",
      "gridPos": {
        "h": 18,
        "w": 24,
        "x": 0,
        "y": 27
      },
      "hideTimeOverride": false,
      "id": 109,
      "links": [],
      "pageSize": null,
      "repeatDirection": "v",
      "scroll": true,
      "showHeader": true,
      "sort": {
        "col": 2,
        "desc": true
      },
      "styles": [
        {
          "alias": "Workload",
          "colorMode": null,
          "colors": [
            "rgba(245, 54, 54, 0.9)",
            "rgba(237, 129, 40, 0.89)",
            "rgba(50, 172, 45, 0.97)"
          ],
          "dateFormat": "YYYY-MM-DD HH:mm:ss",
          "decimals": 2,
          "link": false,
          "linkTargetBlank": false,
          "linkTooltip": "$__cell dashboard",
          "linkUrl": "/dashboard/db/istio-tcp-workload-dashboard?var-namespace=$__cell_2&&var-workload=$__cell",
          "pattern": "destination_workload",
          "preserveFormat": false,
          "sanitize": false,
          "thresholds": [],
          "type": "hidden",
          "unit": "short"
        },
        {
          "alias": "Bytes Sent",
          "colorMode": null,
          "colors": [
            "rgba(245, 54, 54, 0.9)",
            "rgba(237, 129, 40, 0.89)",
            "rgba(50, 172, 45, 0.97)"
          ],
          "dateFormat": "YYYY-MM-DD HH:mm:ss",
          "decimals": 2,
          "pattern": "Value #A",
          "thresholds": [
            ""
          ],
          "type": "number",
          "unit": "Bps"
        },
        {
          "alias": "Bytes Received",
          "colorMode": null,
          "colors": [
            "rgba(245, 54, 54, 0.9)",
            "rgba(237, 129, 40, 0.89)",
            "rgba(50, 172, 45, 0.97)"
          ],
          "dateFormat": "YYYY-MM-DD HH:mm:ss",
          "decimals": 2,
          "pattern": "Value #C",
          "thresholds": [],
          "type": "number",
          "unit": "Bps"
        },
        {
          "alias": "",
          "colorMode": null,
          "colors": [
            "rgba(245, 54, 54, 0.9)",
            "rgba(237, 129, 40, 0.89)",
            "rgba(50, 172, 45, 0.97)"
          ],
          "dateFormat": "YYYY-MM-DD HH:mm:ss",
          "decimals": 2,
          "pattern": "Time",
          "thresholds": [],
          "type": "hidden",
          "unit": "short"
        },
        {
          "alias": "Workload",
          "colorMode": null,
          "colors": [
            "rgba(245, 54, 54, 0.9)",
            "rgba(237, 129, 40, 0.89)",
            "rgba(50, 172, 45, 0.97)"
          ],
          "dateFormat": "YYYY-MM-DD HH:mm:ss",
          "decimals": 2,
          "link": true,
          "linkTooltip": "$__cell dashboard",
          "linkUrl": "/dashboard/db/istio-workload-dashboard?var-namespace=$__cell_3&var-workload=$__cell_2",
          "pattern": "destination_workload_var",
          "thresholds": [],
          "type": "string",
          "unit": "short"
        },
        {
          "alias": "",
          "colorMode": null,
          "colors": [
            "rgba(245, 54, 54, 0.9)",
            "rgba(237, 129, 40, 0.89)",
            "rgba(50, 172, 45, 0.97)"
          ],
          "dateFormat": "YYYY-MM-DD HH:mm:ss",
          "decimals": 2,
          "pattern": "destination_workload_namespace",
          "thresholds": [],
          "type": "hidden",
          "unit": "short"
        },
        {
          "alias": "Service",
          "colorMode": null,
          "colors": [
            "rgba(245, 54, 54, 0.9)",
            "rgba(237, 129, 40, 0.89)",
            "rgba(50, 172, 45, 0.97)"
          ],
          "dateFormat": "YYYY-MM-DD HH:mm:ss",
          "decimals": 2,
          "link": true,
          "linkTooltip": "$__cell dashboard",
          "linkUrl": "/dashboard/db/istio-service-dashboard?var-service=$__cell",
          "pattern": "destination_service",
          "thresholds": [],
          "type": "number",
          "unit": "short"
        }
      ],
      "targets": [
        {
          "expr": "label_join(sum(rate(istio_tcp_received_bytes_total{reporter=\"source\"}[1m])) by (destination_workload, destination_workload_namespace, destination_service), \"destination_workload_var\", \".\", \"destination_workload\", \"destination_workload_namespace\")",
          "format": "table",
          "hide": false,
          "instant": true,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}",
          "refId": "C"
        },
        {
          "expr": "label_join(sum(rate(istio_tcp_sent_bytes_total{reporter=\"source\"}[1m])) by (destination_workload, destination_workload_namespace, destination_service), \"destination_workload_var\", \".\", \"destination_workload\", \"destination_workload_namespace\")",
          "format": "table",
          "hide": false,
          "instant": true,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}",
          "refId": "A"
        }
      ],
      "timeFrom": null,
      "title": "TCP Workloads",
      "transform": "table",
      "transparent": false,
      "type": "table"
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 9,
        "w": 24,
        "x": 0,
        "y": 45
      },
      "id": 111,
      "legend": {
        "alignAsTable": false,
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "rightSide": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(istio_build) by (component, tag)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ component }}: {{ tag }}",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Istio Components by Version",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "transparent": false,
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    }
  ],
  "refresh": "5s",
  "schemaVersion": 16,
  "style": "dark",
  "tags": [],
  "templating": {
    "list": []
  },
  "time": {
    "from": "now-5m",
    "to": "now"
  },
  "timepicker": {
    "refresh_intervals": [
      "5s",
      "10s",
      "30s",
      "1m",
      "5m",
      "15m",
      "30m",
      "1h",
      "2h",
      "1d"
    ],
    "time_options": [
      "5m",
      "15m",
      "1h",
      "6h",
      "12h",
      "24h",
      "2d",
      "7d",
      "30d"
    ]
  },
  "timezone": "browser",
  "title": "Istio Mesh Dashboard",
  "version": 4
}
'
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-grafana-configuration-dashboards-istio-performance-dashboard
  labels:
    app: grafana
    chart: grafana
    heritage: Tiller
    release: istio
    istio: grafana
data:
  istio-performance-dashboard.json: '{
  "__inputs": [
    {
      "name": "DS_PROMETHEUS",
      "label": "Prometheus",
      "description": "",
      "type": "datasource",
      "pluginId": "prometheus",
      "pluginName": "Prometheus"
    }
  ],
  "__requires": [
    {
      "type": "grafana",
      "id": "grafana",
      "name": "Grafana",
      "version": "5.2.3"
    },
    {
      "type": "panel",
      "id": "graph",
      "name": "Graph",
      "version": "5.0.0"
    },
    {
      "type": "datasource",
      "id": "prometheus",
      "name": "Prometheus",
      "version": "5.0.0"
    },
    {
      "type": "panel",
      "id": "text",
      "name": "Text",
      "version": "5.0.0"
    }
  ],
  "annotations": {
    "list": [
      {
        "builtIn": 1,
        "datasource": "-- Grafana --",
        "enable": true,
        "hide": true,
        "iconColor": "rgba(0, 211, 255, 1)",
        "name": "Annotations & Alerts",
        "type": "dashboard"
      }
    ]
  },
  "editable": false,
  "gnetId": null,
  "graphTooltip": 0,
  "id": null,
  "links": [],
  "panels": [
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 9,
        "w": 12,
        "x": 0,
        "y": 0
      },
      "id": 2,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "(sum(rate(container_cpu_usage_seconds_total{pod_name=~\"istio-telemetry-.*\",container_name=~\"mixer|istio-proxy\"}[1m]))/ (round(sum(irate(istio_requests_total[1m])), 0.001)/1000))/ (sum(irate(istio_requests_total{source_workload=\"istio-ingressgateway\"}[1m])) >bool 10)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "istio-telemetry",
          "refId": "A"
        },
        {
          "expr": "sum(rate(container_cpu_usage_seconds_total{pod_name=~\"istio-ingressgateway-.*\",container_name=\"istio-proxy\"}[1m])) / (round(sum(irate(istio_requests_total{source_workload=\"istio-ingressgateway\", reporter=\"source\"}[1m])), 0.001)/1000)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "istio-ingressgateway",
          "refId": "B"
        },
        {
          "expr": "(sum(rate(container_cpu_usage_seconds_total{namespace!=\"istio-system\",container_name=\"istio-proxy\"}[1m]))/ (round(sum(irate(istio_requests_total[1m])), 0.001)/1000))/ (sum(irate(istio_requests_total{source_workload=\"istio-ingressgateway\"}[1m])) >bool 10)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "istio-proxy",
          "refId": "C"
        },
        {
          "expr": "(sum(rate(container_cpu_usage_seconds_total{pod_name=~\"istio-policy-.*\",container_name=~\"mixer|istio-proxy\"}[1m]))/ (round(sum(irate(istio_requests_total[1m])), 0.001)/1000)) / (sum(irate(istio_requests_total{source_workload=\"istio-ingressgateway\"}[1m])) >bool 10)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "istio-policy",
          "refId": "D"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "vCPU / 1k rps",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 9,
        "w": 12,
        "x": 12,
        "y": 0
      },
      "id": 6,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(rate(container_cpu_usage_seconds_total{pod_name=~\"istio-telemetry-.*\",container_name=~\"mixer|istio-proxy\"}[1m]))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "istio-telemetry",
          "refId": "A"
        },
        {
          "expr": "sum(rate(container_cpu_usage_seconds_total{pod_name=~\"istio-ingressgateway-.*\",container_name=\"istio-proxy\"}[1m]))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "istio-ingressgateway",
          "refId": "B"
        },
        {
          "expr": "sum(rate(container_cpu_usage_seconds_total{namespace!=\"istio-system\",container_name=\"istio-proxy\"}[1m]))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "istio-proxy",
          "refId": "C"
        },
        {
          "expr": "sum(rate(container_cpu_usage_seconds_total{pod_name=~\"istio-policy-.*\",container_name=~\"mixer|istio-proxy\"}[1m]))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "istio-policy",
          "refId": "D"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "vCPU",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 9,
        "w": 12,
        "x": 0,
        "y": 9
      },
      "id": 4,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "(sum(container_memory_usage_bytes{pod_name=~\"istio-telemetry-.*\"}) / (sum(irate(istio_requests_total[1m])) / 1000)) / (sum(irate(istio_requests_total{source_workload=\"istio-ingressgateway\"}[1m])) >bool 10)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "istio-telemetry / 1k rps",
          "refId": "A"
        },
        {
          "expr": "sum(container_memory_usage_bytes{pod_name=~\"istio-ingressgateway-.*\"}) / count(container_memory_usage_bytes{pod_name=~\"istio-ingressgateway-.*\",container_name!=\"POD\"})",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "per istio-ingressgateway",
          "refId": "C"
        },
        {
          "expr": "sum(container_memory_usage_bytes{namespace!=\"istio-system\",container_name=\"istio-proxy\"}) / count(container_memory_usage_bytes{namespace!=\"istio-system\",container_name=\"istio-proxy\"})",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "per istio-proxy",
          "refId": "B"
        },
        {
          "expr": "(sum(container_memory_usage_bytes{pod_name=~\"istio-policy-.*\"}) / (sum(irate(istio_requests_total[1m])) / 1000))/ (sum(irate(istio_requests_total{source_workload=\"istio-ingressgateway\"}[1m])) >bool 10)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "istio-policy / 1k rps",
          "refId": "D"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Memory",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "decbytes",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 9,
        "w": 12,
        "x": 12,
        "y": 9
      },
      "id": 5,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(irate(istio_response_bytes_sum{destination_workload=\"istio-telemetry\"}[1m])) + sum(irate(istio_request_bytes_sum{destination_workload=\"istio-telemetry\"}[1m]))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "istio-telemetry",
          "refId": "A"
        },
        {
          "expr": "sum(irate(istio_response_bytes_sum{source_workload=\"istio-ingressgateway\", reporter=\"source\"}[1m]))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "istio-ingressgateway",
          "refId": "C"
        },
        {
          "expr": "sum(irate(istio_response_bytes_sum{source_workload_namespace!=\"istio-system\", reporter=\"source\"}[1m])) + sum(irate(istio_response_bytes_sum{destination_workload_namespace!=\"istio-system\", reporter=\"destination\"}[1m])) + sum(irate(istio_request_bytes_sum{source_workload_namespace!=\"istio-system\", reporter=\"source\"}[1m])) + sum(irate(istio_request_bytes_sum{destination_workload_namespace!=\"istio-system\", reporter=\"destination\"}[1m]))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "istio-proxy",
          "refId": "D"
        },
        {
          "expr": "sum(irate(istio_response_bytes_sum{destination_workload=\"istio-policy\"}[1m])) + sum(irate(istio_request_bytes_sum{destination_workload=\"istio-policy\"}[1m]))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "istio-policy",
          "refId": "E"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Bytes transferred / sec",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "bytes",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 9,
        "w": 24,
        "x": 0,
        "y": 18
      },
      "id": 8,
      "legend": {
        "alignAsTable": false,
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "rightSide": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(istio_build) by (component, tag)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ component }}: {{ tag }}",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Istio Components by Version",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "transparent": false,
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "content": "The charts on this dashboard are intended to show Istio main components cost in terms resources utilization under steady load.\n\n- **vCPU/1k rps:** shows vCPU utilization by the main Istio components normalized by 1000 requests/second. When idle or low traffic, this chart will be blank. The curve for istio-proxy refers to the services sidecars only. \n- **vCPU:** vCPU utilization by Istio components, not normalized.\n- **Memory:** memory footprint for the components. Telemetry and policy are normalized by 1k rps, and no data is shown  when there is no traffic. For ingress and istio-proxy, the data is per instance. \n- **Bytes transferred/ sec:** shows the number of bytes flowing through each Istio component.",
      "gridPos": {
        "h": 4,
        "w": 24,
        "x": 0,
        "y": 18
      },
      "id": 11,
      "links": [],
      "mode": "markdown",
      "title": "Istio Performance Dashboard Readme",
      "type": "text"
    }
  ],
  "schemaVersion": 16,
  "style": "dark",
  "tags": [],
  "templating": {
    "list": []
  },
  "time": {
    "from": "now-5m",
    "to": "now"
  },
  "timepicker": {
    "refresh_intervals": [
      "5s",
      "10s",
      "30s",
      "1m",
      "5m",
      "15m",
      "30m",
      "1h",
      "2h",
      "1d"
    ],
    "time_options": [
      "5m",
      "15m",
      "1h",
      "6h",
      "12h",
      "24h",
      "2d",
      "7d",
      "30d"
    ]
  },
  "timezone": "",
  "title": "Istio Performance Dashboard",
  "version": 4
}
'
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-grafana-configuration-dashboards-istio-service-dashboard
  labels:
    app: grafana
    chart: grafana
    heritage: Tiller
    release: istio
    istio: grafana
data:
  istio-service-dashboard.json: '{
  "annotations": {
    "list": [
      {
        "builtIn": 1,
        "datasource": "-- Grafana --",
        "enable": true,
        "hide": true,
        "iconColor": "rgba(0, 211, 255, 1)",
        "name": "Annotations & Alerts",
        "type": "dashboard"
      }
    ]
  },
  "editable": false,
  "gnetId": null,
  "graphTooltip": 0,
  "iteration": 1536442501501,
  "links": [],
  "panels": [
    {
      "content": "<div class=\"dashboard-header text-center\">\n<span>SERVICE: $service</span>\n</div>",
      "gridPos": {
        "h": 3,
        "w": 24,
        "x": 0,
        "y": 0
      },
      "id": 89,
      "links": [],
      "mode": "html",
      "title": "",
      "transparent": true,
      "type": "text"
    },
    {
      "cacheTimeout": null,
      "colorBackground": false,
      "colorValue": false,
      "colors": [
        "rgba(245, 54, 54, 0.9)",
        "rgba(237, 129, 40, 0.89)",
        "rgba(50, 172, 45, 0.97)"
      ],
      "datasource": "Prometheus",
      "format": "ops",
      "gauge": {
        "maxValue": 100,
        "minValue": 0,
        "show": false,
        "thresholdLabels": false,
        "thresholdMarkers": true
      },
      "gridPos": {
        "h": 4,
        "w": 6,
        "x": 0,
        "y": 3
      },
      "id": 12,
      "interval": null,
      "links": [],
      "mappingType": 1,
      "mappingTypes": [
        {
          "name": "value to text",
          "value": 1
        },
        {
          "name": "range to text",
          "value": 2
        }
      ],
      "maxDataPoints": 100,
      "nullPointMode": "connected",
      "nullText": null,
      "postfix": "",
      "postfixFontSize": "50%",
      "prefix": "",
      "prefixFontSize": "50%",
      "rangeMaps": [
        {
          "from": "null",
          "text": "N/A",
          "to": "null"
        }
      ],
      "sparkline": {
        "fillColor": "rgba(31, 118, 189, 0.18)",
        "full": true,
        "lineColor": "rgb(31, 120, 193)",
        "show": true
      },
      "tableColumn": "",
      "targets": [
        {
          "expr": "round(sum(irate(istio_requests_total{reporter=\"source\",destination_service=~\"$service\"}[5m])), 0.001)",
          "format": "time_series",
          "intervalFactor": 1,
          "refId": "A",
          "step": 4
        }
      ],
      "thresholds": "",
      "title": "Client Request Volume",
      "transparent": false,
      "type": "singlestat",
      "valueFontSize": "80%",
      "valueMaps": [
        {
          "op": "=",
          "text": "N/A",
          "value": "null"
        }
      ],
      "valueName": "current"
    },
    {
      "cacheTimeout": null,
      "colorBackground": false,
      "colorValue": false,
      "colors": [
        "rgba(50, 172, 45, 0.97)",
        "rgba(237, 129, 40, 0.89)",
        "rgba(245, 54, 54, 0.9)"
      ],
      "datasource": "Prometheus",
      "decimals": null,
      "format": "percentunit",
      "gauge": {
        "maxValue": 100,
        "minValue": 80,
        "show": false,
        "thresholdLabels": false,
        "thresholdMarkers": false
      },
      "gridPos": {
        "h": 4,
        "w": 6,
        "x": 6,
        "y": 3
      },
      "id": 14,
      "interval": null,
      "links": [],
      "mappingType": 1,
      "mappingTypes": [
        {
          "name": "value to text",
          "value": 1
        },
        {
          "name": "range to text",
          "value": 2
        }
      ],
      "maxDataPoints": 100,
      "nullPointMode": "connected",
      "nullText": null,
      "postfix": "",
      "postfixFontSize": "50%",
      "prefix": "",
      "prefixFontSize": "50%",
      "rangeMaps": [
        {
          "from": "null",
          "text": "N/A",
          "to": "null"
        }
      ],
      "sparkline": {
        "fillColor": "rgba(31, 118, 189, 0.18)",
        "full": true,
        "lineColor": "rgb(31, 120, 193)",
        "show": true
      },
      "tableColumn": "",
      "targets": [
        {
          "expr": "sum(irate(istio_requests_total{reporter=\"source\",destination_service=~\"$service\",response_code!~\"5.*\"}[5m])) / sum(irate(istio_requests_total{reporter=\"source\",destination_service=~\"$service\"}[5m]))",
          "format": "time_series",
          "intervalFactor": 1,
          "refId": "B"
        }
      ],
      "thresholds": "95, 99, 99.5",
      "title": "Client Success Rate (non-5xx responses)",
      "transparent": false,
      "type": "singlestat",
      "valueFontSize": "80%",
      "valueMaps": [
        {
          "op": "=",
          "text": "N/A",
          "value": "null"
        }
      ],
      "valueName": "avg"
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 4,
        "w": 6,
        "x": 12,
        "y": 3
      },
      "id": 87,
      "legend": {
        "alignAsTable": false,
        "avg": false,
        "current": false,
        "hideEmpty": false,
        "hideZero": false,
        "max": false,
        "min": false,
        "rightSide": true,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_request_duration_seconds_bucket{reporter=\"source\",destination_service=~\"$service\"}[1m])) by (le))",
          "format": "time_series",
          "interval": "",
          "intervalFactor": 1,
          "legendFormat": "P50",
          "refId": "A"
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_request_duration_seconds_bucket{reporter=\"source\",destination_service=~\"$service\"}[1m])) by (le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "P90",
          "refId": "B"
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_request_duration_seconds_bucket{reporter=\"source\",destination_service=~\"$service\"}[1m])) by (le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "P99",
          "refId": "C"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Client Request Duration",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "s",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "cacheTimeout": null,
      "colorBackground": false,
      "colorValue": false,
      "colors": [
        "#299c46",
        "rgba(237, 129, 40, 0.89)",
        "#d44a3a"
      ],
      "datasource": "Prometheus",
      "format": "Bps",
      "gauge": {
        "maxValue": 100,
        "minValue": 0,
        "show": false,
        "thresholdLabels": false,
        "thresholdMarkers": true
      },
      "gridPos": {
        "h": 4,
        "w": 6,
        "x": 18,
        "y": 3
      },
      "id": 84,
      "interval": null,
      "links": [],
      "mappingType": 1,
      "mappingTypes": [
        {
          "name": "value to text",
          "value": 1
        },
        {
          "name": "range to text",
          "value": 2
        }
      ],
      "maxDataPoints": 100,
      "nullPointMode": "connected",
      "nullText": null,
      "postfix": "",
      "postfixFontSize": "50%",
      "prefix": "",
      "prefixFontSize": "50%",
      "rangeMaps": [
        {
          "from": "null",
          "text": "N/A",
          "to": "null"
        }
      ],
      "sparkline": {
        "fillColor": "rgba(31, 118, 189, 0.18)",
        "full": true,
        "lineColor": "rgb(31, 120, 193)",
        "show": true
      },
      "tableColumn": "",
      "targets": [
        {
          "expr": "sum(irate(istio_tcp_received_bytes_total{reporter=\"source\", destination_service=~\"$service\"}[1m]))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "",
          "refId": "A"
        }
      ],
      "thresholds": "",
      "title": "TCP Received Bytes",
      "transparent": false,
      "type": "singlestat",
      "valueFontSize": "80%",
      "valueMaps": [
        {
          "op": "=",
          "text": "N/A",
          "value": "null"
        }
      ],
      "valueName": "avg"
    },
    {
      "cacheTimeout": null,
      "colorBackground": false,
      "colorValue": false,
      "colors": [
        "rgba(245, 54, 54, 0.9)",
        "rgba(237, 129, 40, 0.89)",
        "rgba(50, 172, 45, 0.97)"
      ],
      "datasource": "Prometheus",
      "format": "ops",
      "gauge": {
        "maxValue": 100,
        "minValue": 0,
        "show": false,
        "thresholdLabels": false,
        "thresholdMarkers": true
      },
      "gridPos": {
        "h": 4,
        "w": 6,
        "x": 0,
        "y": 7
      },
      "id": 97,
      "interval": null,
      "links": [],
      "mappingType": 1,
      "mappingTypes": [
        {
          "name": "value to text",
          "value": 1
        },
        {
          "name": "range to text",
          "value": 2
        }
      ],
      "maxDataPoints": 100,
      "nullPointMode": "connected",
      "nullText": null,
      "postfix": "",
      "postfixFontSize": "50%",
      "prefix": "",
      "prefixFontSize": "50%",
      "rangeMaps": [
        {
          "from": "null",
          "text": "N/A",
          "to": "null"
        }
      ],
      "sparkline": {
        "fillColor": "rgba(31, 118, 189, 0.18)",
        "full": true,
        "lineColor": "rgb(31, 120, 193)",
        "show": true
      },
      "tableColumn": "",
      "targets": [
        {
          "expr": "round(sum(irate(istio_requests_total{reporter=\"destination\",destination_service=~\"$service\"}[5m])), 0.001)",
          "format": "time_series",
          "intervalFactor": 1,
          "refId": "A",
          "step": 4
        }
      ],
      "thresholds": "",
      "title": "Server Request Volume",
      "transparent": false,
      "type": "singlestat",
      "valueFontSize": "80%",
      "valueMaps": [
        {
          "op": "=",
          "text": "N/A",
          "value": "null"
        }
      ],
      "valueName": "current"
    },
    {
      "cacheTimeout": null,
      "colorBackground": false,
      "colorValue": false,
      "colors": [
        "rgba(50, 172, 45, 0.97)",
        "rgba(237, 129, 40, 0.89)",
        "rgba(245, 54, 54, 0.9)"
      ],
      "datasource": "Prometheus",
      "decimals": null,
      "format": "percentunit",
      "gauge": {
        "maxValue": 100,
        "minValue": 80,
        "show": false,
        "thresholdLabels": false,
        "thresholdMarkers": false
      },
      "gridPos": {
        "h": 4,
        "w": 6,
        "x": 6,
        "y": 7
      },
      "id": 98,
      "interval": null,
      "links": [],
      "mappingType": 1,
      "mappingTypes": [
        {
          "name": "value to text",
          "value": 1
        },
        {
          "name": "range to text",
          "value": 2
        }
      ],
      "maxDataPoints": 100,
      "nullPointMode": "connected",
      "nullText": null,
      "postfix": "",
      "postfixFontSize": "50%",
      "prefix": "",
      "prefixFontSize": "50%",
      "rangeMaps": [
        {
          "from": "null",
          "text": "N/A",
          "to": "null"
        }
      ],
      "sparkline": {
        "fillColor": "rgba(31, 118, 189, 0.18)",
        "full": true,
        "lineColor": "rgb(31, 120, 193)",
        "show": true
      },
      "tableColumn": "",
      "targets": [
        {
          "expr": "sum(irate(istio_requests_total{reporter=\"destination\",destination_service=~\"$service\",response_code!~\"5.*\"}[5m])) / sum(irate(istio_requests_total{reporter=\"destination\",destination_service=~\"$service\"}[5m]))",
          "format": "time_series",
          "intervalFactor": 1,
          "refId": "B"
        }
      ],
      "thresholds": "95, 99, 99.5",
      "title": "Server Success Rate (non-5xx responses)",
      "transparent": false,
      "type": "singlestat",
      "valueFontSize": "80%",
      "valueMaps": [
        {
          "op": "=",
          "text": "N/A",
          "value": "null"
        }
      ],
      "valueName": "avg"
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 4,
        "w": 6,
        "x": 12,
        "y": 7
      },
      "id": 99,
      "legend": {
        "alignAsTable": false,
        "avg": false,
        "current": false,
        "hideEmpty": false,
        "hideZero": false,
        "max": false,
        "min": false,
        "rightSide": true,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\",destination_service=~\"$service\"}[1m])) by (le))",
          "format": "time_series",
          "interval": "",
          "intervalFactor": 1,
          "legendFormat": "P50",
          "refId": "A"
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\",destination_service=~\"$service\"}[1m])) by (le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "P90",
          "refId": "B"
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\",destination_service=~\"$service\"}[1m])) by (le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "P99",
          "refId": "C"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Server Request Duration",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "s",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "cacheTimeout": null,
      "colorBackground": false,
      "colorValue": false,
      "colors": [
        "#299c46",
        "rgba(237, 129, 40, 0.89)",
        "#d44a3a"
      ],
      "datasource": "Prometheus",
      "format": "Bps",
      "gauge": {
        "maxValue": 100,
        "minValue": 0,
        "show": false,
        "thresholdLabels": false,
        "thresholdMarkers": true
      },
      "gridPos": {
        "h": 4,
        "w": 6,
        "x": 18,
        "y": 7
      },
      "id": 100,
      "interval": null,
      "links": [],
      "mappingType": 1,
      "mappingTypes": [
        {
          "name": "value to text",
          "value": 1
        },
        {
          "name": "range to text",
          "value": 2
        }
      ],
      "maxDataPoints": 100,
      "nullPointMode": "connected",
      "nullText": null,
      "postfix": "",
      "postfixFontSize": "50%",
      "prefix": "",
      "prefixFontSize": "50%",
      "rangeMaps": [
        {
          "from": "null",
          "text": "N/A",
          "to": "null"
        }
      ],
      "sparkline": {
        "fillColor": "rgba(31, 118, 189, 0.18)",
        "full": true,
        "lineColor": "rgb(31, 120, 193)",
        "show": true
      },
      "tableColumn": "",
      "targets": [
        {
          "expr": "sum(irate(istio_tcp_sent_bytes_total{reporter=\"source\", destination_service=~\"$service\"}[1m])) ",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "",
          "refId": "A"
        }
      ],
      "thresholds": "",
      "title": "TCP Sent Bytes",
      "transparent": false,
      "type": "singlestat",
      "valueFontSize": "80%",
      "valueMaps": [
        {
          "op": "=",
          "text": "N/A",
          "value": "null"
        }
      ],
      "valueName": "avg"
    },
    {
      "content": "<div class=\"dashboard-header text-center\">\n<span>CLIENT WORKLOADS</span>\n</div>",
      "gridPos": {
        "h": 3,
        "w": 24,
        "x": 0,
        "y": 11
      },
      "id": 45,
      "links": [],
      "mode": "html",
      "title": "",
      "transparent": true,
      "type": "text"
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 0,
      "gridPos": {
        "h": 6,
        "w": 12,
        "x": 0,
        "y": 14
      },
      "id": 25,
      "legend": {
        "avg": false,
        "current": false,
        "hideEmpty": true,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null as zero",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "round(sum(irate(istio_requests_total{connection_security_policy=\"mutual_tls\",destination_service=~\"$service\",reporter=\"source\",source_workload=~\"$srcwl\",source_workload_namespace=~\"$srcns\"}[5m])) by (source_workload, source_workload_namespace, response_code), 0.001)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ source_workload }}.{{ source_workload_namespace }} : {{ response_code }} (🔐mTLS)",
          "refId": "B",
          "step": 2
        },
        {
          "expr": "round(sum(irate(istio_requests_total{connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", reporter=\"source\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[5m])) by (source_workload, source_workload_namespace, response_code), 0.001)",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ source_workload }}.{{ source_workload_namespace }} : {{ response_code }}",
          "refId": "A",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Incoming Requests by Source And Response Code",
      "tooltip": {
        "shared": false,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": [
          "total"
        ]
      },
      "yaxes": [
        {
          "format": "ops",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 12,
        "x": 12,
        "y": 14
      },
      "id": 26,
      "legend": {
        "avg": false,
        "current": false,
        "hideEmpty": true,
        "hideZero": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(irate(istio_requests_total{reporter=\"source\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\",response_code!~\"5.*\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[5m])) by (source_workload, source_workload_namespace) / sum(irate(istio_requests_total{reporter=\"source\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[5m])) by (source_workload, source_workload_namespace)",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ source_workload }}.{{ source_workload_namespace }} (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "sum(irate(istio_requests_total{reporter=\"source\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\",response_code!~\"5.*\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[5m])) by (source_workload, source_workload_namespace) / sum(irate(istio_requests_total{reporter=\"source\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[5m])) by (source_workload, source_workload_namespace)",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ source_workload }}.{{ source_workload_namespace }}",
          "refId": "B",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Incoming Success Rate (non-5xx responses) By Source",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "percentunit",
          "label": null,
          "logBase": 1,
          "max": "1.01",
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "description": "",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 0,
        "y": 20
      },
      "id": 27,
      "legend": {
        "alignAsTable": false,
        "avg": false,
        "current": false,
        "hideEmpty": true,
        "hideZero": false,
        "max": false,
        "min": false,
        "rightSide": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_request_duration_seconds_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P50 (🔐mTLS)",
          "refId": "D",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_request_duration_seconds_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P90 (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_request_duration_seconds_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P95 (🔐mTLS)",
          "refId": "B",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_request_duration_seconds_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P99 (🔐mTLS)",
          "refId": "C",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_request_duration_seconds_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P50",
          "refId": "E",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_request_duration_seconds_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P90",
          "refId": "F",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_request_duration_seconds_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P95",
          "refId": "G",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_request_duration_seconds_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P99",
          "refId": "H",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Incoming Request Duration by Source",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "s",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 8,
        "y": 20
      },
      "id": 28,
      "legend": {
        "alignAsTable": false,
        "avg": false,
        "current": false,
        "hideEmpty": true,
        "max": false,
        "min": false,
        "rightSide": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_request_bytes_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P50 (🔐mTLS)",
          "refId": "D",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_request_bytes_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}}  P90 (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_request_bytes_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P95 (🔐mTLS)",
          "refId": "B",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_request_bytes_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}}  P99 (🔐mTLS)",
          "refId": "C",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_request_bytes_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P50",
          "refId": "E",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_request_bytes_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P90",
          "refId": "F",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_request_bytes_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P95",
          "refId": "G",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_request_bytes_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P99",
          "refId": "H",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Incoming Request Size By Source",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "decbytes",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 16,
        "y": 20
      },
      "id": 68,
      "legend": {
        "alignAsTable": false,
        "avg": false,
        "current": false,
        "hideEmpty": true,
        "max": false,
        "min": false,
        "rightSide": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_response_bytes_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P50 (🔐mTLS)",
          "refId": "D",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_response_bytes_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}}  P90 (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_response_bytes_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P95 (🔐mTLS)",
          "refId": "B",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_response_bytes_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}}  P99 (🔐mTLS)",
          "refId": "C",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_response_bytes_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P50",
          "refId": "E",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_response_bytes_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P90",
          "refId": "F",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_response_bytes_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P95",
          "refId": "G",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_response_bytes_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P99",
          "refId": "H",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Response Size By Source",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "decbytes",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 12,
        "x": 0,
        "y": 26
      },
      "id": 80,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "round(sum(irate(istio_tcp_received_bytes_total{reporter=\"source\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace), 0.001)",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ source_workload }}.{{ source_workload_namespace}} (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "round(sum(irate(istio_tcp_received_bytes_total{reporter=\"source\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace), 0.001)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ source_workload }}.{{ source_workload_namespace}}",
          "refId": "B",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Bytes Received from Incoming TCP Connection",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "Bps",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 12,
        "x": 12,
        "y": 26
      },
      "id": 82,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "round(sum(irate(istio_tcp_sent_bytes_total{connection_security_policy=\"mutual_tls\", reporter=\"source\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace), 0.001)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ source_workload }}.{{ source_workload_namespace}} (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "round(sum(irate(istio_tcp_sent_bytes_total{connection_security_policy!=\"mutual_tls\", reporter=\"source\", destination_service=~\"$service\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace), 0.001)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ source_workload }}.{{ source_workload_namespace}}",
          "refId": "B",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Bytes Sent to Incoming TCP Connection",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "Bps",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "content": "<div class=\"dashboard-header text-center\">\n<span>SERVICE WORKLOADS</span>\n</div>",
      "gridPos": {
        "h": 3,
        "w": 24,
        "x": 0,
        "y": 32
      },
      "id": 69,
      "links": [],
      "mode": "html",
      "title": "",
      "transparent": true,
      "type": "text"
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 0,
      "gridPos": {
        "h": 6,
        "w": 12,
        "x": 0,
        "y": 35
      },
      "id": 90,
      "legend": {
        "avg": false,
        "current": false,
        "hideEmpty": true,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null as zero",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "round(sum(irate(istio_requests_total{connection_security_policy=\"mutual_tls\",destination_service=~\"$service\",reporter=\"destination\",destination_workload=~\"$dstwl\",destination_workload_namespace=~\"$dstns\"}[5m])) by (destination_workload, destination_workload_namespace, response_code), 0.001)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} : {{ response_code }} (🔐mTLS)",
          "refId": "B",
          "step": 2
        },
        {
          "expr": "round(sum(irate(istio_requests_total{connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", reporter=\"destination\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[5m])) by (destination_workload, destination_workload_namespace, response_code), 0.001)",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} : {{ response_code }}",
          "refId": "A",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Incoming Requests by Destination And Response Code",
      "tooltip": {
        "shared": false,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": [
          "total"
        ]
      },
      "yaxes": [
        {
          "format": "ops",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 12,
        "x": 12,
        "y": 35
      },
      "id": 91,
      "legend": {
        "avg": false,
        "current": false,
        "hideEmpty": true,
        "hideZero": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(irate(istio_requests_total{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\",response_code!~\"5.*\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[5m])) by (destination_workload, destination_workload_namespace) / sum(rate(istio_requests_total{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[5m])) by (destination_workload, destination_workload_namespace)",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "sum(irate(istio_requests_total{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\",response_code!~\"5.*\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[5m])) by (destination_workload, destination_workload_namespace) / sum(rate(istio_requests_total{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[5m])) by (destination_workload, destination_workload_namespace)",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }}",
          "refId": "B",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Incoming Success Rate (non-5xx responses) By Source",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "percentunit",
          "label": null,
          "logBase": 1,
          "max": "1.01",
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "description": "",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 0,
        "y": 41
      },
      "id": 94,
      "legend": {
        "alignAsTable": false,
        "avg": false,
        "current": false,
        "hideEmpty": true,
        "hideZero": false,
        "max": false,
        "min": false,
        "rightSide": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} P50 (🔐mTLS)",
          "refId": "D",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} P90 (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} P95 (🔐mTLS)",
          "refId": "B",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} P99 (🔐mTLS)",
          "refId": "C",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} P50",
          "refId": "E",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} P90",
          "refId": "F",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} P95",
          "refId": "G",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} P99",
          "refId": "H",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Incoming Request Duration by Source",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "s",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 8,
        "y": 41
      },
      "id": 95,
      "legend": {
        "alignAsTable": false,
        "avg": false,
        "current": false,
        "hideEmpty": true,
        "max": false,
        "min": false,
        "rightSide": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_request_bytes_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} P50 (🔐mTLS)",
          "refId": "D",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_request_bytes_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }}  P90 (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_request_bytes_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} P95 (🔐mTLS)",
          "refId": "B",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_request_bytes_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }}  P99 (🔐mTLS)",
          "refId": "C",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_request_bytes_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} P50",
          "refId": "E",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_request_bytes_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} P90",
          "refId": "F",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_request_bytes_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} P95",
          "refId": "G",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_request_bytes_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} P99",
          "refId": "H",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Incoming Request Size By Source",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "decbytes",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 16,
        "y": 41
      },
      "id": 96,
      "legend": {
        "alignAsTable": false,
        "avg": false,
        "current": false,
        "hideEmpty": true,
        "max": false,
        "min": false,
        "rightSide": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_response_bytes_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} P50 (🔐mTLS)",
          "refId": "D",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_response_bytes_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }}  P90 (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_response_bytes_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} P95 (🔐mTLS)",
          "refId": "B",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_response_bytes_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }}  P99 (🔐mTLS)",
          "refId": "C",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_response_bytes_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} P50",
          "refId": "E",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_response_bytes_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} P90",
          "refId": "F",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_response_bytes_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} P95",
          "refId": "G",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_response_bytes_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace }} P99",
          "refId": "H",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Response Size By Source",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "decbytes",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 12,
        "x": 0,
        "y": 47
      },
      "id": 92,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "round(sum(irate(istio_tcp_received_bytes_total{reporter=\"source\", connection_security_policy=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace), 0.001)",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace}} (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "round(sum(irate(istio_tcp_received_bytes_total{reporter=\"source\", connection_security_policy!=\"mutual_tls\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace), 0.001)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{ destination_workload_namespace}}",
          "refId": "B",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Bytes Received from Incoming TCP Connection",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "Bps",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 12,
        "x": 12,
        "y": 47
      },
      "id": 93,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "round(sum(irate(istio_tcp_sent_bytes_total{connection_security_policy=\"mutual_tls\", reporter=\"source\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace), 0.001)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{destination_workload_namespace }} (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "round(sum(irate(istio_tcp_sent_bytes_total{connection_security_policy!=\"mutual_tls\", reporter=\"source\", destination_service=~\"$service\", destination_workload=~\"$dstwl\", destination_workload_namespace=~\"$dstns\"}[1m])) by (destination_workload, destination_workload_namespace), 0.001)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ destination_workload }}.{{destination_workload_namespace }}",
          "refId": "B",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Bytes Sent to Incoming TCP Connection",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "Bps",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    }
  ],
  "refresh": "10s",
  "schemaVersion": 16,
  "style": "dark",
  "tags": [],
  "templating": {
    "list": [
      {
        "allValue": null,
        "datasource": "Prometheus",
        "hide": 0,
        "includeAll": false,
        "label": "Service",
        "multi": false,
        "name": "service",
        "options": [],
        "query": "label_values(destination_service)",
        "refresh": 1,
        "regex": "",
        "sort": 0,
        "tagValuesQuery": "",
        "tags": [],
        "tagsQuery": "",
        "type": "query",
        "useTags": false
      },
      {
        "allValue": null,
        "current": {
          "text": "All",
          "value": "$__all"
        },
        "datasource": "Prometheus",
        "hide": 0,
        "includeAll": true,
        "label": "Client Workload Namespace",
        "multi": true,
        "name": "srcns",
        "options": [],
        "query": "query_result( sum(istio_requests_total{reporter=\"destination\", destination_service=\"$service\"}) by (source_workload_namespace) or sum(istio_tcp_sent_bytes_total{reporter=\"destination\", destination_service=~\"$service\"}) by (source_workload_namespace))",
        "refresh": 1,
        "regex": "/.*namespace=\"([^\"]*).*/",
        "sort": 2,
        "tagValuesQuery": "",
        "tags": [],
        "tagsQuery": "",
        "type": "query",
        "useTags": false
      },
      {
        "allValue": null,
        "current": {
          "text": "All",
          "value": "$__all"
        },
        "datasource": "Prometheus",
        "hide": 0,
        "includeAll": true,
        "label": "Client Workload",
        "multi": true,
        "name": "srcwl",
        "options": [],
        "query": "query_result( sum(istio_requests_total{reporter=\"destination\", destination_service=~\"$service\", source_workload_namespace=~\"$srcns\"}) by (source_workload) or sum(istio_tcp_sent_bytes_total{reporter=\"destination\", destination_service=~\"$service\", source_workload_namespace=~\"$srcns\"}) by (source_workload))",
        "refresh": 1,
        "regex": "/.*workload=\"([^\"]*).*/",
        "sort": 3,
        "tagValuesQuery": "",
        "tags": [],
        "tagsQuery": "",
        "type": "query",
        "useTags": false
      },
      {
        "allValue": null,
        "current": {
          "text": "All",
          "value": "$__all"
        },
        "datasource": "Prometheus",
        "hide": 0,
        "includeAll": true,
        "label": "Service Workload Namespace",
        "multi": true,
        "name": "dstns",
        "options": [],
        "query": "query_result( sum(istio_requests_total{reporter=\"destination\", destination_service=\"$service\"}) by (destination_workload_namespace) or sum(istio_tcp_sent_bytes_total{reporter=\"destination\", destination_service=~\"$service\"}) by (destination_workload_namespace))",
        "refresh": 1,
        "regex": "/.*namespace=\"([^\"]*).*/",
        "sort": 2,
        "tagValuesQuery": "",
        "tags": [],
        "tagsQuery": "",
        "type": "query",
        "useTags": false
      },
      {
        "allValue": null,
        "current": {
          "text": "All",
          "value": "$__all"
        },
        "datasource": "Prometheus",
        "hide": 0,
        "includeAll": true,
        "label": "Service Workload",
        "multi": true,
        "name": "dstwl",
        "options": [],
        "query": "query_result( sum(istio_requests_total{reporter=\"destination\", destination_service=~\"$service\", destination_workload_namespace=~\"$dstns\"}) by (destination_workload) or sum(istio_tcp_sent_bytes_total{reporter=\"destination\", destination_service=~\"$service\", destination_workload_namespace=~\"$dstns\"}) by (destination_workload))",
        "refresh": 1,
        "regex": "/.*workload=\"([^\"]*).*/",
        "sort": 3,
        "tagValuesQuery": "",
        "tags": [],
        "tagsQuery": "",
        "type": "query",
        "useTags": false
      }
    ]
  },
  "time": {
    "from": "now-5m",
    "to": "now"
  },
  "timepicker": {
    "refresh_intervals": [
      "5s",
      "10s",
      "30s",
      "1m",
      "5m",
      "15m",
      "30m",
      "1h",
      "2h",
      "1d"
    ],
    "time_options": [
      "5m",
      "15m",
      "1h",
      "6h",
      "12h",
      "24h",
      "2d",
      "7d",
      "30d"
    ]
  },
  "timezone": "",
  "title": "Istio Service Dashboard",
  "uid": "LJ_uJAvmk",
  "version": 1
}
'
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-grafana-configuration-dashboards-istio-workload-dashboard
  labels:
    app: grafana
    chart: grafana
    heritage: Tiller
    release: istio
    istio: grafana
data:
  istio-workload-dashboard.json: '{
  "__inputs": [
    {
      "name": "DS_PROMETHEUS",
      "label": "Prometheus",
      "description": "",
      "type": "datasource",
      "pluginId": "prometheus",
      "pluginName": "Prometheus"
    }
  ],
  "__requires": [
    {
      "type": "grafana",
      "id": "grafana",
      "name": "Grafana",
      "version": "5.0.4"
    },
    {
      "type": "panel",
      "id": "graph",
      "name": "Graph",
      "version": "5.0.0"
    },
    {
      "type": "datasource",
      "id": "prometheus",
      "name": "Prometheus",
      "version": "5.0.0"
    },
    {
      "type": "panel",
      "id": "singlestat",
      "name": "Singlestat",
      "version": "5.0.0"
    },
    {
      "type": "panel",
      "id": "text",
      "name": "Text",
      "version": "5.0.0"
    }
  ],
  "annotations": {
    "list": [
      {
        "builtIn": 1,
        "datasource": "-- Grafana --",
        "enable": true,
        "hide": true,
        "iconColor": "rgba(0, 211, 255, 1)",
        "name": "Annotations & Alerts",
        "type": "dashboard"
      }
    ]
  },
  "editable": false,
  "gnetId": null,
  "graphTooltip": 0,
  "id": null,
  "iteration": 1531345461465,
  "links": [],
  "panels": [
    {
      "content": "<div class=\"dashboard-header text-center\">\n<span>WORKLOAD: $workload.$namespace</span>\n</div>",
      "gridPos": {
        "h": 3,
        "w": 24,
        "x": 0,
        "y": 0
      },
      "id": 89,
      "links": [],
      "mode": "html",
      "title": "",
      "transparent": true,
      "type": "text"
    },
    {
      "cacheTimeout": null,
      "colorBackground": false,
      "colorValue": false,
      "colors": [
        "rgba(245, 54, 54, 0.9)",
        "rgba(237, 129, 40, 0.89)",
        "rgba(50, 172, 45, 0.97)"
      ],
      "datasource": "Prometheus",
      "format": "ops",
      "gauge": {
        "maxValue": 100,
        "minValue": 0,
        "show": false,
        "thresholdLabels": false,
        "thresholdMarkers": true
      },
      "gridPos": {
        "h": 4,
        "w": 8,
        "x": 0,
        "y": 3
      },
      "id": 12,
      "interval": null,
      "links": [],
      "mappingType": 1,
      "mappingTypes": [
        {
          "name": "value to text",
          "value": 1
        },
        {
          "name": "range to text",
          "value": 2
        }
      ],
      "maxDataPoints": 100,
      "nullPointMode": "connected",
      "nullText": null,
      "postfix": "",
      "postfixFontSize": "50%",
      "prefix": "",
      "prefixFontSize": "50%",
      "rangeMaps": [
        {
          "from": "null",
          "text": "N/A",
          "to": "null"
        }
      ],
      "sparkline": {
        "fillColor": "rgba(31, 118, 189, 0.18)",
        "full": true,
        "lineColor": "rgb(31, 120, 193)",
        "show": true
      },
      "tableColumn": "",
      "targets": [
        {
          "expr": "round(sum(irate(istio_requests_total{reporter=\"destination\",destination_workload_namespace=~\"$namespace\",destination_workload=~\"$workload\"}[5m])), 0.001)",
          "format": "time_series",
          "intervalFactor": 1,
          "refId": "A",
          "step": 4
        }
      ],
      "thresholds": "",
      "title": "Incoming Request Volume",
      "transparent": false,
      "type": "singlestat",
      "valueFontSize": "80%",
      "valueMaps": [
        {
          "op": "=",
          "text": "N/A",
          "value": "null"
        }
      ],
      "valueName": "current"
    },
    {
      "cacheTimeout": null,
      "colorBackground": false,
      "colorValue": false,
      "colors": [
        "rgba(50, 172, 45, 0.97)",
        "rgba(237, 129, 40, 0.89)",
        "rgba(245, 54, 54, 0.9)"
      ],
      "datasource": "Prometheus",
      "decimals": null,
      "format": "percentunit",
      "gauge": {
        "maxValue": 100,
        "minValue": 80,
        "show": false,
        "thresholdLabels": false,
        "thresholdMarkers": false
      },
      "gridPos": {
        "h": 4,
        "w": 8,
        "x": 8,
        "y": 3
      },
      "id": 14,
      "interval": null,
      "links": [],
      "mappingType": 1,
      "mappingTypes": [
        {
          "name": "value to text",
          "value": 1
        },
        {
          "name": "range to text",
          "value": 2
        }
      ],
      "maxDataPoints": 100,
      "nullPointMode": "connected",
      "nullText": null,
      "postfix": "",
      "postfixFontSize": "50%",
      "prefix": "",
      "prefixFontSize": "50%",
      "rangeMaps": [
        {
          "from": "null",
          "text": "N/A",
          "to": "null"
        }
      ],
      "sparkline": {
        "fillColor": "rgba(31, 118, 189, 0.18)",
        "full": true,
        "lineColor": "rgb(31, 120, 193)",
        "show": true
      },
      "tableColumn": "",
      "targets": [
        {
          "expr": "sum(irate(istio_requests_total{reporter=\"destination\",destination_workload_namespace=~\"$namespace\",destination_workload=~\"$workload\",response_code!~\"5.*\"}[5m])) / sum(irate(istio_requests_total{reporter=\"destination\",destination_workload_namespace=~\"$namespace\",destination_workload=~\"$workload\"}[5m]))",
          "format": "time_series",
          "intervalFactor": 1,
          "refId": "B"
        }
      ],
      "thresholds": "95, 99, 99.5",
      "title": "Incoming Success Rate (non-5xx responses)",
      "transparent": false,
      "type": "singlestat",
      "valueFontSize": "80%",
      "valueMaps": [
        {
          "op": "=",
          "text": "N/A",
          "value": "null"
        }
      ],
      "valueName": "avg"
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 4,
        "w": 8,
        "x": 16,
        "y": 3
      },
      "id": 87,
      "legend": {
        "alignAsTable": false,
        "avg": false,
        "current": false,
        "hideEmpty": false,
        "hideZero": false,
        "max": false,
        "min": false,
        "rightSide": true,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\",destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\"}[1m])) by (le))",
          "format": "time_series",
          "interval": "",
          "intervalFactor": 1,
          "legendFormat": "P50",
          "refId": "A"
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\",destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\"}[1m])) by (le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "P90",
          "refId": "B"
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\",destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\"}[1m])) by (le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "P99",
          "refId": "C"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Request Duration",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "s",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ]
    },
    {
      "cacheTimeout": null,
      "colorBackground": false,
      "colorValue": false,
      "colors": [
        "#299c46",
        "rgba(237, 129, 40, 0.89)",
        "#d44a3a"
      ],
      "datasource": "Prometheus",
      "format": "Bps",
      "gauge": {
        "maxValue": 100,
        "minValue": 0,
        "show": false,
        "thresholdLabels": false,
        "thresholdMarkers": true
      },
      "gridPos": {
        "h": 4,
        "w": 12,
        "x": 0,
        "y": 7
      },
      "id": 84,
      "interval": null,
      "links": [],
      "mappingType": 1,
      "mappingTypes": [
        {
          "name": "value to text",
          "value": 1
        },
        {
          "name": "range to text",
          "value": 2
        }
      ],
      "maxDataPoints": 100,
      "nullPointMode": "connected",
      "nullText": null,
      "postfix": "",
      "postfixFontSize": "50%",
      "prefix": "",
      "prefixFontSize": "50%",
      "rangeMaps": [
        {
          "from": "null",
          "text": "N/A",
          "to": "null"
        }
      ],
      "sparkline": {
        "fillColor": "rgba(31, 118, 189, 0.18)",
        "full": true,
        "lineColor": "rgb(31, 120, 193)",
        "show": true
      },
      "tableColumn": "",
      "targets": [
        {
          "expr": "sum(irate(istio_tcp_sent_bytes_total{reporter=\"destination\", destination_workload_namespace=~\"$namespace\", destination_workload=~\"$workload\"}[1m])) + sum(irate(istio_tcp_received_bytes_total{reporter=\"destination\", destination_workload_namespace=~\"$namespace\", destination_workload=~\"$workload\"}[1m]))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "",
          "refId": "A"
        }
      ],
      "thresholds": "",
      "title": "TCP Server Traffic",
      "transparent": false,
      "type": "singlestat",
      "valueFontSize": "80%",
      "valueMaps": [
        {
          "op": "=",
          "text": "N/A",
          "value": "null"
        }
      ],
      "valueName": "avg"
    },
    {
      "cacheTimeout": null,
      "colorBackground": false,
      "colorValue": false,
      "colors": [
        "#299c46",
        "rgba(237, 129, 40, 0.89)",
        "#d44a3a"
      ],
      "datasource": "Prometheus",
      "format": "Bps",
      "gauge": {
        "maxValue": 100,
        "minValue": 0,
        "show": false,
        "thresholdLabels": false,
        "thresholdMarkers": true
      },
      "gridPos": {
        "h": 4,
        "w": 12,
        "x": 12,
        "y": 7
      },
      "id": 85,
      "interval": null,
      "links": [],
      "mappingType": 1,
      "mappingTypes": [
        {
          "name": "value to text",
          "value": 1
        },
        {
          "name": "range to text",
          "value": 2
        }
      ],
      "maxDataPoints": 100,
      "nullPointMode": "connected",
      "nullText": null,
      "postfix": "",
      "postfixFontSize": "50%",
      "prefix": "",
      "prefixFontSize": "50%",
      "rangeMaps": [
        {
          "from": "null",
          "text": "N/A",
          "to": "null"
        }
      ],
      "sparkline": {
        "fillColor": "rgba(31, 118, 189, 0.18)",
        "full": true,
        "lineColor": "rgb(31, 120, 193)",
        "show": true
      },
      "tableColumn": "",
      "targets": [
        {
          "expr": "sum(irate(istio_tcp_sent_bytes_total{reporter=\"source\", source_workload_namespace=~\"$namespace\", source_workload=~\"$workload\"}[1m])) + sum(irate(istio_tcp_received_bytes_total{reporter=\"source\", source_workload_namespace=~\"$namespace\", source_workload=~\"$workload\"}[1m]))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "",
          "refId": "A"
        }
      ],
      "thresholds": "",
      "title": "TCP Client Traffic",
      "transparent": false,
      "type": "singlestat",
      "valueFontSize": "80%",
      "valueMaps": [
        {
          "op": "=",
          "text": "N/A",
          "value": "null"
        }
      ],
      "valueName": "avg"
    },
    {
      "content": "<div class=\"dashboard-header text-center\">\n<span>INBOUND WORKLOADS</span>\n</div>",
      "gridPos": {
        "h": 3,
        "w": 24,
        "x": 0,
        "y": 11
      },
      "id": 45,
      "links": [],
      "mode": "html",
      "title": "",
      "transparent": true,
      "type": "text"
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 0,
      "gridPos": {
        "h": 6,
        "w": 12,
        "x": 0,
        "y": 14
      },
      "id": 25,
      "legend": {
        "avg": false,
        "current": false,
        "hideEmpty": true,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null as zero",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "round(sum(irate(istio_requests_total{connection_security_policy=\"mutual_tls\", destination_workload_namespace=~\"$namespace\", destination_workload=~\"$workload\", reporter=\"destination\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[5m])) by (source_workload, source_workload_namespace, response_code), 0.001)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ source_workload }}.{{ source_workload_namespace }} : {{ response_code }} (🔐mTLS)",
          "refId": "B",
          "step": 2
        },
        {
          "expr": "round(sum(irate(istio_requests_total{connection_security_policy!=\"mutual_tls\", destination_workload_namespace=~\"$namespace\", destination_workload=~\"$workload\", reporter=\"destination\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[5m])) by (source_workload, source_workload_namespace, response_code), 0.001)",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ source_workload }}.{{ source_workload_namespace }} : {{ response_code }}",
          "refId": "A",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Incoming Requests by Source And Response Code",
      "tooltip": {
        "shared": false,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": [
          "total"
        ]
      },
      "yaxes": [
        {
          "format": "ops",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ]
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 12,
        "x": 12,
        "y": 14
      },
      "id": 26,
      "legend": {
        "avg": false,
        "current": false,
        "hideEmpty": true,
        "hideZero": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(irate(istio_requests_total{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_workload_namespace=~\"$namespace\", destination_workload=~\"$workload\",response_code!~\"5.*\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[5m])) by (source_workload, source_workload_namespace) / sum(rate(istio_requests_total{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_workload_namespace=~\"$namespace\", destination_workload=~\"$workload\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[5m])) by (source_workload, source_workload_namespace)",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ source_workload }}.{{ source_workload_namespace }} (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "sum(irate(istio_requests_total{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_workload_namespace=~\"$namespace\", destination_workload=~\"$workload\",response_code!~\"5.*\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[5m])) by (source_workload, source_workload_namespace) / sum(rate(istio_requests_total{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_workload_namespace=~\"$namespace\", destination_workload=~\"$workload\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[5m])) by (source_workload, source_workload_namespace)",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ source_workload }}.{{ source_workload_namespace }}",
          "refId": "B",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Incoming Success Rate (non-5xx responses) By Source",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "percentunit",
          "label": null,
          "logBase": 1,
          "max": "1.01",
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ]
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "description": "",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 0,
        "y": 20
      },
      "id": 27,
      "legend": {
        "alignAsTable": false,
        "avg": false,
        "current": false,
        "hideEmpty": true,
        "hideZero": false,
        "max": false,
        "min": false,
        "rightSide": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P50 (🔐mTLS)",
          "refId": "D",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P90 (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P95 (🔐mTLS)",
          "refId": "B",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P99 (🔐mTLS)",
          "refId": "C",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P50",
          "refId": "E",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P90",
          "refId": "F",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P95",
          "refId": "G",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_request_duration_seconds_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P99",
          "refId": "H",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Incoming Request Duration by Source",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "s",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ]
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 8,
        "y": 20
      },
      "id": 28,
      "legend": {
        "alignAsTable": false,
        "avg": false,
        "current": false,
        "hideEmpty": true,
        "max": false,
        "min": false,
        "rightSide": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_request_bytes_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P50 (🔐mTLS)",
          "refId": "D",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_request_bytes_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}}  P90 (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_request_bytes_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P95 (🔐mTLS)",
          "refId": "B",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_request_bytes_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}}  P99 (🔐mTLS)",
          "refId": "C",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_request_bytes_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P50",
          "refId": "E",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_request_bytes_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P90",
          "refId": "F",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_request_bytes_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P95",
          "refId": "G",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_request_bytes_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P99",
          "refId": "H",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Incoming Request Size By Source",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "decbytes",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ]
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 16,
        "y": 20
      },
      "id": 68,
      "legend": {
        "alignAsTable": false,
        "avg": false,
        "current": false,
        "hideEmpty": true,
        "max": false,
        "min": false,
        "rightSide": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_response_bytes_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P50 (🔐mTLS)",
          "refId": "D",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_response_bytes_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}}  P90 (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_response_bytes_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P95 (🔐mTLS)",
          "refId": "B",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_response_bytes_bucket{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}}  P99 (🔐mTLS)",
          "refId": "C",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_response_bytes_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P50",
          "refId": "E",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_response_bytes_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P90",
          "refId": "F",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_response_bytes_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P95",
          "refId": "G",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_response_bytes_bucket{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_workload=~\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{source_workload}}.{{source_workload_namespace}} P99",
          "refId": "H",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Response Size By Source",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "decbytes",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ]
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 12,
        "x": 0,
        "y": 26
      },
      "id": 80,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "round(sum(irate(istio_tcp_received_bytes_total{reporter=\"destination\", connection_security_policy=\"mutual_tls\", destination_workload_namespace=~\"$namespace\", destination_workload=~\"$workload\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace), 0.001)",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ source_workload }}.{{ source_workload_namespace}} (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "round(sum(irate(istio_tcp_received_bytes_total{reporter=\"destination\", connection_security_policy!=\"mutual_tls\", destination_workload_namespace=~\"$namespace\", destination_workload=~\"$workload\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace), 0.001)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ source_workload }}.{{ source_workload_namespace}}",
          "refId": "B",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Bytes Received from Incoming TCP Connection",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "Bps",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ]
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 12,
        "x": 12,
        "y": 26
      },
      "id": 82,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "round(sum(irate(istio_tcp_sent_bytes_total{connection_security_policy=\"mutual_tls\", reporter=\"destination\", destination_workload_namespace=~\"$namespace\", destination_workload=~\"$workload\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace), 0.001)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ source_workload }}.{{ source_workload_namespace}} (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "round(sum(irate(istio_tcp_sent_bytes_total{connection_security_policy!=\"mutual_tls\", reporter=\"destination\", destination_workload_namespace=~\"$namespace\", destination_workload=~\"$workload\", source_workload=~\"$srcwl\", source_workload_namespace=~\"$srcns\"}[1m])) by (source_workload, source_workload_namespace), 0.001)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ source_workload }}.{{ source_workload_namespace}}",
          "refId": "B",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Bytes Sent to Incoming TCP Connection",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "Bps",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ]
    },
    {
      "content": "<div class=\"dashboard-header text-center\">\n<span>OUTBOUND SERVICES</span>\n</div>",
      "gridPos": {
        "h": 3,
        "w": 24,
        "x": 0,
        "y": 32
      },
      "id": 69,
      "links": [],
      "mode": "html",
      "title": "",
      "transparent": true,
      "type": "text"
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 0,
      "gridPos": {
        "h": 6,
        "w": 12,
        "x": 0,
        "y": 35
      },
      "id": 70,
      "legend": {
        "avg": false,
        "current": false,
        "hideEmpty": true,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null as zero",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "round(sum(irate(istio_requests_total{connection_security_policy=\"mutual_tls\", source_workload_namespace=~\"$namespace\", source_workload=~\"$workload\", reporter=\"source\", destination_service=~\"$dstsvc\"}[5m])) by (destination_service, response_code), 0.001)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} : {{ response_code }} (🔐mTLS)",
          "refId": "B",
          "step": 2
        },
        {
          "expr": "round(sum(irate(istio_requests_total{connection_security_policy!=\"mutual_tls\", source_workload_namespace=~\"$namespace\", source_workload=~\"$workload\", reporter=\"source\", destination_service=~\"$dstsvc\"}[5m])) by (destination_service, response_code), 0.001)",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} : {{ response_code }}",
          "refId": "A",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Outgoing Requests by Destination And Response Code",
      "tooltip": {
        "shared": false,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": [
          "total"
        ]
      },
      "yaxes": [
        {
          "format": "ops",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ]
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 12,
        "x": 12,
        "y": 35
      },
      "id": 71,
      "legend": {
        "avg": false,
        "current": false,
        "hideEmpty": true,
        "hideZero": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(irate(istio_requests_total{reporter=\"source\", connection_security_policy=\"mutual_tls\", source_workload_namespace=~\"$namespace\", source_workload=~\"$workload\",response_code!~\"5.*\", destination_service=~\"$dstsvc\"}[5m])) by (destination_service) / sum(irate(istio_requests_total{reporter=\"source\", connection_security_policy=\"mutual_tls\", source_workload_namespace=~\"$namespace\", source_workload=~\"$workload\", destination_service=~\"$dstsvc\"}[5m])) by (destination_service)",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "sum(irate(istio_requests_total{reporter=\"source\", connection_security_policy!=\"mutual_tls\", source_workload_namespace=~\"$namespace\", source_workload=~\"$workload\",response_code!~\"5.*\", destination_service=~\"$dstsvc\"}[5m])) by (destination_service) / sum(irate(istio_requests_total{reporter=\"source\", connection_security_policy!=\"mutual_tls\", source_workload_namespace=~\"$namespace\", source_workload=~\"$workload\", destination_service=~\"$dstsvc\"}[5m])) by (destination_service)",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{destination_service }}",
          "refId": "B",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Outgoing Success Rate (non-5xx responses) By Destination",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "percentunit",
          "label": null,
          "logBase": 1,
          "max": "1.01",
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ]
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "description": "",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 0,
        "y": 41
      },
      "id": 72,
      "legend": {
        "alignAsTable": false,
        "avg": false,
        "current": false,
        "hideEmpty": true,
        "hideZero": false,
        "max": false,
        "min": false,
        "rightSide": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_request_duration_seconds_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P50 (🔐mTLS)",
          "refId": "D",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_request_duration_seconds_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P90 (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_request_duration_seconds_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P95 (🔐mTLS)",
          "refId": "B",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_request_duration_seconds_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P99 (🔐mTLS)",
          "refId": "C",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_request_duration_seconds_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P50",
          "refId": "E",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_request_duration_seconds_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P90",
          "refId": "F",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_request_duration_seconds_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P95",
          "refId": "G",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_request_duration_seconds_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P99",
          "refId": "H",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Outgoing Request Duration by Destination",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "s",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ]
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 8,
        "y": 41
      },
      "id": 73,
      "legend": {
        "alignAsTable": false,
        "avg": false,
        "current": false,
        "hideEmpty": true,
        "max": false,
        "min": false,
        "rightSide": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_request_bytes_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P50 (🔐mTLS)",
          "refId": "D",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_request_bytes_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P90 (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_request_bytes_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P95 (🔐mTLS)",
          "refId": "B",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_request_bytes_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P99 (🔐mTLS)",
          "refId": "C",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_request_bytes_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P50",
          "refId": "E",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_request_bytes_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P90",
          "refId": "F",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_request_bytes_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P95",
          "refId": "G",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_request_bytes_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P99",
          "refId": "H",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Outgoing Request Size By Destination",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "decbytes",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ]
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 16,
        "y": 41
      },
      "id": 74,
      "legend": {
        "alignAsTable": false,
        "avg": false,
        "current": false,
        "hideEmpty": true,
        "max": false,
        "min": false,
        "rightSide": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_response_bytes_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P50 (🔐mTLS)",
          "refId": "D",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_response_bytes_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P90 (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_response_bytes_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P95 (🔐mTLS)",
          "refId": "B",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_response_bytes_bucket{reporter=\"source\", connection_security_policy=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }}  P99 (🔐mTLS)",
          "refId": "C",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.50, sum(irate(istio_response_bytes_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P50",
          "refId": "E",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.90, sum(irate(istio_response_bytes_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P90",
          "refId": "F",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.95, sum(irate(istio_response_bytes_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P95",
          "refId": "G",
          "step": 2
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(istio_response_bytes_bucket{reporter=\"source\", connection_security_policy!=\"mutual_tls\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service, le))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} P99",
          "refId": "H",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Response Size By Destination",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "decbytes",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ]
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 12,
        "x": 0,
        "y": 47
      },
      "id": 76,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "round(sum(irate(istio_tcp_sent_bytes_total{connection_security_policy=\"mutual_tls\", reporter=\"source\", source_workload_namespace=~\"$namespace\", source_workload=~\"$workload\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service), 0.001)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "round(sum(irate(istio_tcp_sent_bytes_total{connection_security_policy!=\"mutual_tls\", reporter=\"source\", source_workload_namespace=~\"$namespace\", source_workload=~\"$workload\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service), 0.001)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }}",
          "refId": "B",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Bytes Sent on Outgoing TCP Connection",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "Bps",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ]
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 12,
        "x": 12,
        "y": 47
      },
      "id": 78,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "round(sum(irate(istio_tcp_received_bytes_total{reporter=\"source\", connection_security_policy=\"mutual_tls\", source_workload_namespace=~\"$namespace\", source_workload=~\"$workload\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service), 0.001)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }} (🔐mTLS)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "round(sum(irate(istio_tcp_received_bytes_total{reporter=\"source\", connection_security_policy!=\"mutual_tls\", source_workload_namespace=~\"$namespace\", source_workload=~\"$workload\", destination_service=~\"$dstsvc\"}[1m])) by (destination_service), 0.001)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ destination_service }}",
          "refId": "B",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Bytes Received from Outgoing TCP Connection",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "Bps",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": "0",
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ]
    }
  ],
  "refresh": "10s",
  "schemaVersion": 16,
  "style": "dark",
  "tags": [],
  "templating": {
    "list": [
      {
        "allValue": null,
        "current": {},
        "datasource": "Prometheus",
        "hide": 0,
        "includeAll": false,
        "label": "Namespace",
        "multi": false,
        "name": "namespace",
        "options": [],
        "query": "query_result(sum(istio_requests_total) by (destination_workload_namespace) or sum(istio_tcp_sent_bytes_total) by (destination_workload_namespace))",
        "refresh": 1,
        "regex": "/.*_namespace=\"([^\"]*).*/",
        "sort": 0,
        "tagValuesQuery": "",
        "tags": [],
        "tagsQuery": "",
        "type": "query",
        "useTags": false
      },
      {
        "allValue": null,
        "current": {},
        "datasource": "Prometheus",
        "hide": 0,
        "includeAll": false,
        "label": "Workload",
        "multi": false,
        "name": "workload",
        "options": [],
        "query": "query_result((sum(istio_requests_total{destination_workload_namespace=~\"$namespace\"}) by (destination_workload) or sum(istio_requests_total{source_workload_namespace=~\"$namespace\"}) by (source_workload)) or (sum(istio_tcp_sent_bytes_total{destination_workload_namespace=~\"$namespace\"}) by (destination_workload) or sum(istio_tcp_sent_bytes_total{source_workload_namespace=~\"$namespace\"}) by (source_workload)))",
        "refresh": 1,
        "regex": "/.*workload=\"([^\"]*).*/",
        "sort": 1,
        "tagValuesQuery": "",
        "tags": [],
        "tagsQuery": "",
        "type": "query",
        "useTags": false
      },
      {
        "allValue": null,
        "current": {},
        "datasource": "Prometheus",
        "hide": 0,
        "includeAll": true,
        "label": "Inbound Workload Namespace",
        "multi": true,
        "name": "srcns",
        "options": [],
        "query": "query_result( sum(istio_requests_total{reporter=\"destination\", destination_workload=\"$workload\", destination_workload_namespace=~\"$namespace\"}) by (source_workload_namespace) or sum(istio_tcp_sent_bytes_total{reporter=\"destination\", destination_workload=\"$workload\", destination_workload_namespace=~\"$namespace\"}) by (source_workload_namespace))",
        "refresh": 1,
        "regex": "/.*namespace=\"([^\"]*).*/",
        "sort": 2,
        "tagValuesQuery": "",
        "tags": [],
        "tagsQuery": "",
        "type": "query",
        "useTags": false
      },
      {
        "allValue": null,
        "current": {},
        "datasource": "Prometheus",
        "hide": 0,
        "includeAll": true,
        "label": "Inbound Workload",
        "multi": true,
        "name": "srcwl",
        "options": [],
        "query": "query_result( sum(istio_requests_total{reporter=\"destination\", destination_workload=\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload_namespace=~\"$srcns\"}) by (source_workload) or sum(istio_tcp_sent_bytes_total{reporter=\"destination\", destination_workload=\"$workload\", destination_workload_namespace=~\"$namespace\", source_workload_namespace=~\"$srcns\"}) by (source_workload))",
        "refresh": 1,
        "regex": "/.*workload=\"([^\"]*).*/",
        "sort": 3,
        "tagValuesQuery": "",
        "tags": [],
        "tagsQuery": "",
        "type": "query",
        "useTags": false
      },
      {
        "allValue": null,
        "current": {},
        "datasource": "Prometheus",
        "hide": 0,
        "includeAll": true,
        "label": "Destination Service",
        "multi": true,
        "name": "dstsvc",
        "options": [],
        "query": "query_result( sum(istio_requests_total{reporter=\"source\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\"}) by (destination_service) or sum(istio_tcp_sent_bytes_total{reporter=\"source\", source_workload=~\"$workload\", source_workload_namespace=~\"$namespace\"}) by (destination_service))",
        "refresh": 1,
        "regex": "/.*destination_service=\"([^\"]*).*/",
        "sort": 4,
        "tagValuesQuery": "",
        "tags": [],
        "tagsQuery": "",
        "type": "query",
        "useTags": false
      }
    ]
  },
  "time": {
    "from": "now-5m",
    "to": "now"
  },
  "timepicker": {
    "refresh_intervals": [
      "5s",
      "10s",
      "30s",
      "1m",
      "5m",
      "15m",
      "30m",
      "1h",
      "2h",
      "1d"
    ],
    "time_options": [
      "5m",
      "15m",
      "1h",
      "6h",
      "12h",
      "24h",
      "2d",
      "7d",
      "30d"
    ]
  },
  "timezone": "",
  "title": "Istio Workload Dashboard",
  "uid": "UbsSZTDik",
  "version": 1
}
'
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-grafana-configuration-dashboards-mixer-dashboard
  labels:
    app: grafana
    chart: grafana
    heritage: Tiller
    release: istio
    istio: grafana
data:
  mixer-dashboard.json: '{
  "__inputs": [
    {
      "name": "DS_PROMETHEUS",
      "label": "Prometheus",
      "description": "",
      "type": "datasource",
      "pluginId": "prometheus",
      "pluginName": "Prometheus"
    }
  ],
  "__requires": [
    {
      "type": "grafana",
      "id": "grafana",
      "name": "Grafana",
      "version": "5.2.3"
    },
    {
      "type": "panel",
      "id": "graph",
      "name": "Graph",
      "version": "5.0.0"
    },
    {
      "type": "datasource",
      "id": "prometheus",
      "name": "Prometheus",
      "version": "5.0.0"
    },
    {
      "type": "panel",
      "id": "text",
      "name": "Text",
      "version": "5.0.0"
    }
  ],
  "annotations": {
    "list": [
      {
        "builtIn": 1,
        "datasource": "-- Grafana --",
        "enable": true,
        "hide": true,
        "iconColor": "rgba(0, 211, 255, 1)",
        "limit": 100,
        "name": "Annotations & Alerts",
        "showIn": 0,
        "type": "dashboard"
      }
    ]
  },
  "editable": false,
  "gnetId": null,
  "graphTooltip": 1,
  "id": null,
  "iteration": 1543881232533,
  "links": [],
  "panels": [
    {
      "content": "<center><h2>Deployed Versions</h2></center>",
      "gridPos": {
        "h": 3,
        "w": 24,
        "x": 0,
        "y": 0
      },
      "height": "40",
      "id": 62,
      "links": [],
      "mode": "html",
      "title": "",
      "transparent": true,
      "type": "text"
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 5,
        "w": 24,
        "x": 0,
        "y": 3
      },
      "id": 64,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(istio_build{component=\"mixer\"}) by (tag)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ tag }}",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Mixer Versions",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "content": "<center><h2>Resource Usage</h2></center>",
      "gridPos": {
        "h": 3,
        "w": 24,
        "x": 0,
        "y": 8
      },
      "height": "40",
      "id": 29,
      "links": [],
      "mode": "html",
      "title": "",
      "transparent": true,
      "type": "text"
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 7,
        "w": 6,
        "x": 0,
        "y": 11
      },
      "id": 5,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(process_virtual_memory_bytes{job=~\"istio-telemetry|istio-policy\"}) by (job)",
          "format": "time_series",
          "instant": false,
          "intervalFactor": 2,
          "legendFormat": "Virtual Memory ({{ job }})",
          "refId": "I"
        },
        {
          "expr": "sum(process_resident_memory_bytes{job=~\"istio-telemetry|istio-policy\"}) by (job)",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "Resident Memory ({{ job }})",
          "refId": "H"
        },
        {
          "expr": "sum(go_memstats_heap_sys_bytes{job=~\"istio-telemetry|istio-policy\"}) by (job)",
          "format": "time_series",
          "hide": true,
          "intervalFactor": 2,
          "legendFormat": "heap sys ({{ job }})",
          "refId": "A"
        },
        {
          "expr": "sum(go_memstats_heap_alloc_bytes{job=~\"istio-telemetry|istio-policy\"}) by (job)",
          "format": "time_series",
          "hide": true,
          "intervalFactor": 2,
          "legendFormat": "heap alloc ({{ job }})",
          "refId": "D"
        },
        {
          "expr": "sum(go_memstats_alloc_bytes{job=~\"istio-telemetry|istio-policy\"}) by (job)",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "Alloc ({{ job }})",
          "refId": "F"
        },
        {
          "expr": "sum(go_memstats_heap_inuse_bytes{job=~\"istio-telemetry|istio-policy\"}) by (job)",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 2,
          "legendFormat": "Heap in-use ({{ job }})",
          "refId": "E"
        },
        {
          "expr": "sum(go_memstats_stack_inuse_bytes{job=~\"istio-telemetry|istio-policy\"}) by (job)",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "Stack in-use ({{ job }})",
          "refId": "G"
        },
        {
          "expr": "sum(label_replace(container_memory_usage_bytes{container_name=~\"mixer|istio-proxy\", pod_name=~\"istio-telemetry-.*|istio-policy-.*\"}, \"service\", \"$1\" , \"pod_name\", \"(istio-telemetry|istio-policy)-.*\")) by (service)",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 2,
          "legendFormat": "{{ service }} total (k8s)",
          "refId": "C"
        },
        {
          "expr": "sum(label_replace(container_memory_usage_bytes{container_name=~\"mixer|istio-proxy\", pod_name=~\"istio-telemetry-.*|istio-policy-.*\"}, \"service\", \"$1\" , \"pod_name\", \"(istio-telemetry|istio-policy)-.*\")) by (container_name, service)",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 2,
          "legendFormat": "{{ service }} - {{ container_name }} (k8s)",
          "refId": "B"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Memory",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "bytes",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 7,
        "w": 6,
        "x": 6,
        "y": 11
      },
      "id": 6,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "label_replace(sum(rate(container_cpu_usage_seconds_total{container_name=~\"mixer|istio-proxy\", pod_name=~\"istio-telemetry-.*|istio-policy-.*\"}[1m])) by (pod_name), \"service\", \"$1\" , \"pod_name\", \"(istio-telemetry|istio-policy)-.*\")",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 2,
          "legendFormat": "{{ service }} total (k8s)",
          "refId": "A"
        },
        {
          "expr": "label_replace(sum(rate(container_cpu_usage_seconds_total{container_name=~\"mixer|istio-proxy\", pod_name=~\"istio-telemetry-.*|istio-policy-.*\"}[1m])) by (container_name, pod_name), \"service\", \"$1\" , \"pod_name\", \"(istio-telemetry|istio-policy)-.*\")",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 2,
          "legendFormat": "{{ service }} - {{ container_name }} (k8s)",
          "refId": "B"
        },
        {
          "expr": "sum(irate(process_cpu_seconds_total{job=~\"istio-telemetry|istio-policy\"}[1m])) by (job)",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 2,
          "legendFormat": "{{ job }} (self-reported)",
          "refId": "C"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "CPU",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 7,
        "w": 6,
        "x": 12,
        "y": 11
      },
      "id": 7,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(process_open_fds{job=~\"istio-telemetry|istio-policy\"}) by (job)",
          "format": "time_series",
          "hide": true,
          "instant": false,
          "interval": "",
          "intervalFactor": 2,
          "legendFormat": "Open FDs ({{ job }})",
          "refId": "A"
        },
        {
          "expr": "sum(label_replace(container_fs_usage_bytes{container_name=~\"mixer|istio-proxy\", pod_name=~\"istio-telemetry-.*|istio-policy-.*\"}, \"service\", \"$1\" , \"pod_name\", \"(istio-telemetry|istio-policy)-.*\")) by (container_name, service)",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "{{ service }} - {{ container_name }}",
          "refId": "B"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Disk",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "bytes",
          "label": "",
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "decimals": null,
          "format": "none",
          "label": "",
          "logBase": 1024,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 7,
        "w": 6,
        "x": 18,
        "y": 11
      },
      "id": 4,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": false,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(go_goroutines{job=~\"istio-telemetry|istio-policy\"}) by (job)",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "Number of Goroutines ({{ job }})",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Goroutines",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": "",
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "content": "<center><h2>Mixer Overview</h2></center>",
      "gridPos": {
        "h": 3,
        "w": 24,
        "x": 0,
        "y": 18
      },
      "height": "40px",
      "id": 30,
      "links": [],
      "mode": "html",
      "title": "",
      "transparent": true,
      "type": "text"
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 6,
        "x": 0,
        "y": 21
      },
      "id": 9,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(rate(grpc_io_server_completed_rpcs[1m]))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 2,
          "legendFormat": "mixer (Total)",
          "refId": "B"
        },
        {
          "expr": "sum(rate(grpc_io_server_completed_rpcs[1m])) by (grpc_server_method)",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "mixer ({{ grpc_server_method }})",
          "refId": "C"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Incoming Requests",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "ops",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 6,
        "x": 6,
        "y": 21
      },
      "id": 8,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [
        {
          "alias": "{}",
          "yaxis": 1
        }
      ],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "histogram_quantile(0.5, sum(rate(grpc_io_server_server_latency_bucket{}[1m])) by (grpc_server_method, le))",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "{{ grpc_server_method }} 0.5",
          "refId": "B"
        },
        {
          "expr": "histogram_quantile(0.9, sum(rate(grpc_io_server_server_latency_bucket{}[1m])) by (grpc_server_method, le))",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "{{ grpc_server_method }} 0.9",
          "refId": "C"
        },
        {
          "expr": "histogram_quantile(0.99, sum(rate(grpc_io_server_server_latency_bucket{}[1m])) by (grpc_server_method, le))",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "{{ grpc_server_method }} 0.99",
          "refId": "D"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Response Durations",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "ms",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 6,
        "x": 12,
        "y": 21
      },
      "id": 11,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(rate(grpc_server_handled_total{grpc_code=~\"Unknown|Unimplemented|Internal|DataLoss\"}[1m])) by (grpc_method)",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "Mixer {{ grpc_method }}",
          "refId": "B"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Server Error Rate (5xx responses)",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 6,
        "x": 18,
        "y": 21
      },
      "id": 12,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(irate(grpc_server_handled_total{grpc_code!=\"OK\",grpc_service=~\".*Mixer\"}[1m])) by (grpc_method)",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "Mixer {{ grpc_method }}",
          "refId": "B"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Non-successes (4xxs)",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "content": "<center><h2>Adapters and Config</h2></center>",
      "gridPos": {
        "h": 3,
        "w": 24,
        "x": 0,
        "y": 27
      },
      "id": 28,
      "links": [],
      "mode": "html",
      "title": "",
      "transparent": true,
      "type": "text"
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 7,
        "w": 12,
        "x": 0,
        "y": 30
      },
      "id": 13,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(irate(mixer_runtime_dispatches_total{adapter=~\"$adapter\"}[1m])) by (adapter)",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "{{ adapter }}",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Adapter Dispatch Count",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 7,
        "w": 12,
        "x": 12,
        "y": 30
      },
      "id": 14,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "histogram_quantile(0.5, sum(irate(mixer_runtime_dispatch_duration_seconds_bucket{adapter=~\"$adapter\"}[1m])) by (adapter, le))",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "{{ adapter }} - p50",
          "refId": "A"
        },
        {
          "expr": "histogram_quantile(0.9, sum(irate(mixer_runtime_dispatch_duration_seconds_bucket{adapter=~\"$adapter\"}[1m])) by (adapter, le))",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "{{ adapter }} - p90 ",
          "refId": "B"
        },
        {
          "expr": "histogram_quantile(0.99, sum(irate(mixer_runtime_dispatch_duration_seconds_bucket{adapter=~\"$adapter\"}[1m])) by (adapter, le))",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "{{ adapter }} - p99",
          "refId": "C"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Adapter Dispatch Duration",
      "tooltip": {
        "shared": true,
        "sort": 1,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "s",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 7,
        "w": 6,
        "x": 0,
        "y": 37
      },
      "id": 60,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "scalar(topk(1, max(mixer_config_rule_config_count) by (configID)))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Rules",
          "refId": "A"
        },
        {
          "expr": "scalar(topk(1, max(mixer_config_rule_config_error_count) by (configID)))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Config Errors",
          "refId": "B"
        },
        {
          "expr": "scalar(topk(1, max(mixer_config_rule_config_match_error_count) by (configID)))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Match Errors",
          "refId": "C"
        },
        {
          "expr": "scalar(topk(1, max(mixer_config_unsatisfied_action_handler_count) by (configID)))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Unsatisfied Actions",
          "refId": "D"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Rules",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 7,
        "w": 6,
        "x": 6,
        "y": 37
      },
      "id": 56,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "scalar(topk(1, max(mixer_config_instance_config_count) by (configID)))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Instances",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Instances in Latest Config",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 7,
        "w": 6,
        "x": 12,
        "y": 37
      },
      "id": 54,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "scalar(topk(1, max(mixer_config_handler_config_count) by (configID)))",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Handlers",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Handlers in Latest Config",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 7,
        "w": 6,
        "x": 18,
        "y": 37
      },
      "id": 58,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "scalar(topk(1, max(mixer_config_attribute_count) by (configID)))",
          "format": "time_series",
          "instant": false,
          "intervalFactor": 1,
          "legendFormat": "Attributes",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Attributes in Latest Config",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "content": "<center><h2>Individual Adapters</h2></center>",
      "gridPos": {
        "h": 3,
        "w": 24,
        "x": 0,
        "y": 44
      },
      "id": 23,
      "links": [],
      "mode": "html",
      "title": "",
      "transparent": true,
      "type": "text"
    },
    {
      "collapsed": false,
      "gridPos": {
        "h": 1,
        "w": 24,
        "x": 0,
        "y": 47
      },
      "id": 46,
      "panels": [],
      "repeat": "adapter",
      "title": "$adapter Adapter",
      "type": "row"
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 7,
        "w": 12,
        "x": 0,
        "y": 48
      },
      "id": 17,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "label_replace(irate(mixer_runtime_dispatches_total{adapter=\"$adapter\"}[1m]),\"handler\", \"$1 ($3)\", \"handler\", \"(.*)\\\\.(.*)\\\\.(.*)\")",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "{{ handler }} (error: {{  error }})",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Dispatch Count By Handler",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 7,
        "w": 12,
        "x": 12,
        "y": 48
      },
      "id": 18,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "label_replace(histogram_quantile(0.5, sum(rate(mixer_runtime_dispatch_duration_seconds_bucket{adapter=\"$adapter\"}[1m])) by (handler, error, le)),  \"handler_short\", \"$1 ($3)\", \"handler\", \"(.*)\\\\.(.*)\\\\.(.*)\")",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "p50 - {{ handler_short }} (error: {{ error }})",
          "refId": "A"
        },
        {
          "expr": "label_replace(histogram_quantile(0.9, sum(irate(mixer_runtime_dispatch_duration_seconds_bucket{adapter=\"$adapter\"}[1m])) by (handler, error, le)),  \"handler_short\", \"$1 ($3)\", \"handler\", \"(.*)\\\\.(.*)\\\\.(.*)\")",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "p90 - {{ handler_short }} (error: {{ error }})",
          "refId": "D"
        },
        {
          "expr": "label_replace(histogram_quantile(0.99, sum(irate(mixer_runtime_dispatch_duration_seconds_bucket{adapter=\"$adapter\"}[1m])) by (handler, error, le)),  \"handler_short\", \"$1 ($3)\", \"handler\", \"(.*)\\\\.(.*)\\\\.(.*)\")",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "p99 - {{ handler_short }} (error: {{ error }})",
          "refId": "E"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Dispatch Duration By Handler",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "s",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    }
  ],
  "refresh": "5s",
  "schemaVersion": 16,
  "style": "dark",
  "tags": [],
  "templating": {
    "list": [
      {
        "allValue": null,
        "current": {},
        "datasource": "Prometheus",
        "hide": 0,
        "includeAll": true,
        "label": "Adapter",
        "multi": true,
        "name": "adapter",
        "options": [],
        "query": "label_values(adapter)",
        "refresh": 2,
        "regex": "",
        "sort": 1,
        "tagValuesQuery": "",
        "tags": [],
        "tagsQuery": "",
        "type": "query",
        "useTags": false
      }
    ]
  },
  "time": {
    "from": "now-5m",
    "to": "now"
  },
  "timepicker": {
    "refresh_intervals": [
      "5s",
      "10s",
      "30s",
      "1m",
      "5m",
      "15m",
      "30m",
      "1h",
      "2h",
      "1d"
    ],
    "time_options": [
      "5m",
      "15m",
      "1h",
      "6h",
      "12h",
      "24h",
      "2d",
      "7d",
      "30d"
    ]
  },
  "timezone": "",
  "title": "Istio Mixer Dashboard",
  "version": 4
}
'
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-grafana-configuration-dashboards-pilot-dashboard
  labels:
    app: grafana
    chart: grafana
    heritage: Tiller
    release: istio
    istio: grafana
data:
  pilot-dashboard.json: '{
  "__inputs": [
    {
      "name": "DS_PROMETHEUS",
      "label": "Prometheus",
      "description": "",
      "type": "datasource",
      "pluginId": "prometheus",
      "pluginName": "Prometheus"
    }
  ],
  "__requires": [
    {
      "type": "grafana",
      "id": "grafana",
      "name": "Grafana",
      "version": "5.2.3"
    },
    {
      "type": "panel",
      "id": "graph",
      "name": "Graph",
      "version": "5.0.0"
    },
    {
      "type": "datasource",
      "id": "prometheus",
      "name": "Prometheus",
      "version": "5.0.0"
    },
    {
      "type": "panel",
      "id": "text",
      "name": "Text",
      "version": "5.0.0"
    }
  ],
  "annotations": {
    "list": [
      {
        "builtIn": 1,
        "datasource": "-- Grafana --",
        "enable": true,
        "hide": true,
        "iconColor": "rgba(0, 211, 255, 1)",
        "name": "Annotations & Alerts",
        "type": "dashboard"
      }
    ]
  },
  "editable": false,
  "gnetId": null,
  "graphTooltip": 1,
  "id": null,
  "links": [],
  "panels": [
    {
      "content": "<center><h2>Deployed Versions</h2></center>",
      "gridPos": {
        "h": 3,
        "w": 24,
        "x": 0,
        "y": 0
      },
      "height": "40",
      "id": 58,
      "links": [],
      "mode": "html",
      "title": "",
      "transparent": true,
      "type": "text"
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 5,
        "w": 24,
        "x": 0,
        "y": 3
      },
      "id": 56,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(istio_build{component=\"pilot\"}) by (tag)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ tag }}",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Pilot Versions",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "content": "<center><h2>Resource Usage</h2></center>",
      "gridPos": {
        "h": 3,
        "w": 24,
        "x": 0,
        "y": 8
      },
      "height": "40",
      "id": 29,
      "links": [],
      "mode": "html",
      "title": "",
      "transparent": true,
      "type": "text"
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 7,
        "w": 6,
        "x": 0,
        "y": 11
      },
      "id": 5,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "process_virtual_memory_bytes{job=\"pilot\"}",
          "format": "time_series",
          "instant": false,
          "intervalFactor": 2,
          "legendFormat": "Virtual Memory",
          "refId": "I",
          "step": 2
        },
        {
          "expr": "process_resident_memory_bytes{job=\"pilot\"}",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "Resident Memory",
          "refId": "H",
          "step": 2
        },
        {
          "expr": "go_memstats_heap_sys_bytes{job=\"pilot\"}",
          "format": "time_series",
          "hide": true,
          "intervalFactor": 2,
          "legendFormat": "heap sys",
          "refId": "A"
        },
        {
          "expr": "go_memstats_heap_alloc_bytes{job=\"pilot\"}",
          "format": "time_series",
          "hide": true,
          "intervalFactor": 2,
          "legendFormat": "heap alloc",
          "refId": "D"
        },
        {
          "expr": "go_memstats_alloc_bytes{job=\"pilot\"}",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "Alloc",
          "refId": "F",
          "step": 2
        },
        {
          "expr": "go_memstats_heap_inuse_bytes{job=\"pilot\"}",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 2,
          "legendFormat": "Heap in-use",
          "refId": "E",
          "step": 2
        },
        {
          "expr": "go_memstats_stack_inuse_bytes{job=\"pilot\"}",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "Stack in-use",
          "refId": "G",
          "step": 2
        },
        {
          "expr": "sum(container_memory_usage_bytes{container_name=~\"discovery|istio-proxy\", pod_name=~\"istio-pilot-.*\"})",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 2,
          "legendFormat": "Total (k8s)",
          "refId": "C",
          "step": 2
        },
        {
          "expr": "container_memory_usage_bytes{container_name=~\"discovery|istio-proxy\", pod_name=~\"istio-pilot-.*\"}",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 2,
          "legendFormat": "{{ container_name }} (k8s)",
          "refId": "B",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Memory",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "bytes",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 7,
        "w": 6,
        "x": 6,
        "y": 11
      },
      "id": 6,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(rate(container_cpu_usage_seconds_total{container_name=~\"discovery|istio-proxy\", pod_name=~\"istio-pilot-.*\"}[1m]))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 2,
          "legendFormat": "Total (k8s)",
          "refId": "A",
          "step": 2
        },
        {
          "expr": "sum(rate(container_cpu_usage_seconds_total{container_name=~\"discovery|istio-proxy\", pod_name=~\"istio-pilot-.*\"}[1m])) by (container_name)",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 2,
          "legendFormat": "{{ container_name }} (k8s)",
          "refId": "B",
          "step": 2
        },
        {
          "expr": "irate(process_cpu_seconds_total{job=\"pilot\"}[1m])",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 2,
          "legendFormat": "pilot (self-reported)",
          "refId": "C",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "CPU",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 7,
        "w": 6,
        "x": 12,
        "y": 11
      },
      "id": 7,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "process_open_fds{job=\"pilot\"}",
          "format": "time_series",
          "hide": true,
          "instant": false,
          "interval": "",
          "intervalFactor": 2,
          "legendFormat": "Open FDs (pilot)",
          "refId": "A"
        },
        {
          "expr": "container_fs_usage_bytes{container_name=~\"discovery|istio-proxy\", pod_name=~\"istio-pilot-.*\"}",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "{{ container_name }}",
          "refId": "B",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Disk",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "bytes",
          "label": "",
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "decimals": null,
          "format": "none",
          "label": "",
          "logBase": 1024,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 7,
        "w": 6,
        "x": 18,
        "y": 11
      },
      "id": 4,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": false,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "go_goroutines{job=\"pilot\"}",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "Number of Goroutines",
          "refId": "A",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Goroutines",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": "",
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "content": "<center><h2>xDS</h2></center>",
      "gridPos": {
        "h": 3,
        "w": 24,
        "x": 0,
        "y": 18
      },
      "id": 28,
      "links": [],
      "mode": "html",
      "title": "",
      "transparent": true,
      "type": "text"
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 0,
        "y": 21
      },
      "id": 40,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(irate(envoy_cluster_update_success{cluster_name=\"xds-grpc\"}[1m]))",
          "format": "time_series",
          "hide": false,
          "intervalFactor": 1,
          "legendFormat": "XDS GRPC Successes",
          "refId": "C"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Updates",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "ops",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "ops",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 8,
        "y": 21
      },
      "id": 42,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "round(sum(rate(envoy_cluster_update_attempt{cluster_name=\"xds-grpc\"}[1m])) - sum(rate(envoy_cluster_update_success{cluster_name=\"xds-grpc\"}[1m])))",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "XDS GRPC ",
          "refId": "A",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Failures",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "ops",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": false
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 6,
        "w": 8,
        "x": 16,
        "y": 21
      },
      "id": 41,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(envoy_cluster_upstream_cx_active{cluster_name=\"xds-grpc\"})",
          "format": "time_series",
          "intervalFactor": 2,
          "legendFormat": "Pilot (XDS GRPC)",
          "refId": "C",
          "step": 2
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Active Connections",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 8,
        "w": 8,
        "x": 0,
        "y": 27
      },
      "id": 45,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "pilot_conflict_inbound_listener{job=\"pilot\"}",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Inbound Listeners",
          "refId": "B"
        },
        {
          "expr": "pilot_conflict_outbound_listener_http_over_current_tcp{job=\"pilot\"}",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Outbound Listeners (http over current tcp)",
          "refId": "A"
        },
        {
          "expr": "pilot_conflict_outbound_listener_tcp_over_current_tcp{job=\"pilot\"}",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Outbound Listeners (tcp over current tcp)",
          "refId": "C"
        },
        {
          "expr": "pilot_conflict_outbound_listener_tcp_over_current_http{job=\"pilot\"}",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Outbound Listeners (tcp over current http)",
          "refId": "D"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Conflicts",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 8,
        "w": 8,
        "x": 8,
        "y": 27
      },
      "id": 47,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "pilot_virt_services{job=\"pilot\"}",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Virtual Services",
          "refId": "A"
        },
        {
          "expr": "pilot_services{job=\"pilot\"}",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Services",
          "refId": "B"
        },
        {
          "expr": "label_replace(sum(pilot_xds_cds_reject{job=\"pilot\"}) by (node, err), \"node\", \"$1\", \"node\", \".*~.*~(.*)~.*\")",
          "format": "time_series",
          "hide": true,
          "intervalFactor": 1,
          "legendFormat": "Rejected CDS Configs - {{ node }}: {{ err }}",
          "refId": "C"
        },
        {
          "expr": "pilot_xds_eds_reject{job=\"pilot\"}",
          "format": "time_series",
          "hide": true,
          "intervalFactor": 1,
          "legendFormat": "Rejected EDS Configs",
          "refId": "D"
        },
        {
          "expr": "pilot_xds{job=\"pilot\"}",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Connected Endpoints",
          "refId": "E"
        },
        {
          "expr": "rate(pilot_xds_write_timeout{job=\"pilot\"}[1m])",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Write Timeouts",
          "refId": "F"
        },
        {
          "expr": "rate(pilot_xds_push_timeout{job=\"pilot\"}[1m])",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Push Timeouts",
          "refId": "G"
        },
        {
          "expr": "rate(pilot_xds_pushes{job=\"pilot\"}[1m])",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Pushes ({{ type }})",
          "refId": "H"
        },
        {
          "expr": "rate(pilot_xds_push_errors{job=\"pilot\"}[1m])",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "Push Errors ({{ type }})",
          "refId": "I"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "ADS Monitoring",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 8,
        "w": 8,
        "x": 16,
        "y": 27
      },
      "id": 49,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "label_replace(sum(pilot_xds_cds_reject{job=\"pilot\"}) by (node, err), \"node\", \"$1\", \"node\", \".*~.*~(.*)~.*\")",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ node }}  ({{ err }})",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Rejected CDS Configs",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 7,
        "w": 8,
        "x": 0,
        "y": 35
      },
      "id": 52,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "label_replace(sum(pilot_xds_eds_reject{job=\"pilot\"}) by (node, err), \"node\", \"$1\", \"node\", \".*~.*~(.*)~.*\")",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ node }} ({{err}})",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Rejected EDS Configs",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 7,
        "w": 8,
        "x": 8,
        "y": 35
      },
      "id": 54,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "label_replace(sum(pilot_xds_lds_reject{job=\"pilot\"}) by (node, err), \"node\", \"$1\", \"node\", \".*~.*~(.*)~.*\")",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ node }} ({{err}})",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Rejected LDS Configs",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 7,
        "w": 8,
        "x": 16,
        "y": 35
      },
      "id": 53,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "label_replace(sum(pilot_xds_rds_reject{job=\"pilot\"}) by (node, err), \"node\", \"$1\", \"node\", \".*~.*~(.*)~.*\")",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ node }} ({{err}})",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "Rejected RDS Configs",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "aliasColors": {
        "outbound|80||default-http-backend.kube-system.svc.cluster.local": "rgba(255, 255, 255, 0.97)"
      },
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "Prometheus",
      "fill": 1,
      "gridPos": {
        "h": 7,
        "w": 8,
        "x": 0,
        "y": 42
      },
      "id": 51,
      "legend": {
        "avg": false,
        "current": false,
        "max": false,
        "min": false,
        "show": true,
        "total": false,
        "values": false
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [
        {
          "alias": "outbound|80||default-http-backend.kube-system.svc.cluster.local",
          "yaxis": 1
        }
      ],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "expr": "sum(pilot_xds_eds_instances{job=\"pilot\"}) by (cluster)",
          "format": "time_series",
          "intervalFactor": 1,
          "legendFormat": "{{ cluster }}",
          "refId": "A"
        }
      ],
      "thresholds": [],
      "timeFrom": null,
      "timeShift": null,
      "title": "EDS Instances",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    }
  ],
  "refresh": "5s",
  "schemaVersion": 16,
  "style": "dark",
  "tags": [],
  "templating": {
    "list": []
  },
  "time": {
    "from": "now-5m",
    "to": "now"
  },
  "timepicker": {
    "refresh_intervals": [
      "5s",
      "10s",
      "30s",
      "1m",
      "5m",
      "15m",
      "30m",
      "1h",
      "2h",
      "1d"
    ],
    "time_options": [
      "5m",
      "15m",
      "1h",
      "6h",
      "12h",
      "24h",
      "2d",
      "7d",
      "30d"
    ]
  },
  "timezone": "browser",
  "title": "Istio Pilot Dashboard",
  "version": 4
}
'
---

---
# Source: istio/charts/grafana/templates/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-grafana
  labels:
    app: grafana
    chart: grafana
    heritage: Tiller
    release: istio
    istio: grafana
data:
  datasources.yaml: |
    apiVersion: 1
    datasources:
    - access: proxy
      editable: true
      isDefault: true
      jsonData:
        timeInterval: 5s
      name: Prometheus
      orgId: 1
      type: prometheus
      url: http://prometheus:9090
    
  dashboardproviders.yaml: |
    apiVersion: 1
    providers:
    - disableDeletion: false
      folder: istio
      name: istio
      options:
        path: /var/lib/grafana/dashboards/istio
      orgId: 1
      type: file
    
---
# Source: istio/charts/kiali/templates/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kiali
  labels:
    app: kiali
    chart: kiali
    heritage: Tiller
    release: istio
data:
  config.yaml: |
    istio_namespace: istio-system
    server:
      port: 20001
    external_services:
      istio:
        url_service_version: http://istio-pilot:8080/version
      jaeger:
        url: 
      grafana:
        url: 

---
# Source: istio/charts/prometheus/templates/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus
  labels:
    app: prometheus
    chart: prometheus
    heritage: Tiller
    release: istio
data:
  prometheus.yml: |-
    global:
      scrape_interval: 15s
    scrape_configs:

    - job_name: 'istio-mesh'
      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - istio-system

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: istio-telemetry;prometheus

    # Scrape config for envoy stats
    - job_name: 'envoy-stats'
      metrics_path: /stats/prometheus
      kubernetes_sd_configs:
      - role: pod

      relabel_configs:
      - source_labels: [__meta_kubernetes_pod_container_port_name]
        action: keep
        regex: '.*-envoy-prom'
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:15090
        target_label: __address__
      - action: labelmap
        regex: __meta_kubernetes_pod_label_(.+)
      - source_labels: [__meta_kubernetes_namespace]
        action: replace
        target_label: namespace
      - source_labels: [__meta_kubernetes_pod_name]
        action: replace
        target_label: pod_name

      metric_relabel_configs:
      # Exclude some of the envoy metrics that have massive cardinality
      # This list may need to be pruned further moving forward, as informed
      # by performance and scalability testing.
      - source_labels: [ cluster_name ]
        regex: '(outbound|inbound|prometheus_stats).*'
        action: drop
      - source_labels: [ tcp_prefix ]
        regex: '(outbound|inbound|prometheus_stats).*'
        action: drop
      - source_labels: [ listener_address ]
        regex: '(.+)'
        action: drop
      - source_labels: [ http_conn_manager_listener_prefix ]
        regex: '(.+)'
        action: drop
      - source_labels: [ http_conn_manager_prefix ]
        regex: '(.+)'
        action: drop
      - source_labels: [ __name__ ]
        regex: 'envoy_tls.*'
        action: drop
      - source_labels: [ __name__ ]
        regex: 'envoy_tcp_downstream.*'
        action: drop
      - source_labels: [ __name__ ]
        regex: 'envoy_http_(stats|admin).*'
        action: drop
      - source_labels: [ __name__ ]
        regex: 'envoy_cluster_(lb|retry|bind|internal|max|original).*'
        action: drop

    - job_name: 'istio-policy'
      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - istio-system


      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: istio-policy;http-monitoring

    - job_name: 'istio-telemetry'
      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - istio-system

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: istio-telemetry;http-monitoring

    - job_name: 'pilot'
      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - istio-system

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: istio-pilot;http-monitoring

    - job_name: 'galley'
      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - istio-system

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: istio-galley;http-monitoring

    - job_name: 'citadel'
      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - istio-system

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: istio-citadel;http-monitoring

    # scrape config for API servers
    - job_name: 'kubernetes-apiservers'
      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - default
      scheme: https
      tls_config:
        ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
      bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: kubernetes;https

    # scrape config for nodes (kubelet)
    - job_name: 'kubernetes-nodes'
      scheme: https
      tls_config:
        ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
      bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
      kubernetes_sd_configs:
      - role: node
      relabel_configs:
      - action: labelmap
        regex: __meta_kubernetes_node_label_(.+)
      - target_label: __address__
        replacement: kubernetes.default.svc:443
      - source_labels: [__meta_kubernetes_node_name]
        regex: (.+)
        target_label: __metrics_path__
        replacement: /api/v1/nodes/${1}/proxy/metrics

    # Scrape config for Kubelet cAdvisor.
    #
    # This is required for Kubernetes 1.7.3 and later, where cAdvisor metrics
    # (those whose names begin with 'container_') have been removed from the
    # Kubelet metrics endpoint.  This job scrapes the cAdvisor endpoint to
    # retrieve those metrics.
    #
    # In Kubernetes 1.7.0-1.7.2, these metrics are only exposed on the cAdvisor
    # HTTP endpoint; use "replacement: /api/v1/nodes/${1}:4194/proxy/metrics"
    # in that case (and ensure cAdvisor's HTTP server hasn't been disabled with
    # the --cadvisor-port=0 Kubelet flag).
    #
    # This job is not necessary and should be removed in Kubernetes 1.6 and
    # earlier versions, or it will cause the metrics to be scraped twice.
    - job_name: 'kubernetes-cadvisor'
      scheme: https
      tls_config:
        ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
      bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
      kubernetes_sd_configs:
      - role: node
      relabel_configs:
      - action: labelmap
        regex: __meta_kubernetes_node_label_(.+)
      - target_label: __address__
        replacement: kubernetes.default.svc:443
      - source_labels: [__meta_kubernetes_node_name]
        regex: (.+)
        target_label: __metrics_path__
        replacement: /api/v1/nodes/${1}/proxy/metrics/cadvisor

    # scrape config for service endpoints.
    - job_name: 'kubernetes-service-endpoints'
      kubernetes_sd_configs:
      - role: endpoints
      relabel_configs:
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
        action: replace
        target_label: __scheme__
        regex: (https?)
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
        action: replace
        target_label: __address__
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
      - action: labelmap
        regex: __meta_kubernetes_service_label_(.+)
      - source_labels: [__meta_kubernetes_namespace]
        action: replace
        target_label: kubernetes_namespace
      - source_labels: [__meta_kubernetes_service_name]
        action: replace
        target_label: kubernetes_name

    - job_name: 'kubernetes-pods'
      kubernetes_sd_configs:
      - role: pod
      relabel_configs:  # If first two labels are present, pod should be scraped  by the istio-secure job.
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      # Keep target if there's no sidecar or if prometheus.io/scheme is explicitly set to "http"
      - source_labels: [__meta_kubernetes_pod_annotation_sidecar_istio_io_status, __meta_kubernetes_pod_annotation_prometheus_io_scheme]
        action: keep
        regex: ((;.*)|(.*;http))
      - source_labels: [__meta_kubernetes_pod_annotation_istio_mtls]
        action: drop
        regex: (true)
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target_label: __address__
      - action: labelmap
        regex: __meta_kubernetes_pod_label_(.+)
      - source_labels: [__meta_kubernetes_namespace]
        action: replace
        target_label: namespace
      - source_labels: [__meta_kubernetes_pod_name]
        action: replace
        target_label: pod_name

    - job_name: 'kubernetes-pods-istio-secure'
      scheme: https
      tls_config:
        ca_file: /etc/istio-certs/root-cert.pem
        cert_file: /etc/istio-certs/cert-chain.pem
        key_file: /etc/istio-certs/key.pem
        insecure_skip_verify: true  # prometheus does not support secure naming.
      kubernetes_sd_configs:
      - role: pod
      relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      # sidecar status annotation is added by sidecar injector and
      # istio_workload_mtls_ability can be specifically placed on a pod to indicate its ability to receive mtls traffic.
      - source_labels: [__meta_kubernetes_pod_annotation_sidecar_istio_io_status, __meta_kubernetes_pod_annotation_istio_mtls]
        action: keep
        regex: (([^;]+);([^;]*))|(([^;]*);(true))
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scheme]
        action: drop
        regex: (http)
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__]  # Only keep address that is host:port
        action: keep    # otherwise an extra target with ':443' is added for https scheme
        regex: ([^:]+):(\d+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target_label: __address__
      - action: labelmap
        regex: __meta_kubernetes_pod_label_(.+)
      - source_labels: [__meta_kubernetes_namespace]
        action: replace
        target_label: namespace
      - source_labels: [__meta_kubernetes_pod_name]
        action: replace
        target_label: pod_name
---
# Source: istio/charts/security/templates/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-security-custom-resources
  labels:
    app: security
    chart: security
    heritage: Tiller
    release: istio
    istio: citadel
data:
  custom-resources.yaml: |-    
    # Authentication policy to enable permissive mode for all services (that have sidecar) in the mesh.
    apiVersion: "authentication.istio.io/v1alpha1"
    kind: "MeshPolicy"
    metadata:
      name: "default"
      labels:
        app: security
        chart: security
        heritage: Tiller
        release: istio
    spec:
      peers:
      - mtls:
          mode: PERMISSIVE
  run.sh: |-    
    #!/bin/sh
    
    set -x
    
    if [ "$#" -ne "1" ]; then
        echo "first argument should be path to custom resource yaml"
        exit 1
    fi
    
    pathToResourceYAML=${1}
    
    kubectl get validatingwebhookconfiguration istio-galley 2>/dev/null
    if [ "$?" -eq 0 ]; then
        echo "istio-galley validatingwebhookconfiguration found - waiting for istio-galley deployment to be ready"
        while true; do
            kubectl -n istio-system get deployment istio-galley 2>/dev/null
            if [ "$?" -eq 0 ]; then
                break
            fi
            sleep 1
        done
        kubectl -n istio-system rollout status deployment istio-galley
        if [ "$?" -ne 0 ]; then
            echo "istio-galley deployment rollout status check failed"
            exit 1
        fi
        echo "istio-galley deployment ready for configuration validation"
    fi
    sleep 5
    kubectl apply -f ${pathToResourceYAML}
    

---
# Source: istio/templates/configmap.yaml

apiVersion: v1
kind: ConfigMap
metadata:
  name: istio
  labels:
    app: istio
    chart: istio
    heritage: Tiller
    release: istio
data:
  mesh: |-
    # Set the following variable to true to disable policy checks by the Mixer.
    # Note that metrics will still be reported to the Mixer.
    disablePolicyChecks: false

    # Set enableTracing to false to disable request tracing.
    enableTracing: true

    # Set accessLogFile to empty string to disable access log.
    accessLogFile: "/dev/stdout"

    # If accessLogEncoding is TEXT, value will be used directly as the log format
    # example: "[%START_TIME%] %REQ(:METHOD)% %REQ(X-ENVOY-ORIGINAL-PATH?:PATH)% %PROTOCOL%\n"
    # If AccessLogEncoding is JSON, value will be parsed as map[string]string
    # example: '{"start_time": "%START_TIME%", "req_method": "%REQ(:METHOD)%"}'
    # Leave empty to use default log format
    accessLogFormat: ""

    # Set accessLogEncoding to JSON or TEXT to configure sidecar access log
    accessLogEncoding: 'TEXT'
    mixerCheckServer: istio-policy.istio-system.svc.cluster.local:9091
    mixerReportServer: istio-telemetry.istio-system.svc.cluster.local:9091
    # policyCheckFailOpen allows traffic in cases when the mixer policy service cannot be reached.
    # Default is false which means the traffic is denied when the client is unable to connect to Mixer.
    policyCheckFailOpen: false
    # Let Pilot give ingresses the public IP of the Istio ingressgateway
    ingressService: istio-ingressgateway

    # Default connect timeout for dynamic clusters generated by Pilot and returned via XDS
    connectTimeout: 10s
    
    # DNS refresh rate for Envoy clusters of type STRICT_DNS
    dnsRefreshRate: 5s

    # Unix Domain Socket through which envoy communicates with NodeAgent SDS to get
    # key/cert for mTLS. Use secret-mount files instead of SDS if set to empty. 
    sdsUdsPath: 

    # This flag is used by secret discovery service(SDS). 
    # If set to true(prerequisite: https://kubernetes.io/docs/concepts/storage/volumes/#projected), Istio will inject volumes mount 
    # for k8s service account JWT, so that K8s API server mounts k8s service account JWT to envoy container, which 
    # will be used to generate key/cert eventually. This isn't supported for non-k8s case.
    enableSdsTokenMount: false

    # This flag is used by secret discovery service(SDS). 
    # If set to true, envoy will fetch normal k8s service account JWT from '/var/run/secrets/kubernetes.io/serviceaccount/token' 
    # (https://kubernetes.io/docs/tasks/access-application-cluster/access-cluster/#accessing-the-api-from-a-pod) 
    # and pass to sds server, which will be used to request key/cert eventually. 
    # this flag is ignored if enableSdsTokenMount is set.
    # This isn't supported for non-k8s case.
    sdsUseK8sSaJwt: false

    # The trust domain corresponds to the trust root of a system.
    # Refer to https://github.com/spiffe/spiffe/blob/master/standards/SPIFFE-ID.md#21-trust-domain
    trustDomain: 

    # Set the default behavior of the sidecar for handling outbound traffic from the application:
    # ALLOW_ANY - outbound traffic to unknown destinations will be allowed, in case there are no
    #   services or ServiceEntries for the destination port
    # REGISTRY_ONLY - restrict outbound traffic to services defined in the service registry as well
    #   as those defined through ServiceEntries  
    outboundTrafficPolicy:
      mode: ALLOW_ANY

    localityLbSetting:
      {}
      

    # The namespace to treat as the administrative root namespace for istio
    # configuration.    
    rootNamespace: istio-system
    configSources:
    - address: istio-galley.istio-system.svc:9901

    defaultConfig:
      #
      # TCP connection timeout between Envoy & the application, and between Envoys.  Used for static clusters
      # defined in Envoy's configuration file
      connectTimeout: 10s
      #
      ### ADVANCED SETTINGS #############
      # Where should envoy's configuration be stored in the istio-proxy container
      configPath: "/etc/istio/proxy"
      binaryPath: "/usr/local/bin/envoy"
      # The pseudo service name used for Envoy.
      serviceCluster: istio-proxy
      # These settings that determine how long an old Envoy
      # process should be kept alive after an occasional reload.
      drainDuration: 45s
      parentShutdownDuration: 1m0s
      #
      # The mode used to redirect inbound connections to Envoy. This setting
      # has no effect on outbound traffic: iptables REDIRECT is always used for
      # outbound connections.
      # If "REDIRECT", use iptables REDIRECT to NAT and redirect to Envoy.
      # The "REDIRECT" mode loses source addresses during redirection.
      # If "TPROXY", use iptables TPROXY to redirect to Envoy.
      # The "TPROXY" mode preserves both the source and destination IP
      # addresses and ports, so that they can be used for advanced filtering
      # and manipulation.
      # The "TPROXY" mode also configures the sidecar to run with the
      # CAP_NET_ADMIN capability, which is required to use TPROXY.
      #interceptionMode: REDIRECT
      #
      # Port where Envoy listens (on local host) for admin commands
      # You can exec into the istio-proxy container in a pod and
      # curl the admin port (curl http://localhost:15000/) to obtain
      # diagnostic information from Envoy. See
      # https://lyft.github.io/envoy/docs/operations/admin.html
      # for more details
      proxyAdminPort: 15000
      #
      # Set concurrency to a specific number to control the number of Proxy worker threads.
      # If set to 0 (default), then start worker thread for each CPU thread/core.
      concurrency: 2
      #
      tracing:
        zipkin:
          # Address of the Zipkin collector
          address: zipkin.istio-system:9411
      #
      # Mutual TLS authentication between sidecars and istio control plane.
      controlPlaneAuthPolicy: NONE
      #
      # Address where istio Pilot service is running
      discoveryAddress: istio-pilot.istio-system:15010
  
  # Configuration file for the mesh networks to be used by the Split Horizon EDS.
  meshNetworks: |-
    networks: {}

---
# Source: istio/templates/sidecar-injector-configmap.yaml

apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-sidecar-injector
  labels:
    app: istio
    chart: istio
    heritage: Tiller
    release: istio
    istio: sidecar-injector
data:
  config: |-
    policy: enabled
    template: |-
      rewriteAppHTTPProbe: false
      initContainers:
      [[ if ne (annotation .ObjectMeta `+"`"+`sidecar.istio.io/interceptionMode`+"`"+` .ProxyConfig.InterceptionMode) "NONE" ]]
      - name: istio-init
        image: "docker.io/istio/proxy_init:1.1.6"
        args:
        - "-p"
        - [[ .MeshConfig.ProxyListenPort ]]
        - "-u"
        - 1337
        - "-m"
        - [[ annotation .ObjectMeta `+"`"+`sidecar.istio.io/interceptionMode`+"`"+` .ProxyConfig.InterceptionMode ]]
        - "-i"
        - "[[ annotation .ObjectMeta `+"`"+`traffic.sidecar.istio.io/includeOutboundIPRanges`+"`"+`  "*"  ]]"
        - "-x"
        - "[[ annotation .ObjectMeta `+"`"+`traffic.sidecar.istio.io/excludeOutboundIPRanges`+"`"+`  ""  ]]"
        - "-b"
        - "[[ annotation .ObjectMeta `+"`"+`traffic.sidecar.istio.io/includeInboundPorts`+"`"+` (includeInboundPorts .Spec.Containers) ]]"
        - "-d"
        - "[[ excludeInboundPort (annotation .ObjectMeta `+"`"+`status.sidecar.istio.io/port`+"`"+`  15020 ) (annotation .ObjectMeta `+"`"+`traffic.sidecar.istio.io/excludeInboundPorts`+"`"+`  "" ) ]]"
        [[ if (isset .ObjectMeta.Annotations `+"`"+`traffic.sidecar.istio.io/kubevirtInterfaces`+"`"+`) -]]
        - "-k"
        - "[[ index .ObjectMeta.Annotations `+"`"+`traffic.sidecar.istio.io/kubevirtInterfaces`+"`"+` ]]"
        [[ end -]]
        imagePullPolicy: IfNotPresent
        resources:
          requests:
            cpu: 10m
            memory: 10Mi
          limits:
            cpu: 100m
            memory: 50Mi
        securityContext:
          runAsUser: 0
          runAsNonRoot: false
          capabilities:
            add:
            - NET_ADMIN
        restartPolicy: Always
      [[ end -]]
      containers:
      - name: istio-proxy
        image: [[ annotation .ObjectMeta `+"`"+`sidecar.istio.io/proxyImage`+"`"+`  "docker.io/istio/proxyv2:1.1.6"  ]]
        ports:
        - containerPort: 15090
          protocol: TCP
          name: http-envoy-prom
        args:
        - proxy
        - sidecar
        - --domain
        - $(POD_NAMESPACE).svc.cluster.local
        - --configPath
        - [[ .ProxyConfig.ConfigPath ]]
        - --binaryPath
        - [[ .ProxyConfig.BinaryPath ]]
        - --serviceCluster
        [[ if ne "" (index .ObjectMeta.Labels "app") -]]
        - [[ index .ObjectMeta.Labels "app" ]].$(POD_NAMESPACE)
        [[ else -]]
        - [[ valueOrDefault .DeploymentMeta.Name "istio-proxy" ]].[[ valueOrDefault .DeploymentMeta.Namespace "default" ]]
        [[ end -]]
        - --drainDuration
        - [[ formatDuration .ProxyConfig.DrainDuration ]]
        - --parentShutdownDuration
        - [[ formatDuration .ProxyConfig.ParentShutdownDuration ]]
        - --discoveryAddress
        - [[ annotation .ObjectMeta `+"`"+`sidecar.istio.io/discoveryAddress`+"`"+` .ProxyConfig.DiscoveryAddress ]]
        - --zipkinAddress
        - [[ .ProxyConfig.GetTracing.GetZipkin.GetAddress ]]
        - --connectTimeout
        - [[ formatDuration .ProxyConfig.ConnectTimeout ]]
        - --proxyAdminPort
        - [[ .ProxyConfig.ProxyAdminPort ]]
        [[ if gt .ProxyConfig.Concurrency 0 -]]
        - --concurrency
        - [[ .ProxyConfig.Concurrency ]]
        [[ end -]]
        - --controlPlaneAuthPolicy
        - [[ annotation .ObjectMeta `+"`"+`sidecar.istio.io/controlPlaneAuthPolicy`+"`"+` .ProxyConfig.ControlPlaneAuthPolicy ]]
      [[- if (ne (annotation .ObjectMeta `+"`"+`status.sidecar.istio.io/port`+"`"+`  15020 ) "0") ]]
        - --statusPort
        - [[ annotation .ObjectMeta `+"`"+`status.sidecar.istio.io/port`+"`"+`  15020  ]]
        - --applicationPorts
        - "[[ annotation .ObjectMeta `+"`"+`readiness.status.sidecar.istio.io/applicationPorts`+"`"+` (applicationPorts .Spec.Containers) ]]"
      [[- end ]]
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: INSTANCE_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        
        - name: ISTIO_META_POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: ISTIO_META_CONFIG_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: ISTIO_META_INTERCEPTION_MODE
          value: [[ or (index .ObjectMeta.Annotations "sidecar.istio.io/interceptionMode") .ProxyConfig.InterceptionMode.String ]]
        [[ if .ObjectMeta.Annotations ]]
        - name: ISTIO_METAJSON_ANNOTATIONS
          value: |
                 [[ toJSON .ObjectMeta.Annotations ]]
        [[ end ]]
        [[ if .ObjectMeta.Labels ]]
        - name: ISTIO_METAJSON_LABELS
          value: |
                 [[ toJSON .ObjectMeta.Labels ]]
        [[ end ]]
        [[- if (isset .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/bootstrapOverride`+"`"+`) ]]
        - name: ISTIO_BOOTSTRAP_OVERRIDE
          value: "/etc/istio/custom-bootstrap/custom_bootstrap.json"
        [[- end ]]
        imagePullPolicy: IfNotPresent
        [[ if (ne (annotation .ObjectMeta `+"`"+`status.sidecar.istio.io/port`+"`"+`  15020 ) "0") ]]
        readinessProbe:
          httpGet:
            path: /healthz/ready
            port: [[ annotation .ObjectMeta `+"`"+`status.sidecar.istio.io/port`+"`"+`  15020  ]]
          initialDelaySeconds: [[ annotation .ObjectMeta `+"`"+`readiness.status.sidecar.istio.io/initialDelaySeconds`+"`"+`  1  ]]
          periodSeconds: [[ annotation .ObjectMeta `+"`"+`readiness.status.sidecar.istio.io/periodSeconds`+"`"+`  2  ]]
          failureThreshold: [[ annotation .ObjectMeta `+"`"+`readiness.status.sidecar.istio.io/failureThreshold`+"`"+`  30  ]]
        [[ end -]]securityContext:
          readOnlyRootFilesystem: true
          [[ if eq (annotation .ObjectMeta `+"`"+`sidecar.istio.io/interceptionMode`+"`"+` .ProxyConfig.InterceptionMode) "TPROXY" -]]
          capabilities:
            add:
            - NET_ADMIN
          runAsGroup: 1337
          [[ else -]]
          
          runAsUser: 1337
          [[- end ]]
        resources:
          [[ if or (isset .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/proxyCPU`+"`"+`) (isset .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/proxyMemory`+"`"+`) -]]
          requests:
            [[ if (isset .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/proxyCPU`+"`"+`) -]]
            cpu: "[[ index .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/proxyCPU`+"`"+` ]]"
            [[ end ]]
            [[ if (isset .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/proxyMemory`+"`"+`) -]]
            memory: "[[ index .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/proxyMemory`+"`"+` ]]"
            [[ end ]]
        [[ else -]]
          limits:
            cpu: 2000m
            memory: 128Mi
          requests:
            cpu: 10m
            memory: 40Mi
          
        [[ end -]]
        volumeMounts:
        [[- if (isset .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/bootstrapOverride`+"`"+`) ]]
        - mountPath: /etc/istio/custom-bootstrap
          name: custom-bootstrap-volume
        [[- end ]]
        - mountPath: /etc/istio/proxy
          name: istio-envoy
        - mountPath: /etc/certs/
          name: istio-certs
          readOnly: true
          [[- if isset .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/userVolumeMount`+"`"+` ]]
          [[ range $index, $value := fromJSON (index .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/userVolumeMount`+"`"+`) ]]
        - name: "[[ $index ]]"
          [[ toYaml $value | indent 4 ]]
          [[ end ]]
          [[- end ]]
      volumes:
      [[- if (isset .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/bootstrapOverride`+"`"+`) ]]
      - name: custom-bootstrap-volume
        configMap:
          name: [[ annotation .ObjectMeta `+"`"+`sidecar.istio.io/bootstrapOverride`+"`"+` `+"`"+``+"`"+` ]]
      [[- end ]]
      - emptyDir:
          medium: Memory
        name: istio-envoy
      - name: istio-certs
        secret:
          optional: true
          [[ if eq .Spec.ServiceAccountName "" -]]
          secretName: istio.default
          [[ else -]]
          secretName: [[ printf "istio.%s" .Spec.ServiceAccountName ]]
          [[ end -]]
        [[- if isset .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/userVolume`+"`"+` ]]
        [[ range $index, $value := fromJSON (index .ObjectMeta.Annotations `+"`"+`sidecar.istio.io/userVolume`+"`"+`) ]]
      - name: "[[ $index ]]"
        [[ toYaml $value | indent 2 ]]
        [[ end ]]
        [[ end ]]

---
# Source: istio/charts/galley/templates/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-galley-service-account
  labels:
    app: galley
    chart: galley
    heritage: Tiller
    release: istio

---
# Source: istio/charts/gateways/templates/serviceaccount.yaml

apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-egressgateway-service-account
  labels:
    app: istio-egressgateway
    chart: gateways
    heritage: Tiller
    release: istio
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-ingressgateway-service-account
  labels:
    app: istio-ingressgateway
    chart: gateways
    heritage: Tiller
    release: istio
---


---
# Source: istio/charts/grafana/templates/create-custom-resources-job.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-grafana-post-install-account
  labels:
    app: grafana
    chart: grafana
    heritage: Tiller
    release: istio
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: istio-grafana-post-install-istio-system
  labels:
    app: grafana
    chart: grafana
    heritage: Tiller
    release: istio
rules:
- apiGroups: ["authentication.istio.io"] # needed to create default authn policy
  resources: ["*"]
  verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: istio-grafana-post-install-role-binding-istio-system
  labels:
    app: grafana
    chart: grafana
    heritage: Tiller
    release: istio
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-grafana-post-install-istio-system
subjects:
  - kind: ServiceAccount
    name: istio-grafana-post-install-account
    namespace: istio-system
---
apiVersion: batch/v1
kind: Job
metadata:
  name: istio-grafana-post-install-1.1.6
  annotations:
    "helm.sh/hook": post-install
    "helm.sh/hook-delete-policy": hook-succeeded
  labels:
    app: grafana
    chart: grafana
    heritage: Tiller
    release: istio
spec:
  template:
    metadata:
      name: istio-grafana-post-install
      labels:
        app: istio-grafana
        chart: grafana
        heritage: Tiller
        release: istio
    spec:
      serviceAccountName: istio-grafana-post-install-account
      containers:
        - name: kubectl
          image: "docker.io/istio/kubectl:1.1.6"
          command: [ "/bin/bash", "/tmp/grafana/run.sh", "/tmp/grafana/custom-resources.yaml" ]
          volumeMounts:
            - mountPath: "/tmp/grafana"
              name: tmp-configmap-grafana
      volumes:
        - name: tmp-configmap-grafana
          configMap:
            name: istio-grafana-custom-resources
      restartPolicy: OnFailure
      affinity:      
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x      

---
# Source: istio/charts/kiali/templates/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kiali-service-account
  labels:
    app: kiali
    chart: kiali
    heritage: Tiller
    release: istio

---
# Source: istio/charts/mixer/templates/serviceaccount.yaml

apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-mixer-service-account
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio

---
# Source: istio/charts/pilot/templates/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-pilot-service-account
  labels:
    app: pilot
    chart: pilot
    heritage: Tiller
    release: istio

---
# Source: istio/charts/prometheus/templates/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: prometheus
  labels:
    app: prometheus
    chart: prometheus
    heritage: Tiller
    release: istio

---
# Source: istio/charts/security/templates/cleanup-secrets.yaml
# The reason for creating a ServiceAccount and ClusterRole specifically for this
# post-delete hooked job is because the citadel ServiceAccount is being deleted
# before this hook is launched. On the other hand, running this hook before the
# deletion of the citadel (e.g. pre-delete) won't delete the secrets because they
# will be re-created immediately by the to-be-deleted citadel.
#
# It's also important that the ServiceAccount, ClusterRole and ClusterRoleBinding
# will be ready before running the hooked Job therefore the hook weights.

apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-cleanup-secrets-service-account
  annotations:
    "helm.sh/hook": post-delete
    "helm.sh/hook-delete-policy": hook-succeeded
    "helm.sh/hook-weight": "1"
  labels:
    app: security
    chart: security
    heritage: Tiller
    release: istio
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: istio-cleanup-secrets-istio-system
  annotations:
    "helm.sh/hook": post-delete
    "helm.sh/hook-delete-policy": hook-succeeded
    "helm.sh/hook-weight": "1"
  labels:
    app: security
    chart: security
    heritage: Tiller
    release: istio
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["list", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: istio-cleanup-secrets-istio-system
  annotations:
    "helm.sh/hook": post-delete
    "helm.sh/hook-delete-policy": hook-succeeded
    "helm.sh/hook-weight": "2"
  labels:
    app: security
    chart: security
    heritage: Tiller
    release: istio
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-cleanup-secrets-istio-system
subjects:
  - kind: ServiceAccount
    name: istio-cleanup-secrets-service-account
    namespace: istio-system
---
apiVersion: batch/v1
kind: Job
metadata:
  name: istio-cleanup-secrets-1.1.6
  annotations:
    "helm.sh/hook": post-delete
    "helm.sh/hook-delete-policy": hook-succeeded
    "helm.sh/hook-weight": "3"
  labels:
    app: security
    chart: security
    heritage: Tiller
    release: istio
spec:
  template:
    metadata:
      name: istio-cleanup-secrets
      labels:
        app: security
        chart: security
        heritage: Tiller
        release: istio
    spec:
      serviceAccountName: istio-cleanup-secrets-service-account
      containers:
        - name: kubectl
          image: "docker.io/istio/kubectl:1.1.6"
          imagePullPolicy: IfNotPresent
          command:
          - /bin/bash
          - -c
          - >
              kubectl get secret --all-namespaces | grep "istio.io/key-and-cert" |  while read -r entry; do
                ns=$(echo $entry | awk '{print $1}');
                name=$(echo $entry | awk '{print $2}');
                kubectl delete secret $name -n $ns;
              done
      restartPolicy: OnFailure
      affinity:      
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x      

---
# Source: istio/charts/security/templates/create-custom-resources-job.yaml

apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-security-post-install-account
  labels:
    app: security
    chart: security
    heritage: Tiller
    release: istio
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: istio-security-post-install-istio-system
  labels:
    app: security
    chart: security
    heritage: Tiller
    release: istio
rules:
- apiGroups: ["authentication.istio.io"] # needed to create default authn policy
  resources: ["*"]
  verbs: ["*"]
- apiGroups: ["networking.istio.io"] # needed to create security destination rules
  resources: ["*"]
  verbs: ["*"]
- apiGroups: ["admissionregistration.k8s.io"]
  resources: ["validatingwebhookconfigurations"]
  verbs: ["get"]
- apiGroups: ["extensions", "apps"]
  resources: ["deployments", "replicasets"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: istio-security-post-install-role-binding-istio-system
  labels:
    app: security
    chart: security
    heritage: Tiller
    release: istio
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-security-post-install-istio-system
subjects:
  - kind: ServiceAccount
    name: istio-security-post-install-account
    namespace: istio-system
---
apiVersion: batch/v1
kind: Job
metadata:
  name: istio-security-post-install-1.1.6
  annotations:
    "helm.sh/hook": post-install
    "helm.sh/hook-delete-policy": hook-succeeded
  labels:
    app: security
    chart: security
    heritage: Tiller
    release: istio
spec:
  template:
    metadata:
      name: istio-security-post-install
      labels:
        app: security
        chart: security
        heritage: Tiller
        release: istio
    spec:
      serviceAccountName: istio-security-post-install-account
      containers:
        - name: kubectl
          image: "docker.io/istio/kubectl:1.1.6"
          imagePullPolicy: IfNotPresent
          command: [ "/bin/bash", "/tmp/security/run.sh", "/tmp/security/custom-resources.yaml" ]
          volumeMounts:
            - mountPath: "/tmp/security"
              name: tmp-configmap-security
      volumes:
        - name: tmp-configmap-security
          configMap:
            name: istio-security-custom-resources
      restartPolicy: OnFailure
      affinity:      
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x      

---
# Source: istio/charts/security/templates/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-citadel-service-account
  labels:
    app: security
    chart: security
    heritage: Tiller
    release: istio

---
# Source: istio/charts/sidecarInjectorWebhook/templates/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-sidecar-injector-service-account
  labels:
    app: sidecarInjectorWebhook
    chart: sidecarInjectorWebhook
    heritage: Tiller
    release: istio
    istio: sidecar-injector

---
# Source: istio/templates/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: istio-multi

---
# Source: istio/charts/galley/templates/clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: istio-galley-istio-system
  labels:
    app: galley
    chart: galley
    heritage: Tiller
    release: istio
rules:
- apiGroups: ["admissionregistration.k8s.io"]
  resources: ["validatingwebhookconfigurations"]
  verbs: ["*"]
- apiGroups: ["config.istio.io"] # istio mixer CRD watcher
  resources: ["*"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["networking.istio.io"]
  resources: ["*"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["authentication.istio.io"]
  resources: ["*"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["rbac.istio.io"]
  resources: ["*"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["extensions","apps"]
  resources: ["deployments"]
  resourceNames: ["istio-galley"]
  verbs: ["get"]
- apiGroups: [""]
  resources: ["pods", "nodes", "services", "endpoints"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["extensions"]
  resources: ["ingresses"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["extensions", "apps"]
  resources: ["deployments/finalizers"]
  resourceNames: ["istio-galley"]
  verbs: ["update"]

---
# Source: istio/charts/gateways/templates/clusterrole.yaml

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: istio-egressgateway-istio-system
  labels:
    app: egressgateway
    chart: gateways
    heritage: Tiller
    release: istio
rules:
- apiGroups: ["networking.istio.io"]
  resources: ["virtualservices", "destinationrules", "gateways"]
  verbs: ["get", "watch", "list", "update"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: istio-ingressgateway-istio-system
  labels:
    app: ingressgateway
    chart: gateways
    heritage: Tiller
    release: istio
rules:
- apiGroups: ["networking.istio.io"]
  resources: ["virtualservices", "destinationrules", "gateways"]
  verbs: ["get", "watch", "list", "update"]
---

---
# Source: istio/charts/kiali/templates/clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kiali
  labels:
    app: kiali
    chart: kiali
    heritage: Tiller
    release: istio
rules:
- apiGroups: [""]
  resources:
  - configmaps
  - endpoints
  - namespaces
  - nodes
  - pods
  - services
  - replicationcontrollers
  verbs:
  - get
  - list
  - watch
- apiGroups: ["extensions", "apps"]
  resources:
  - deployments
  - statefulsets
  - replicasets
  verbs:
  - get
  - list
  - watch
- apiGroups: ["autoscaling"]
  resources:
  - horizontalpodautoscalers
  verbs:
  - get
  - list
  - watch
- apiGroups: ["batch"]
  resources:
  - cronjobs
  - jobs
  verbs:
  - get
  - list
  - watch
- apiGroups: ["config.istio.io"]
  resources:
  - apikeys
  - authorizations
  - checknothings
  - circonuses
  - deniers
  - fluentds
  - handlers
  - kubernetesenvs
  - kuberneteses
  - listcheckers
  - listentries
  - logentries
  - memquotas
  - metrics
  - opas
  - prometheuses
  - quotas
  - quotaspecbindings
  - quotaspecs
  - rbacs
  - reportnothings
  - rules
  - solarwindses
  - stackdrivers
  - statsds
  - stdios
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - watch
- apiGroups: ["networking.istio.io"]
  resources:
  - destinationrules
  - gateways
  - serviceentries
  - virtualservices
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - watch
- apiGroups: ["authentication.istio.io"]
  resources:
  - policies
  - meshpolicies
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - watch
- apiGroups: ["rbac.istio.io"]
  resources:
  - clusterrbacconfigs
  - rbacconfigs
  - serviceroles
  - servicerolebindings
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - watch
- apiGroups: ["monitoring.kiali.io"]
  resources:
  - monitoringdashboards
  verbs:
  - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kiali-viewer
  labels:
    app: kiali
    chart: kiali
    heritage: Tiller
    release: istio
rules:
- apiGroups: [""]
  resources:
  - configmaps
  - endpoints
  - namespaces
  - nodes
  - pods
  - services
  - replicationcontrollers
  verbs:
  - get
  - list
  - watch
- apiGroups: ["extensions", "apps"]
  resources:
  - deployments
  - statefulsets
  - replicasets
  verbs:
  - get
  - list
  - watch
- apiGroups: ["autoscaling"]
  resources:
  - horizontalpodautoscalers
  verbs:
  - get
  - list
  - watch
- apiGroups: ["batch"]
  resources:
  - cronjobs
  - jobs
  verbs:
  - get
  - list
  - watch
- apiGroups: ["config.istio.io"]
  resources:
  - apikeys
  - authorizations
  - checknothings
  - circonuses
  - deniers
  - fluentds
  - handlers
  - kubernetesenvs
  - kuberneteses
  - listcheckers
  - listentries
  - logentries
  - memquotas
  - metrics
  - opas
  - prometheuses
  - quotas
  - quotaspecbindings
  - quotaspecs
  - rbacs
  - reportnothings
  - rules
  - servicecontrolreports
  - servicecontrols
  - solarwindses
  - stackdrivers
  - statsds
  - stdios
  verbs:
  - get
  - list
  - watch
- apiGroups: ["networking.istio.io"]
  resources:
  - destinationrules
  - gateways
  - serviceentries
  - virtualservices
  verbs:
  - get
  - list
  - watch
- apiGroups: ["authentication.istio.io"]
  resources:
  - policies
  - meshpolicies
  verbs:
  - get
  - list
  - watch
- apiGroups: ["rbac.istio.io"]
  resources:
  - clusterrbacconfigs
  - rbacconfigs
  - serviceroles
  - servicerolebindings
  verbs:
  - get
  - list
  - watch
- apiGroups: ["monitoring.kiali.io"]
  resources:
  - monitoringdashboards
  verbs:
  - get

---
# Source: istio/charts/mixer/templates/clusterrole.yaml

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: istio-mixer-istio-system
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
rules:
- apiGroups: ["config.istio.io"] # istio CRD watcher
  resources: ["*"]
  verbs: ["create", "get", "list", "watch", "patch"]
- apiGroups: ["apiextensions.k8s.io"]
  resources: ["customresourcedefinitions"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["configmaps", "endpoints", "pods", "services", "namespaces", "secrets", "replicationcontrollers"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["extensions", "apps"]
  resources: ["replicasets"]
  verbs: ["get", "list", "watch"]

---
# Source: istio/charts/pilot/templates/clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: istio-pilot-istio-system
  labels:
    app: pilot
    chart: pilot
    heritage: Tiller
    release: istio
rules:
- apiGroups: ["config.istio.io"]
  resources: ["*"]
  verbs: ["*"]
- apiGroups: ["rbac.istio.io"]
  resources: ["*"]
  verbs: ["get", "watch", "list"]
- apiGroups: ["networking.istio.io"]
  resources: ["*"]
  verbs: ["*"]
- apiGroups: ["authentication.istio.io"]
  resources: ["*"]
  verbs: ["*"]
- apiGroups: ["apiextensions.k8s.io"]
  resources: ["customresourcedefinitions"]
  verbs: ["*"]
- apiGroups: ["extensions", "networking.k8s.io"]
  resources: ["ingresses", "ingresses/status"]
  verbs: ["*"]
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["create", "get", "list", "watch", "update"]
- apiGroups: [""]
  resources: ["endpoints", "pods", "services", "namespaces", "nodes", "secrets"]
  verbs: ["get", "list", "watch"]

---
# Source: istio/charts/prometheus/templates/clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: prometheus-istio-system
  labels:
    app: prometheus
    chart: prometheus
    heritage: Tiller
    release: istio
rules:
- apiGroups: [""]
  resources:
  - nodes
  - services
  - endpoints
  - pods
  - nodes/proxy
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources:
  - configmaps
  verbs: ["get"]
- nonResourceURLs: ["/metrics"]
  verbs: ["get"]

---
# Source: istio/charts/security/templates/clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: istio-citadel-istio-system
  labels:
    app: security
    chart: security
    heritage: Tiller
    release: istio
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["create", "get", "update"]
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["create", "get", "watch", "list", "update", "delete"]
- apiGroups: [""]
  resources: ["serviceaccounts", "services"]
  verbs: ["get", "watch", "list"]
- apiGroups: ["authentication.k8s.io"]
  resources: ["tokenreviews"]
  verbs: ["create"]

---
# Source: istio/charts/sidecarInjectorWebhook/templates/clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: istio-sidecar-injector-istio-system
  labels:
    app: sidecarInjectorWebhook
    chart: sidecarInjectorWebhook
    heritage: Tiller
    release: istio
    istio: sidecar-injector
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["admissionregistration.k8s.io"]
  resources: ["mutatingwebhookconfigurations"]
  verbs: ["get", "list", "watch", "patch"]

---
# Source: istio/templates/clusterrole.yaml
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: istio-reader
rules:
  - apiGroups: ['']
    resources: ['nodes', 'pods', 'services', 'endpoints', "replicationcontrollers"]
    verbs: ['get', 'watch', 'list']
  - apiGroups: ["extensions", "apps"]
    resources: ["replicasets"]
    verbs: ["get", "list", "watch"]

---
# Source: istio/charts/galley/templates/clusterrolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: istio-galley-admin-role-binding-istio-system
  labels:
    app: galley
    chart: galley
    heritage: Tiller
    release: istio
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-galley-istio-system
subjects:
  - kind: ServiceAccount
    name: istio-galley-service-account
    namespace: istio-system

---
# Source: istio/charts/gateways/templates/clusterrolebindings.yaml

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: istio-egressgateway-istio-system
  labels:
    app: egressgateway
    chart: gateways
    heritage: Tiller
    release: istio
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-egressgateway-istio-system
subjects:
- kind: ServiceAccount
  name: istio-egressgateway-service-account
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: istio-ingressgateway-istio-system
  labels:
    app: ingressgateway
    chart: gateways
    heritage: Tiller
    release: istio
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-ingressgateway-istio-system
subjects:
- kind: ServiceAccount
  name: istio-ingressgateway-service-account
---

---
# Source: istio/charts/kiali/templates/clusterrolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: istio-kiali-admin-role-binding-istio-system
  labels:
    app: kiali
    chart: kiali
    heritage: Tiller
    release: istio
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kiali
subjects:
- kind: ServiceAccount
  name: kiali-service-account

---
# Source: istio/charts/mixer/templates/clusterrolebinding.yaml

apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: istio-mixer-admin-role-binding-istio-system
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-mixer-istio-system
subjects:
  - kind: ServiceAccount
    name: istio-mixer-service-account
    namespace: istio-system

---
# Source: istio/charts/pilot/templates/clusterrolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: istio-pilot-istio-system
  labels:
    app: pilot
    chart: pilot
    heritage: Tiller
    release: istio
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-pilot-istio-system
subjects:
  - kind: ServiceAccount
    name: istio-pilot-service-account
    namespace: istio-system

---
# Source: istio/charts/prometheus/templates/clusterrolebindings.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: prometheus-istio-system
  labels:
    app: prometheus
    chart: prometheus
    heritage: Tiller
    release: istio
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prometheus-istio-system
subjects:
- kind: ServiceAccount
  name: prometheus

---
# Source: istio/charts/security/templates/clusterrolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: istio-citadel-istio-system
  labels:
    app: security
    chart: security
    heritage: Tiller
    release: istio
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-citadel-istio-system
subjects:
  - kind: ServiceAccount
    name: istio-citadel-service-account
    namespace: istio-system

---
# Source: istio/charts/sidecarInjectorWebhook/templates/clusterrolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: istio-sidecar-injector-admin-role-binding-istio-system
  labels:
    app: sidecarInjectorWebhook
    chart: sidecarInjectorWebhook
    heritage: Tiller
    release: istio
    istio: sidecar-injector
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-sidecar-injector-istio-system
subjects:
  - kind: ServiceAccount
    name: istio-sidecar-injector-service-account
    namespace: istio-system

---
# Source: istio/templates/clusterrolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: istio-multi
  labels:
    chart: istio-1.1.0
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: istio-reader
subjects:
- kind: ServiceAccount
  name: istio-multi

---
# Source: istio/charts/gateways/templates/role.yaml

apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: istio-ingressgateway-sds
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get", "watch", "list"]
---

---
# Source: istio/charts/gateways/templates/rolebindings.yaml

apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: istio-ingressgateway-sds
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: istio-ingressgateway-sds
subjects:
- kind: ServiceAccount
  name: istio-ingressgateway-service-account
---

---
# Source: istio/charts/galley/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: istio-galley
  labels:
    app: galley
    chart: galley
    heritage: Tiller
    release: istio
    istio: galley
spec:
  ports:
  - port: 443
    name: https-validation
  - port: 15014
    name: http-monitoring
  - port: 9901
    name: grpc-mcp
  selector:
    istio: galley

---
# Source: istio/charts/gateways/templates/service.yaml

apiVersion: v1
kind: Service
metadata:
  name: istio-egressgateway
  labels:
    chart: gateways
    heritage: Tiller
    release: istio
    app: istio-egressgateway
    istio: egressgateway
spec:
  type: ClusterIP
  selector:
    release: istio
    app: istio-egressgateway
    istio: egressgateway
  ports:
    -
      name: http2
      port: 80
    -
      name: https
      port: 443
    -
      name: tls
      port: 15443
      targetPort: 15443
---
apiVersion: v1
kind: Service
metadata:
  name: istio-ingressgateway
  annotations:
    beta.cloud.google.com/backend-config: '{"ports": {"http2":"iap-backendconfig"}}'
  labels:
    chart: gateways
    heritage: Tiller
    release: istio
    app: istio-ingressgateway
    istio: ingressgateway
spec:
  type: NodePort
  selector:
    release: istio
    app: istio-ingressgateway
    istio: ingressgateway
  ports:
    -
      name: status-port
      port: 15020
      targetPort: 15020
    -
      name: http2
      nodePort: 31380
      port: 80
      targetPort: 80
    -
      name: https
      nodePort: 31390
      port: 443
    -
      name: tcp
      nodePort: 31400
      port: 31400
    -
      name: https-kiali
      port: 15029
      targetPort: 15029
    -
      name: https-prometheus
      port: 15030
      targetPort: 15030
    -
      name: https-grafana
      port: 15031
      targetPort: 15031
    -
      name: https-tracing
      port: 15032
      targetPort: 15032
    -
      name: tls
      port: 15443
      targetPort: 15443
---

---
# Source: istio/charts/grafana/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: grafana
  labels:
    app: grafana
    chart: grafana
    heritage: Tiller
    release: istio
spec:
  type: ClusterIP
  ports:
    - port: 3000
      targetPort: 3000
      protocol: TCP
      name: http
  selector:
    app: grafana
  
---
# Source: istio/charts/kiali/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: kiali
  labels:
    app: kiali
    chart: kiali
    heritage: Tiller
    release: istio
spec:
  ports:
  - name: http-kiali
    protocol: TCP
    port: 20001
  selector:
    app: kiali

---
# Source: istio/charts/mixer/templates/service.yaml

apiVersion: v1
kind: Service
metadata:
  name: istio-policy
  annotations:
   networking.istio.io/exportTo: "*"
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
    istio: mixer
spec:
  ports:
  - name: grpc-mixer
    port: 9091
  - name: grpc-mixer-mtls
    port: 15004
  - name: http-monitoring
    port: 15014
  selector:
    istio: mixer
    istio-mixer-type: policy
---
apiVersion: v1
kind: Service
metadata:
  name: istio-telemetry
  annotations:
   networking.istio.io/exportTo: "*"
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
    istio: mixer
spec:
  ports:
  - name: grpc-mixer
    port: 9091
  - name: grpc-mixer-mtls
    port: 15004
  - name: http-monitoring
    port: 15014
  - name: prometheus
    port: 42422
  selector:
    istio: mixer
    istio-mixer-type: telemetry
---


---
# Source: istio/charts/pilot/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: istio-pilot
  labels:
    app: pilot
    chart: pilot
    heritage: Tiller
    release: istio
    istio: pilot
spec:
  ports:
  - port: 15010
    name: grpc-xds # direct
  - port: 15011
    name: https-xds # mTLS
  - port: 8080
    name: http-legacy-discovery # direct
  - port: 15014
    name: http-monitoring
  selector:
    istio: pilot

---
# Source: istio/charts/prometheus/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: prometheus
  annotations:
    prometheus.io/scrape: 'true'
  labels:
    app: prometheus
    chart: prometheus
    heritage: Tiller
    release: istio
spec:
  selector:
    app: prometheus
  ports:
  - name: http-prometheus
    protocol: TCP
    port: 9090

---
# Source: istio/charts/security/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  # we use the normal name here (e.g. 'prometheus')
  # as grafana is configured to use this as a data source
  name: istio-citadel
  labels:
    app: security
    chart: security
    heritage: Tiller
    release: istio
    istio: citadel
spec:
  ports:
    - name: grpc-citadel
      port: 8060
      targetPort: 8060
      protocol: TCP
    - name: http-monitoring
      port: 15014
  selector:
    istio: citadel

---
# Source: istio/charts/sidecarInjectorWebhook/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: istio-sidecar-injector
  labels:
    app: sidecarInjectorWebhook
    chart: sidecarInjectorWebhook
    heritage: Tiller
    release: istio
    istio: sidecar-injector
spec:
  ports:
  - port: 443
  selector:
    istio: sidecar-injector

---
# Source: istio/charts/galley/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: istio-galley
  labels:
    app: galley
    chart: galley
    heritage: Tiller
    release: istio
    istio: galley
spec:
  replicas: 1
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: galley
      chart: galley
      heritage: Tiller
      release: istio      
      istio: galley
  template:
    metadata:
      labels:
        app: galley
        chart: galley
        heritage: Tiller
        release: istio      
        istio: galley
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      serviceAccountName: istio-galley-service-account
      containers:
        - name: galley
          image: "docker.io/istio/galley:1.1.6"
          imagePullPolicy: IfNotPresent
          ports:
          - containerPort: 443
          - containerPort: 15014
          - containerPort: 9901
          command:
          - /usr/local/bin/galley
          - server
          - --meshConfigFile=/etc/mesh-config/mesh
          - --livenessProbeInterval=1s
          - --livenessProbePath=/healthliveness
          - --readinessProbePath=/healthready
          - --readinessProbeInterval=1s
          - --deployment-namespace=istio-system
          - --insecure=true
          - --validation-webhook-config-file
          - /etc/config/validatingwebhookconfiguration.yaml
          - --monitoringPort=15014
          - --log_output_level=default:info
          - --enable-validation=true
          volumeMounts:
          - name: certs
            mountPath: /etc/certs
            readOnly: true
          - name: config
            mountPath: /etc/config
            readOnly: true
          - name: mesh-config
            mountPath: /etc/mesh-config
            readOnly: true
          livenessProbe:
            exec:
              command:
                - /usr/local/bin/galley
                - probe
                - --probe-path=/healthliveness
                - --interval=10s
            initialDelaySeconds: 5
            periodSeconds: 5
          readinessProbe:
            exec:
              command:
                - /usr/local/bin/galley
                - probe
                - --probe-path=/healthready
                - --interval=10s
            initialDelaySeconds: 5
            periodSeconds: 5
          resources:
            requests:
              cpu: 10m
            
      volumes:
      - name: certs
        secret:
          secretName: istio.istio-galley-service-account
      - name: config
        configMap:
          name: istio-galley-configuration
      - name: mesh-config
        configMap:
          name: istio
      affinity:      
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x      

---
# Source: istio/charts/gateways/templates/deployment.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  name: istio-egressgateway
  labels:
    chart: gateways
    heritage: Tiller
    release: istio
    app: istio-egressgateway
    istio: egressgateway
spec:
  selector:
    matchLabels:
      chart: gateways
      heritage: Tiller
      release: istio
      app: istio-egressgateway
      istio: egressgateway
  template:
    metadata:
      labels:
        chart: gateways
        heritage: Tiller
        release: istio
        app: istio-egressgateway
        istio: egressgateway
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      serviceAccountName: istio-egressgateway-service-account
      containers:
        - name: istio-proxy
          image: "docker.io/istio/proxyv2:1.1.6"
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 80
            - containerPort: 443
            - containerPort: 15443
            - containerPort: 15090
              protocol: TCP
              name: http-envoy-prom
          args:
          - proxy
          - router
          - --domain
          - $(POD_NAMESPACE).svc.cluster.local
          - --log_output_level=default:info
          - --drainDuration
          - '45s' #drainDuration
          - --parentShutdownDuration
          - '1m0s' #parentShutdownDuration
          - --connectTimeout
          - '10s' #connectTimeout
          - --serviceCluster
          - istio-egressgateway
          - --zipkinAddress
          - zipkin:9411
          - --proxyAdminPort
          - "15000"
          - --statusPort
          - "15020"
          - --controlPlaneAuthPolicy
          - NONE
          - --discoveryAddress
          - istio-pilot:15010
          readinessProbe:
            failureThreshold: 30
            httpGet:
              path: /healthz/ready
              port: 15020
              scheme: HTTP
            initialDelaySeconds: 1
            periodSeconds: 2
            successThreshold: 1
            timeoutSeconds: 1
          resources:
            limits:
              cpu: 100m
              memory: 128Mi
            requests:
              cpu: 10m
              memory: 40Mi
            
          env:
          - name: POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.namespace
          - name: INSTANCE_IP
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: status.podIP
          - name: HOST_IP
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: status.hostIP
          - name: ISTIO_META_POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: ISTIO_META_CONFIG_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          - name: ISTIO_META_ROUTER_MODE
            value: sni-dnat
          volumeMounts:
          - name: istio-certs
            mountPath: /etc/certs
            readOnly: true
          - name: egressgateway-certs
            mountPath: "/etc/istio/egressgateway-certs"
            readOnly: true
          - name: egressgateway-ca-certs
            mountPath: "/etc/istio/egressgateway-ca-certs"
            readOnly: true
      volumes:
      - name: istio-certs
        secret:
          secretName: istio.istio-egressgateway-service-account
          optional: true
      - name: egressgateway-certs
        secret:
          secretName: "istio-egressgateway-certs"
          optional: true
      - name: egressgateway-ca-certs
        secret:
          secretName: "istio-egressgateway-ca-certs"
          optional: true
      affinity:      
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x      
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: istio-ingressgateway
  labels:
    chart: gateways
    heritage: Tiller
    release: istio
    app: istio-ingressgateway
    istio: ingressgateway
spec:
  selector:
    matchLabels:
      chart: gateways
      heritage: Tiller
      release: istio
      app: istio-ingressgateway
      istio: ingressgateway
  template:
    metadata:
      labels:
        chart: gateways
        heritage: Tiller
        release: istio
        app: istio-ingressgateway
        istio: ingressgateway
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      serviceAccountName: istio-ingressgateway-service-account
      containers:
        - name: istio-proxy
          image: "docker.io/istio/proxyv2:1.1.6"
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 15020
            - containerPort: 80
            - containerPort: 443
            - containerPort: 31400
            - containerPort: 15029
            - containerPort: 15030
            - containerPort: 15031
            - containerPort: 15032
            - containerPort: 15443
            - containerPort: 15090
              protocol: TCP
              name: http-envoy-prom
          args:
          - proxy
          - router
          - --domain
          - $(POD_NAMESPACE).svc.cluster.local
          - --log_output_level=default:info
          - --drainDuration
          - '45s' #drainDuration
          - --parentShutdownDuration
          - '1m0s' #parentShutdownDuration
          - --connectTimeout
          - '10s' #connectTimeout
          - --serviceCluster
          - istio-ingressgateway
          - --zipkinAddress
          - zipkin:9411
          - --proxyAdminPort
          - "15000"
          - --statusPort
          - "15020"
          - --controlPlaneAuthPolicy
          - NONE
          - --discoveryAddress
          - istio-pilot:15010
          readinessProbe:
            failureThreshold: 30
            httpGet:
              path: /healthz/ready
              port: 15020
              scheme: HTTP
            initialDelaySeconds: 1
            periodSeconds: 2
            successThreshold: 1
            timeoutSeconds: 1
          resources:
            limits:
              cpu: 100m
              memory: 128Mi
            requests:
              cpu: 10m
              memory: 40Mi
            
          env:
          - name: POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.namespace
          - name: INSTANCE_IP
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: status.podIP
          - name: HOST_IP
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: status.hostIP
          - name: ISTIO_META_POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: ISTIO_META_CONFIG_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          - name: ISTIO_META_ROUTER_MODE
            value: sni-dnat
          volumeMounts:
          - name: istio-certs
            mountPath: /etc/certs
            readOnly: true
          - name: ingressgateway-certs
            mountPath: "/etc/istio/ingressgateway-certs"
            readOnly: true
          - name: ingressgateway-ca-certs
            mountPath: "/etc/istio/ingressgateway-ca-certs"
            readOnly: true
      volumes:
      - name: istio-certs
        secret:
          secretName: istio.istio-ingressgateway-service-account
          optional: true
      - name: ingressgateway-certs
        secret:
          secretName: "istio-ingressgateway-certs"
          optional: true
      - name: ingressgateway-ca-certs
        secret:
          secretName: "istio-ingressgateway-ca-certs"
          optional: true
      affinity:      
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x      
---

---
# Source: istio/charts/grafana/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
  labels:
    app: grafana
    chart: grafana
    heritage: Tiller
    release: istio
spec:
  replicas: 1
  selector:
    matchLabels:
      app: grafana
      chart: grafana
      heritage: Tiller
      release: istio
  template:
    metadata:
      labels:
        app: grafana
        chart: grafana
        heritage: Tiller
        release: istio
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      securityContext:
        runAsUser: 472
        fsGroup: 472
      containers:
        - name: grafana
          image: "grafana/grafana:6.0.2"
          imagePullPolicy: IfNotPresent
          ports:
          - containerPort: 3000
          readinessProbe:
            httpGet:
              path: /login
              port: 3000
          env:
          - name: GRAFANA_PORT
            value: "3000"
          - name: GF_AUTH_BASIC_ENABLED
            value: "false"
          - name: GF_AUTH_ANONYMOUS_ENABLED
            value: "true"
          - name: GF_AUTH_ANONYMOUS_ORG_ROLE
            value: Admin
          - name: GF_PATHS_DATA
            value: /data/grafana
          resources:
            requests:
              cpu: 10m
            
          volumeMounts:
          - name: data
            mountPath: /data/grafana
          - name: dashboards-istio-galley-dashboard
            mountPath: "/var/lib/grafana/dashboards/istio/galley-dashboard.json"
            subPath: galley-dashboard.json
            readOnly: true
          - name: dashboards-istio-istio-mesh-dashboard
            mountPath: "/var/lib/grafana/dashboards/istio/istio-mesh-dashboard.json"
            subPath: istio-mesh-dashboard.json
            readOnly: true
          - name: dashboards-istio-istio-performance-dashboard
            mountPath: "/var/lib/grafana/dashboards/istio/istio-performance-dashboard.json"
            subPath: istio-performance-dashboard.json
            readOnly: true
          - name: dashboards-istio-istio-service-dashboard
            mountPath: "/var/lib/grafana/dashboards/istio/istio-service-dashboard.json"
            subPath: istio-service-dashboard.json
            readOnly: true
          - name: dashboards-istio-istio-workload-dashboard
            mountPath: "/var/lib/grafana/dashboards/istio/istio-workload-dashboard.json"
            subPath: istio-workload-dashboard.json
            readOnly: true
          - name: dashboards-istio-mixer-dashboard
            mountPath: "/var/lib/grafana/dashboards/istio/mixer-dashboard.json"
            subPath: mixer-dashboard.json
            readOnly: true
          - name: dashboards-istio-pilot-dashboard
            mountPath: "/var/lib/grafana/dashboards/istio/pilot-dashboard.json"
            subPath: pilot-dashboard.json
            readOnly: true
          - name: config
            mountPath: "/etc/grafana/provisioning/datasources/datasources.yaml"
            subPath: datasources.yaml
          - name: config
            mountPath: "/etc/grafana/provisioning/dashboards/dashboardproviders.yaml"
            subPath: dashboardproviders.yaml
      affinity:      
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x      
      volumes:
      - name: config
        configMap:
          name: istio-grafana
      - name: data
        emptyDir: {}
      - name: dashboards-istio-galley-dashboard
        configMap:
          name:  istio-grafana-configuration-dashboards-galley-dashboard
      - name: dashboards-istio-istio-mesh-dashboard
        configMap:
          name:  istio-grafana-configuration-dashboards-istio-mesh-dashboard
      - name: dashboards-istio-istio-performance-dashboard
        configMap:
          name:  istio-grafana-configuration-dashboards-istio-performance-dashboard
      - name: dashboards-istio-istio-service-dashboard
        configMap:
          name:  istio-grafana-configuration-dashboards-istio-service-dashboard
      - name: dashboards-istio-istio-workload-dashboard
        configMap:
          name:  istio-grafana-configuration-dashboards-istio-workload-dashboard
      - name: dashboards-istio-mixer-dashboard
        configMap:
          name:  istio-grafana-configuration-dashboards-mixer-dashboard
      - name: dashboards-istio-pilot-dashboard
        configMap:
          name:  istio-grafana-configuration-dashboards-pilot-dashboard

---
# Source: istio/charts/kiali/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kiali
  labels:
    app: kiali
    chart: kiali
    heritage: Tiller
    release: istio
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kiali
  template:
    metadata:
      name: kiali
      labels:
        app: kiali
        chart: kiali
        heritage: Tiller
        release: istio
      annotations:
        sidecar.istio.io/inject: "false"
        scheduler.alpha.kubernetes.io/critical-pod: ""
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
    spec:
      serviceAccountName: kiali-service-account
      containers:
      - image: "docker.io/kiali/kiali:v0.16"
        name: kiali
        command:
        - "/opt/kiali/kiali"
        - "-config"
        - "/kiali-configuration/config.yaml"
        - "-v"
        - "4"
        env:
        - name: ACTIVE_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: PROMETHEUS_SERVICE_URL
          value: http://prometheus:9090
        - name: SERVER_WEB_ROOT
          value: /kiali
        volumeMounts:
        - name: kiali-configuration
          mountPath: "/kiali-configuration"
        - name: kiali-secret
          mountPath: "/kiali-secret"
        resources:
          requests:
            cpu: 10m
          
      volumes:
      - name: kiali-configuration
        configMap:
          name: kiali
      - name: kiali-secret
        secret:
          secretName: kiali
          optional: true
      affinity:      
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x      

---
# Source: istio/charts/mixer/templates/deployment.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  name: istio-policy
  labels:
    app: istio-mixer
    chart: mixer
    heritage: Tiller
    release: istio
    istio: mixer
spec:
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      istio: mixer
      istio-mixer-type: policy
  template:
    metadata:
      labels:
        app: policy
        chart: mixer
        heritage: Tiller
        release: istio
        istio: mixer
        istio-mixer-type: policy
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      serviceAccountName: istio-mixer-service-account
      volumes:
      - name: istio-certs
        secret:
          secretName: istio.istio-mixer-service-account
          optional: true
      - name: uds-socket
        emptyDir: {}
      - name: policy-adapter-secret
        secret:
          secretName: policy-adapter-secret
          optional: true
      affinity:      
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x      
      containers:
      - name: mixer
        image: "docker.io/istio/mixer:1.1.6"
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 15014
        - containerPort: 42422
        args:
          - --monitoringPort=15014
          - --address
          - unix:///sock/mixer.socket
          - --log_output_level=default:info
          - --configStoreURL=mcp://istio-galley.istio-system.svc:9901
          - --configDefaultNamespace=istio-system
          - --useAdapterCRDs=true
          - --trace_zipkin_url=http://zipkin:9411/api/v1/spans
        env:
        - name: GODEBUG
          value: "gctrace=1"
        - name: GOMAXPROCS
          value: "6"
        resources:
          limits:
            cpu: 100m
            memory: 100Mi
          requests:
            cpu: 10m
            memory: 100Mi
          
        volumeMounts:
        - name: istio-certs
          mountPath: /etc/certs
          readOnly: true
        - name: uds-socket
          mountPath: /sock
        livenessProbe:
          httpGet:
            path: /version
            port: 15014
          initialDelaySeconds: 5
          periodSeconds: 5
      - name: istio-proxy
        image: "docker.io/istio/proxyv2:1.1.6"
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 9091
        - containerPort: 15004
        - containerPort: 15090
          protocol: TCP
          name: http-envoy-prom
        args:
        - proxy
        - --domain
        - $(POD_NAMESPACE).svc.cluster.local
        - --serviceCluster
        - istio-policy
        - --templateFile
        - /etc/istio/proxy/envoy_policy.yaml.tmpl
        - --controlPlaneAuthPolicy
        - NONE
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: INSTANCE_IP
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.podIP
        resources:
          limits:
            cpu: 2000m
            memory: 128Mi
          requests:
            cpu: 10m
            memory: 40Mi
          
        volumeMounts:
        - name: istio-certs
          mountPath: /etc/certs
          readOnly: true
        - name: uds-socket
          mountPath: /sock
        - name: policy-adapter-secret
          mountPath: /var/run/secrets/istio.io/policy/adapter
          readOnly: true

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: istio-telemetry
  labels:
    app: istio-mixer
    chart: mixer
    heritage: Tiller
    release: istio
    istio: mixer
spec:
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      istio: mixer
      istio-mixer-type: telemetry
  template:
    metadata:
      labels:
        app: telemetry
        chart: mixer
        heritage: Tiller
        release: istio
        istio: mixer
        istio-mixer-type: telemetry
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      serviceAccountName: istio-mixer-service-account
      volumes:
      - name: istio-certs
        secret:
          secretName: istio.istio-mixer-service-account
          optional: true
      - name: uds-socket
        emptyDir: {}
      - name: telemetry-adapter-secret
        secret:
          secretName: telemetry-adapter-secret
          optional: true
      affinity:      
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x      
      containers:
      - name: mixer
        image: "docker.io/istio/mixer:1.1.6"
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 15014
        - containerPort: 42422
        args:
          - --monitoringPort=15014
          - --address
          - unix:///sock/mixer.socket
          - --log_output_level=default:info
          - --configStoreURL=mcp://istio-galley.istio-system.svc:9901
          - --configDefaultNamespace=istio-system
          - --useAdapterCRDs=true
          - --trace_zipkin_url=http://zipkin:9411/api/v1/spans
          - --averageLatencyThreshold
          - 100ms
          - --loadsheddingMode
          - enforce
        env:
        - name: GODEBUG
          value: "gctrace=1"
        - name: GOMAXPROCS
          value: "6"
        resources:
          limits:
            cpu: 100m
            memory: 100Mi
          requests:
            cpu: 50m
            memory: 100Mi
          
        volumeMounts:
        - name: istio-certs
          mountPath: /etc/certs
          readOnly: true
        - name: telemetry-adapter-secret
          mountPath: /var/run/secrets/istio.io/telemetry/adapter
          readOnly: true
        - name: uds-socket
          mountPath: /sock
        livenessProbe:
          httpGet:
            path: /version
            port: 15014
          initialDelaySeconds: 5
          periodSeconds: 5
      - name: istio-proxy
        image: "docker.io/istio/proxyv2:1.1.6"
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 9091
        - containerPort: 15004
        - containerPort: 15090
          protocol: TCP
          name: http-envoy-prom
        args:
        - proxy
        - --domain
        - $(POD_NAMESPACE).svc.cluster.local
        - --serviceCluster
        - istio-telemetry
        - --templateFile
        - /etc/istio/proxy/envoy_telemetry.yaml.tmpl
        - --controlPlaneAuthPolicy
        - NONE
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: INSTANCE_IP
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.podIP
        resources:
          limits:
            cpu: 2000m
            memory: 128Mi
          requests:
            cpu: 10m
            memory: 40Mi
          
        volumeMounts:
        - name: istio-certs
          mountPath: /etc/certs
          readOnly: true
        - name: uds-socket
          mountPath: /sock

--- 

---
# Source: istio/charts/pilot/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: istio-pilot
  # TODO: default template doesn't have this, which one is right ?
  labels:
    app: pilot
    chart: pilot
    heritage: Tiller
    release: istio
    istio: pilot
  annotations:
    checksum/config-volume: f8da08b6b8c170dde721efd680270b2901e750d4aa186ebb6c22bef5b78a43f9
spec:
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      istio: pilot
  template:
    metadata:
      labels:
        app: pilot
        chart: pilot
        heritage: Tiller
        release: istio
        istio: pilot
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      serviceAccountName: istio-pilot-service-account
      containers:
        - name: discovery
          image: "docker.io/istio/pilot:1.1.6"
          imagePullPolicy: IfNotPresent
          args:
          - "discovery"
          - --monitoringAddr=:15014
          - --log_output_level=default:info
          - --domain
          - cluster.local
          - --secureGrpcAddr
          - ""
          - --keepaliveMaxServerConnectionAge
          - "30m"
          ports:
          - containerPort: 8080
          - containerPort: 15010
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 30
            timeoutSeconds: 5
          env:
          - name: POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.namespace
          - name: GODEBUG
            value: "gctrace=1"
          - name: PILOT_PUSH_THROTTLE
            value: "100"
          - name: PILOT_TRACE_SAMPLING
            value: "100"
          - name: PILOT_DISABLE_XDS_MARSHALING_TO_ANY
            value: "1"
          resources:
            limits:
              cpu: 100m
              memory: 200Mi
            requests:
              cpu: 10m
              memory: 100Mi
            
          volumeMounts:
          - name: config-volume
            mountPath: /etc/istio/config
          - name: istio-certs
            mountPath: /etc/certs
            readOnly: true
        - name: istio-proxy
          image: "docker.io/istio/proxyv2:1.1.6"
          imagePullPolicy: IfNotPresent
          ports:
          - containerPort: 15003
          - containerPort: 15005
          - containerPort: 15007
          - containerPort: 15011
          args:
          - proxy
          - --domain
          - $(POD_NAMESPACE).svc.cluster.local
          - --serviceCluster
          - istio-pilot
          - --templateFile
          - /etc/istio/proxy/envoy_pilot.yaml.tmpl
          - --controlPlaneAuthPolicy
          - NONE
          env:
          - name: POD_NAME
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.namespace
          - name: INSTANCE_IP
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: status.podIP
          resources:
            limits:
              cpu: 2000m
              memory: 128Mi
            requests:
              cpu: 10m
              memory: 40Mi
            
          volumeMounts:
          - name: istio-certs
            mountPath: /etc/certs
            readOnly: true
      volumes:
      - name: config-volume
        configMap:
          name: istio
      - name: istio-certs
        secret:
          secretName: istio.istio-pilot-service-account
          optional: true
      affinity:      
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x      

---
# Source: istio/charts/prometheus/templates/deployment.yaml
# TODO: the original template has service account, roles, etc
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus
  labels:
    app: prometheus
    chart: prometheus
    heritage: Tiller
    release: istio
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      labels:
        app: prometheus
        chart: prometheus
        heritage: Tiller
        release: istio
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      serviceAccountName: prometheus
      containers:
        - name: prometheus
          image: "docker.io/prom/prometheus:v2.3.1"
          imagePullPolicy: IfNotPresent
          args:
            - '--storage.tsdb.retention=6h'
            - '--config.file=/etc/prometheus/prometheus.yml'
          ports:
            - containerPort: 9090
              name: http
          livenessProbe:
            httpGet:
              path: /-/healthy
              port: 9090
          readinessProbe:
            httpGet:
              path: /-/ready
              port: 9090
          resources:
            requests:
              cpu: 10m
            
          volumeMounts:
          - name: config-volume
            mountPath: /etc/prometheus
          - mountPath: /etc/istio-certs
            name: istio-certs
      volumes:
      - name: config-volume
        configMap:
          name: prometheus
      - name: istio-certs
        secret:
          defaultMode: 420
          secretName: istio.default
      affinity:      
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x      

---
# Source: istio/charts/security/templates/deployment.yaml
# istio CA watching all namespaces
apiVersion: apps/v1
kind: Deployment
metadata:
  name: istio-citadel
  labels:
    app: security
    chart: security
    heritage: Tiller
    release: istio
    istio: citadel
spec:
  replicas: 1
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: security
      chart: security
      heritage: Tiller
      release: istio
      istio: citadel
  template:
    metadata:
      labels:
        app: security
        chart: security
        heritage: Tiller
        release: istio
        istio: citadel
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      serviceAccountName: istio-citadel-service-account
      containers:
        - name: citadel
          image: "docker.io/istio/citadel:1.1.6"
          imagePullPolicy: IfNotPresent
          args:
            - --append-dns-names=true
            - --grpc-port=8060
            - --grpc-hostname=citadel
            - --citadel-storage-namespace=istio-system
            - --custom-dns-names=istio-pilot-service-account.istio-system:istio-pilot.istio-system
            - --monitoring-port=15014
            - --self-signed-ca=true
          livenessProbe:
            httpGet:
              path: /version
              port: 15014
            initialDelaySeconds: 5
            periodSeconds: 5
          resources:
            requests:
              cpu: 10m
            
      affinity:      
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x      

---
# Source: istio/charts/sidecarInjectorWebhook/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: istio-sidecar-injector
  labels:
    app: sidecarInjectorWebhook
    chart: sidecarInjectorWebhook
    heritage: Tiller
    release: istio
    istio: sidecar-injector
spec:
  replicas: 1
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: sidecarInjectorWebhook
      chart: sidecarInjectorWebhook
      heritage: Tiller
      release: istio
      istio: sidecar-injector
  template:
    metadata:
      labels:
        app: sidecarInjectorWebhook
        chart: sidecarInjectorWebhook
        heritage: Tiller
        release: istio
        istio: sidecar-injector
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      serviceAccountName: istio-sidecar-injector-service-account
      containers:
        - name: sidecar-injector-webhook
          image: "docker.io/istio/sidecar_injector:1.1.6"
          imagePullPolicy: IfNotPresent
          args:
            - --caCertFile=/etc/istio/certs/root-cert.pem
            - --tlsCertFile=/etc/istio/certs/cert-chain.pem
            - --tlsKeyFile=/etc/istio/certs/key.pem
            - --injectConfig=/etc/istio/inject/config
            - --meshConfig=/etc/istio/config/mesh
            - --healthCheckInterval=2s
            - --healthCheckFile=/health
          volumeMounts:
          - name: config-volume
            mountPath: /etc/istio/config
            readOnly: true
          - name: certs
            mountPath: /etc/istio/certs
            readOnly: true
          - name: inject-config
            mountPath: /etc/istio/inject
            readOnly: true
          livenessProbe:
            exec:
              command:
                - /usr/local/bin/sidecar-injector
                - probe
                - --probe-path=/health
                - --interval=4s
            initialDelaySeconds: 4
            periodSeconds: 4
          readinessProbe:
            exec:
              command:
                - /usr/local/bin/sidecar-injector
                - probe
                - --probe-path=/health
                - --interval=4s
            initialDelaySeconds: 4
            periodSeconds: 4
          resources:
            requests:
              cpu: 10m
            
      volumes:
      - name: config-volume
        configMap:
          name: istio
      - name: certs
        secret:
          secretName: istio.istio-sidecar-injector-service-account
      - name: inject-config
        configMap:
          name: istio-sidecar-injector
          items:
          - key: config
            path: config
      affinity:      
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x      

---
# Source: istio/charts/tracing/templates/deployment-jaeger.yaml


apiVersion: apps/v1
kind: Deployment
metadata:
  name: istio-tracing
  labels:
    app: jaeger
    chart: tracing
    heritage: Tiller
    release: istio
spec:
  selector:
    matchLabels:
      app: jaeger
      chart: tracing
      heritage: Tiller
      release: istio
  template:
    metadata:
      labels:
        app: jaeger
        chart: tracing
        heritage: Tiller
        release: istio
      annotations:
        sidecar.istio.io/inject: "false"
        prometheus.io/scrape: "true"
        prometheus.io/port: "16686"
        prometheus.io/path: "/jaeger/metrics"
    spec:
      containers:
        - name: jaeger
          image: "docker.io/jaegertracing/all-in-one:1.9"
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 9411
            - containerPort: 16686
            - containerPort: 5775
              protocol: UDP
            - containerPort: 6831
              protocol: UDP
            - containerPort: 6832
              protocol: UDP
          env:
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                apiVersion: v1
                fieldPath: metadata.namespace
          - name: COLLECTOR_ZIPKIN_HTTP_PORT
            value: "9411"
          - name: MEMORY_MAX_TRACES
            value: "50000"
          - name: QUERY_BASE_PATH
            value:  /jaeger 
          livenessProbe:
            httpGet:
              path: /
              port: 16686
          readinessProbe:
            httpGet:
              path: /
              port: 16686
          resources:
            requests:
              cpu: 10m
            
      affinity:      
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
                - ppc64le
                - s390x
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - amd64
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - ppc64le
          - weight: 2
            preference:
              matchExpressions:
              - key: beta.kubernetes.io/arch
                operator: In
                values:
                - s390x      


---
# Source: istio/charts/gateways/templates/autoscale.yaml

apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
  name: istio-egressgateway
  labels:
    app: egressgateway
    chart: gateways
    heritage: Tiller
    release: istio
spec:
  maxReplicas: 5
  minReplicas: 1
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: istio-egressgateway
  metrics:
    - type: Resource
      resource:
        name: cpu
        targetAverageUtilization: 80
---
apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
  name: istio-ingressgateway
  labels:
    app: ingressgateway
    chart: gateways
    heritage: Tiller
    release: istio
spec:
  maxReplicas: 5
  minReplicas: 1
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: istio-ingressgateway
  metrics:
    - type: Resource
      resource:
        name: cpu
        targetAverageUtilization: 80
---

---
# Source: istio/charts/mixer/templates/autoscale.yaml

apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
  name: istio-policy
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
    maxReplicas: 5
    minReplicas: 1
    scaleTargetRef:
      apiVersion: apps/v1
      kind: Deployment
      name: istio-policy
    metrics:
    - type: Resource
      resource:
        name: cpu
        targetAverageUtilization: 80
---
apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
  name: istio-telemetry
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
    maxReplicas: 5
    minReplicas: 1
    scaleTargetRef:
      apiVersion: apps/v1
      kind: Deployment
      name: istio-telemetry
    metrics:
    - type: Resource
      resource:
        name: cpu
        targetAverageUtilization: 80
---

---
# Source: istio/charts/pilot/templates/autoscale.yaml

apiVersion: autoscaling/v2beta1
kind: HorizontalPodAutoscaler
metadata:
  name: istio-pilot
  labels:
    app: pilot
    chart: pilot
    heritage: Tiller
    release: istio
spec:
  maxReplicas: 5
  minReplicas: 1
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: istio-pilot
  metrics:
  - type: Resource
    resource:
      name: cpu
      targetAverageUtilization: 80
---

---
# Source: istio/charts/tracing/templates/service-jaeger.yaml

apiVersion: v1
kind: Service
metadata:
  name: jaeger-query
  annotations:
  labels:
    app: jaeger
    jaeger-infra: jaeger-service
    chart: tracing
    heritage: Tiller
    release: istio
spec:
  ports:
    - name: query-http
      port: 16686
      protocol: TCP
      targetPort: 16686
  selector:
    app: jaeger

---

apiVersion: v1
kind: Service
metadata:
  name: jaeger-collector
  labels:
    app: jaeger
    jaeger-infra: collector-service
    chart: tracing
    heritage: Tiller
    release: istio
spec:
  ports:
  - name: jaeger-collector-tchannel
    port: 14267
    protocol: TCP
    targetPort: 14267
  - name: jaeger-collector-http
    port: 14268
    targetPort: 14268
    protocol: TCP
  selector:
    app: jaeger
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  name: jaeger-agent
  labels:
    app: jaeger
    jaeger-infra: agent-service
    chart: tracing
    heritage: Tiller
    release: istio
spec:
  ports:
  - name: agent-zipkin-thrift
    port: 5775
    protocol: UDP
    targetPort: 5775
  - name: agent-compact
    port: 6831
    protocol: UDP
    targetPort: 6831
  - name: agent-binary
    port: 6832
    protocol: UDP
    targetPort: 6832
  clusterIP: None
  selector:
    app: jaeger



---
# Source: istio/charts/tracing/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: zipkin
  labels:
    app: jaeger
    chart: tracing
    heritage: Tiller
    release: istio
spec:
  type: ClusterIP
  ports:
    - port: 9411
      targetPort: 9411
      protocol: TCP
      name: http
  selector:
    app: jaeger
--- 
apiVersion: v1
kind: Service
metadata:
  name: tracing
  annotations:
  labels:
    app: jaeger
    chart: tracing
    heritage: Tiller
    release: istio
spec:
  ports:
    - name: http-query
      port: 80
      protocol: TCP

      targetPort: 16686

  selector:
    app: jaeger

---
# Source: istio/charts/sidecarInjectorWebhook/templates/mutatingwebhook.yaml
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  name: istio-sidecar-injector
  labels:
    app: sidecarInjectorWebhook
    chart: sidecarInjectorWebhook
    heritage: Tiller
    release: istio
webhooks:
  - name: sidecar-injector.istio.io
    clientConfig:
      service:
        name: istio-sidecar-injector
        namespace: istio-system
        path: "/inject"
      caBundle: ""
    rules:
      - operations: [ "CREATE" ]
        apiGroups: [""]
        apiVersions: ["v1"]
        resources: ["pods"]
    failurePolicy: Fail
    namespaceSelector:
      matchLabels:
        istio-injection: enabled


---
# Source: istio/charts/galley/templates/poddisruptionbudget.yaml

apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: istio-galley
  labels:
    app: galley
    chart: galley
    heritage: Tiller
    release: istio
    istio: galley
spec:

  minAvailable: 1
  selector:
    matchLabels:
      app: galley
      release: istio
      istio: galley

---
# Source: istio/charts/gateways/templates/poddisruptionbudget.yaml

apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: istio-egressgateway
  labels:
    chart: gateways
    heritage: Tiller
    release: istio
    app: istio-egressgateway
    istio: egressgateway
spec:

  minAvailable: 1
  selector:
    matchLabels:
      release: istio
      app: istio-egressgateway
      istio: egressgateway
---
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: istio-ingressgateway
  labels:
    chart: gateways
    heritage: Tiller
    release: istio
    app: istio-ingressgateway
    istio: ingressgateway
spec:

  minAvailable: 1
  selector:
    matchLabels:
      release: istio
      app: istio-ingressgateway
      istio: ingressgateway
---

---
# Source: istio/charts/mixer/templates/poddisruptionbudget.yaml

apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: istio-policy
  labels:
    app: policy
    chart: mixer
    heritage: Tiller
    release: istio
    version: 1.1.0
    istio: mixer
    istio-mixer-type: policy
spec:

  minAvailable: 1
  selector:
    matchLabels:
      app: policy
      release: istio
      istio: mixer
      istio-mixer-type: policy
---
apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: istio-telemetry
  labels:
    app: telemetry
    chart: mixer
    heritage: Tiller
    release: istio
    version: 1.1.0
    istio: mixer
    istio-mixer-type: telemetry
spec:

  minAvailable: 1
  selector:
    matchLabels:
      app: telemetry
      release: istio
      istio: mixer
      istio-mixer-type: telemetry
---
# Source: istio/charts/pilot/templates/poddisruptionbudget.yaml

apiVersion: policy/v1beta1
kind: PodDisruptionBudget
metadata:
  name: istio-pilot
  labels:
    app: pilot
    chart: pilot
    heritage: Tiller
    release: istio
    istio: pilot
spec:

  minAvailable: 1
  selector:
    matchLabels:
      app: pilot
      release: istio
      istio: pilot

---
apiVersion: "config.istio.io/v1alpha2"
kind: attributemanifest
metadata:
  name: istioproxy
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  attributes:
    origin.ip:
      valueType: IP_ADDRESS
    origin.uid:
      valueType: STRING
    origin.user:
      valueType: STRING
    request.headers:
      valueType: STRING_MAP
    request.id:
      valueType: STRING
    request.host:
      valueType: STRING
    request.method:
      valueType: STRING
    request.path:
      valueType: STRING
    request.url_path:
      valueType: STRING
    request.query_params:
      valueType: STRING_MAP
    request.reason:
      valueType: STRING
    request.referer:
      valueType: STRING
    request.scheme:
      valueType: STRING
    request.total_size:
      valueType: INT64
    request.size:
      valueType: INT64
    request.time:
      valueType: TIMESTAMP
    request.useragent:
      valueType: STRING
    response.code:
      valueType: INT64
    response.duration:
      valueType: DURATION
    response.headers:
      valueType: STRING_MAP
    response.total_size:
      valueType: INT64
    response.size:
      valueType: INT64
    response.time:
      valueType: TIMESTAMP
    response.grpc_status:
      valueType: STRING
    response.grpc_message:
      valueType: STRING
    source.uid:
      valueType: STRING
    source.user: # DEPRECATED
      valueType: STRING
    source.principal:
      valueType: STRING
    destination.uid:
      valueType: STRING
    destination.principal:
      valueType: STRING
    destination.port:
      valueType: INT64
    connection.event:
      valueType: STRING
    connection.id:
      valueType: STRING
    connection.received.bytes:
      valueType: INT64
    connection.received.bytes_total:
      valueType: INT64
    connection.sent.bytes:
      valueType: INT64
    connection.sent.bytes_total:
      valueType: INT64
    connection.duration:
      valueType: DURATION
    connection.mtls:
      valueType: BOOL
    connection.requested_server_name:
      valueType: STRING
    context.protocol:
      valueType: STRING
    context.proxy_error_code:
      valueType: STRING
    context.timestamp:
      valueType: TIMESTAMP
    context.time:
      valueType: TIMESTAMP
    # Deprecated, kept for compatibility
    context.reporter.local:
      valueType: BOOL
    context.reporter.kind:
      valueType: STRING
    context.reporter.uid:
      valueType: STRING
    api.service:
      valueType: STRING
    api.version:
      valueType: STRING
    api.operation:
      valueType: STRING
    api.protocol:
      valueType: STRING
    request.auth.principal:
      valueType: STRING
    request.auth.audiences:
      valueType: STRING
    request.auth.presenter:
      valueType: STRING
    request.auth.claims:
      valueType: STRING_MAP
    request.auth.raw_claims:
      valueType: STRING
    request.api_key:
      valueType: STRING
    rbac.permissive.response_code:
      valueType: STRING
    rbac.permissive.effective_policy_id:
      valueType: STRING
    check.error_code:
      valueType: INT64
    check.error_message:
      valueType: STRING
    check.cache_hit:
      valueType: BOOL
    quota.cache_hit:
      valueType: BOOL

---
apiVersion: "config.istio.io/v1alpha2"
kind: attributemanifest
metadata:
  name: kubernetes
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  attributes:
    source.ip:
      valueType: IP_ADDRESS
    source.labels:
      valueType: STRING_MAP
    source.metadata:
      valueType: STRING_MAP
    source.name:
      valueType: STRING
    source.namespace:
      valueType: STRING
    source.owner:
      valueType: STRING
    source.serviceAccount:
      valueType: STRING
    source.services:
      valueType: STRING
    source.workload.uid:
      valueType: STRING
    source.workload.name:
      valueType: STRING
    source.workload.namespace:
      valueType: STRING
    destination.ip:
      valueType: IP_ADDRESS
    destination.labels:
      valueType: STRING_MAP
    destination.metadata:
      valueType: STRING_MAP
    destination.owner:
      valueType: STRING
    destination.name:
      valueType: STRING
    destination.container.name:
      valueType: STRING
    destination.namespace:
      valueType: STRING
    destination.service.uid:
      valueType: STRING
    destination.service.name:
      valueType: STRING
    destination.service.namespace:
      valueType: STRING
    destination.service.host:
      valueType: STRING
    destination.serviceAccount:
      valueType: STRING
    destination.workload.uid:
      valueType: STRING
    destination.workload.name:
      valueType: STRING
    destination.workload.namespace:
      valueType: STRING
---
apiVersion: "config.istio.io/v1alpha2"
kind: handler
metadata:
  name: stdio
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  compiledAdapter: stdio
  params:
    outputAsJson: true
---
apiVersion: "config.istio.io/v1alpha2"
kind: logentry
metadata:
  name: accesslog
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  severity: '"Info"'
  timestamp: request.time
  variables:
    sourceIp: source.ip | ip("0.0.0.0")
    sourceApp: source.labels["app"] | ""
    sourcePrincipal: source.principal | ""
    sourceName: source.name | ""
    sourceWorkload: source.workload.name | ""
    sourceNamespace: source.namespace | ""
    sourceOwner: source.owner | ""
    destinationApp: destination.labels["app"] | ""
    destinationIp: destination.ip | ip("0.0.0.0")
    destinationServiceHost: destination.service.host | ""
    destinationWorkload: destination.workload.name | ""
    destinationName: destination.name | ""
    destinationNamespace: destination.namespace | ""
    destinationOwner: destination.owner | ""
    destinationPrincipal: destination.principal | ""
    apiClaims: request.auth.raw_claims | ""
    apiKey: request.api_key | request.headers["x-api-key"] | ""
    protocol: request.scheme | context.protocol | "http"
    method: request.method | ""
    url: request.path | ""
    responseCode: response.code | 0
    responseFlags: context.proxy_error_code | ""
    responseSize: response.size | 0
    permissiveResponseCode: rbac.permissive.response_code | "none"
    permissiveResponsePolicyID: rbac.permissive.effective_policy_id | "none"
    requestSize: request.size | 0
    requestId: request.headers["x-request-id"] | ""
    clientTraceId: request.headers["x-client-trace-id"] | ""
    latency: response.duration | "0ms"
    connection_security_policy: conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))
    requestedServerName: connection.requested_server_name | ""
    userAgent: request.useragent | ""
    responseTimestamp: response.time
    receivedBytes: request.total_size | 0
    sentBytes: response.total_size | 0
    referer: request.referer | ""
    httpAuthority: request.headers[":authority"] | request.host | ""
    xForwardedFor: request.headers["x-forwarded-for"] | "0.0.0.0"
    reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")
    grpcStatus: response.grpc_status | ""
    grpcMessage: response.grpc_message | ""
  monitored_resource_type: '"global"'
---
apiVersion: "config.istio.io/v1alpha2"
kind: logentry
metadata:
  name: tcpaccesslog
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  severity: '"Info"'
  timestamp: context.time | timestamp("2017-01-01T00:00:00Z")
  variables:
    connectionEvent: connection.event | ""
    sourceIp: source.ip | ip("0.0.0.0")
    sourceApp: source.labels["app"] | ""
    sourcePrincipal: source.principal | ""
    sourceName: source.name | ""
    sourceWorkload: source.workload.name | ""
    sourceNamespace: source.namespace | ""
    sourceOwner: source.owner | ""
    destinationApp: destination.labels["app"] | ""
    destinationIp: destination.ip | ip("0.0.0.0")
    destinationServiceHost: destination.service.host | ""
    destinationWorkload: destination.workload.name | ""
    destinationName: destination.name | ""
    destinationNamespace: destination.namespace | ""
    destinationOwner: destination.owner | ""
    destinationPrincipal: destination.principal | ""
    protocol: context.protocol | "tcp"
    connectionDuration: connection.duration | "0ms"
    connection_security_policy: conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))
    requestedServerName: connection.requested_server_name | ""
    receivedBytes: connection.received.bytes | 0
    sentBytes: connection.sent.bytes | 0
    totalReceivedBytes: connection.received.bytes_total | 0
    totalSentBytes: connection.sent.bytes_total | 0
    reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")
    responseFlags: context.proxy_error_code | ""
  monitored_resource_type: '"global"'
---
apiVersion: "config.istio.io/v1alpha2"
kind: rule
metadata:
  name: stdio
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  match: context.protocol == "http" || context.protocol == "grpc"
  actions:
  - handler: stdio
    instances:
    - accesslog.logentry
---
apiVersion: "config.istio.io/v1alpha2"
kind: rule
metadata:
  name: stdiotcp
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  match: context.protocol == "tcp"
  actions:
  - handler: stdio
    instances:
    - tcpaccesslog.logentry
---
apiVersion: "config.istio.io/v1alpha2"
kind: metric
metadata:
  name: requestcount
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  value: "1"
  dimensions:
    reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")
    source_workload: source.workload.name | "unknown"
    source_workload_namespace: source.workload.namespace | "unknown"
    source_principal: source.principal | "unknown"
    source_app: source.labels["app"] | "unknown"
    source_version: source.labels["version"] | "unknown"
    destination_workload: destination.workload.name | "unknown"
    destination_workload_namespace: destination.workload.namespace | "unknown"
    destination_principal: destination.principal | "unknown"
    destination_app: destination.labels["app"] | "unknown"
    destination_version: destination.labels["version"] | "unknown"
    destination_service: destination.service.host | "unknown"
    destination_service_name: destination.service.name | "unknown"
    destination_service_namespace: destination.service.namespace | "unknown"
    request_protocol: api.protocol | context.protocol | "unknown"
    response_code: response.code | 200
    response_flags: context.proxy_error_code | "-"
    permissive_response_code: rbac.permissive.response_code | "none"
    permissive_response_policyid: rbac.permissive.effective_policy_id | "none"
    connection_security_policy: conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))
  monitored_resource_type: '"UNSPECIFIED"'
---
apiVersion: "config.istio.io/v1alpha2"
kind: metric
metadata:
  name: requestduration
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  value: response.duration | "0ms"
  dimensions:
    reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")
    source_workload: source.workload.name | "unknown"
    source_workload_namespace: source.workload.namespace | "unknown"
    source_principal: source.principal | "unknown"
    source_app: source.labels["app"] | "unknown"
    source_version: source.labels["version"] | "unknown"
    destination_workload: destination.workload.name | "unknown"
    destination_workload_namespace: destination.workload.namespace | "unknown"
    destination_principal: destination.principal | "unknown"
    destination_app: destination.labels["app"] | "unknown"
    destination_version: destination.labels["version"] | "unknown"
    destination_service: destination.service.host | "unknown"
    destination_service_name: destination.service.name | "unknown"
    destination_service_namespace: destination.service.namespace | "unknown"
    request_protocol: api.protocol | context.protocol | "unknown"
    response_code: response.code | 200
    response_flags: context.proxy_error_code | "-"
    permissive_response_code: rbac.permissive.response_code | "none" 
    permissive_response_policyid: rbac.permissive.effective_policy_id | "none"
    connection_security_policy: conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))
  monitored_resource_type: '"UNSPECIFIED"'
---
apiVersion: "config.istio.io/v1alpha2"
kind: metric
metadata:
  name: requestsize
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  value: request.size | 0
  dimensions:
    reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")
    source_workload: source.workload.name | "unknown"
    source_workload_namespace: source.workload.namespace | "unknown"
    source_principal: source.principal | "unknown"
    source_app: source.labels["app"] | "unknown"
    source_version: source.labels["version"] | "unknown"
    destination_workload: destination.workload.name | "unknown"
    destination_workload_namespace: destination.workload.namespace | "unknown"
    destination_principal: destination.principal | "unknown"
    destination_app: destination.labels["app"] | "unknown"
    destination_version: destination.labels["version"] | "unknown"
    destination_service: destination.service.host | "unknown"
    destination_service_name: destination.service.name | "unknown"
    destination_service_namespace: destination.service.namespace | "unknown"
    request_protocol: api.protocol | context.protocol | "unknown"
    response_code: response.code | 200
    response_flags: context.proxy_error_code | "-"
    permissive_response_code: rbac.permissive.response_code | "none" 
    permissive_response_policyid: rbac.permissive.effective_policy_id | "none"
    connection_security_policy: conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))
  monitored_resource_type: '"UNSPECIFIED"'
---
apiVersion: "config.istio.io/v1alpha2"
kind: metric
metadata:
  name: responsesize
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  value: response.size | 0
  dimensions:
    reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")
    source_workload: source.workload.name | "unknown"
    source_workload_namespace: source.workload.namespace | "unknown"
    source_principal: source.principal | "unknown"
    source_app: source.labels["app"] | "unknown"
    source_version: source.labels["version"] | "unknown"
    destination_workload: destination.workload.name | "unknown"
    destination_workload_namespace: destination.workload.namespace | "unknown"
    destination_principal: destination.principal | "unknown"
    destination_app: destination.labels["app"] | "unknown"
    destination_version: destination.labels["version"] | "unknown"
    destination_service: destination.service.host | "unknown"
    destination_service_name: destination.service.name | "unknown"
    destination_service_namespace: destination.service.namespace | "unknown"
    request_protocol: api.protocol | context.protocol | "unknown"
    response_code: response.code | 200
    response_flags: context.proxy_error_code | "-"
    permissive_response_code: rbac.permissive.response_code | "none" 
    permissive_response_policyid: rbac.permissive.effective_policy_id | "none"
    connection_security_policy: conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))
  monitored_resource_type: '"UNSPECIFIED"'
---
apiVersion: "config.istio.io/v1alpha2"
kind: metric
metadata:
  name: tcpbytesent
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  value: connection.sent.bytes | 0
  dimensions:
    reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")
    source_workload: source.workload.name | "unknown"
    source_workload_namespace: source.workload.namespace | "unknown"
    source_principal: source.principal | "unknown"
    source_app: source.labels["app"] | "unknown"
    source_version: source.labels["version"] | "unknown"
    destination_workload: destination.workload.name | "unknown"
    destination_workload_namespace: destination.workload.namespace | "unknown"
    destination_principal: destination.principal | "unknown"
    destination_app: destination.labels["app"] | "unknown"
    destination_version: destination.labels["version"] | "unknown"
    destination_service: destination.service.host | "unknown"
    destination_service_name: destination.service.name | "unknown"
    destination_service_namespace: destination.service.namespace | "unknown"
    connection_security_policy: conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))
    response_flags: context.proxy_error_code | "-"
  monitored_resource_type: '"UNSPECIFIED"'
---
apiVersion: "config.istio.io/v1alpha2"
kind: metric
metadata:
  name: tcpbytereceived
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  value: connection.received.bytes | 0
  dimensions:
    reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")
    source_workload: source.workload.name | "unknown"
    source_workload_namespace: source.workload.namespace | "unknown"
    source_principal: source.principal | "unknown"
    source_app: source.labels["app"] | "unknown"
    source_version: source.labels["version"] | "unknown"
    destination_workload: destination.workload.name | "unknown"
    destination_workload_namespace: destination.workload.namespace | "unknown"
    destination_principal: destination.principal | "unknown"
    destination_app: destination.labels["app"] | "unknown"
    destination_version: destination.labels["version"] | "unknown"
    destination_service: destination.service.host | "unknown"
    destination_service_name: destination.service.name | "unknown"
    destination_service_namespace: destination.service.namespace | "unknown"
    connection_security_policy: conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))
    response_flags: context.proxy_error_code | "-"
  monitored_resource_type: '"UNSPECIFIED"'
---
apiVersion: "config.istio.io/v1alpha2"
kind: metric
metadata:
  name: tcpconnectionsopened
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  value: "1"
  dimensions:
    reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")
    source_workload: source.workload.name | "unknown"
    source_workload_namespace: source.workload.namespace | "unknown"
    source_principal: source.principal | "unknown"
    source_app: source.labels["app"] | "unknown"
    source_version: source.labels["version"] | "unknown"
    destination_workload: destination.workload.name | "unknown"
    destination_workload_namespace: destination.workload.namespace | "unknown"
    destination_principal: destination.principal | "unknown"
    destination_app: destination.labels["app"] | "unknown"
    destination_version: destination.labels["version"] | "unknown"
    destination_service: destination.service.name | "unknown"
    destination_service_name: destination.service.name | "unknown"
    destination_service_namespace: destination.service.namespace | "unknown"
    connection_security_policy: conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))
    response_flags: context.proxy_error_code | "-"
  monitored_resource_type: '"UNSPECIFIED"'
---
apiVersion: "config.istio.io/v1alpha2"
kind: metric
metadata:
  name: tcpconnectionsclosed
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  value: "1"
  dimensions:
    reporter: conditional((context.reporter.kind | "inbound") == "outbound", "source", "destination")
    source_workload: source.workload.name | "unknown"
    source_workload_namespace: source.workload.namespace | "unknown"
    source_principal: source.principal | "unknown"
    source_app: source.labels["app"] | "unknown"
    source_version: source.labels["version"] | "unknown"
    destination_workload: destination.workload.name | "unknown"
    destination_workload_namespace: destination.workload.namespace | "unknown"
    destination_principal: destination.principal | "unknown"
    destination_app: destination.labels["app"] | "unknown"
    destination_version: destination.labels["version"] | "unknown"
    destination_service: destination.service.name | "unknown"
    destination_service_name: destination.service.name | "unknown"
    destination_service_namespace: destination.service.namespace | "unknown"
    connection_security_policy: conditional((context.reporter.kind | "inbound") == "outbound", "unknown", conditional(connection.mtls | false, "mutual_tls", "none"))
    response_flags: context.proxy_error_code | "-"
  monitored_resource_type: '"UNSPECIFIED"'
---
apiVersion: "config.istio.io/v1alpha2"
kind: handler
metadata:
  name: prometheus
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  compiledAdapter: prometheus
  params:
    metricsExpirationPolicy:
      metricsExpiryDuration: "10m"
    metrics:
    - name: requests_total
      instance_name: requestcount.metric.istio-system
      kind: COUNTER
      label_names:
      - reporter
      - source_app
      - source_principal
      - source_workload
      - source_workload_namespace
      - source_version
      - destination_app
      - destination_principal
      - destination_workload
      - destination_workload_namespace
      - destination_version
      - destination_service
      - destination_service_name
      - destination_service_namespace
      - request_protocol
      - response_code
      - response_flags
      - permissive_response_code
      - permissive_response_policyid
      - connection_security_policy
    - name: request_duration_seconds
      instance_name: requestduration.metric.istio-system
      kind: DISTRIBUTION
      label_names:
      - reporter
      - source_app
      - source_principal
      - source_workload
      - source_workload_namespace
      - source_version
      - destination_app
      - destination_principal
      - destination_workload
      - destination_workload_namespace
      - destination_version
      - destination_service
      - destination_service_name
      - destination_service_namespace
      - request_protocol
      - response_code
      - response_flags
      - permissive_response_code
      - permissive_response_policyid
      - connection_security_policy
      buckets:
        explicit_buckets:
          bounds: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
    - name: request_bytes
      instance_name: requestsize.metric.istio-system
      kind: DISTRIBUTION
      label_names:
      - reporter
      - source_app
      - source_principal
      - source_workload
      - source_workload_namespace
      - source_version
      - destination_app
      - destination_principal
      - destination_workload
      - destination_workload_namespace
      - destination_version
      - destination_service
      - destination_service_name
      - destination_service_namespace
      - request_protocol
      - response_code
      - response_flags
      - permissive_response_code
      - permissive_response_policyid
      - connection_security_policy
      buckets:
        exponentialBuckets:
          numFiniteBuckets: 8
          scale: 1
          growthFactor: 10
    - name: response_bytes
      instance_name: responsesize.metric.istio-system
      kind: DISTRIBUTION
      label_names:
      - reporter
      - source_app
      - source_principal
      - source_workload
      - source_workload_namespace
      - source_version
      - destination_app
      - destination_principal
      - destination_workload
      - destination_workload_namespace
      - destination_version
      - destination_service
      - destination_service_name
      - destination_service_namespace
      - request_protocol
      - response_code
      - response_flags
      - permissive_response_code
      - permissive_response_policyid
      - connection_security_policy
      buckets:
        exponentialBuckets:
          numFiniteBuckets: 8
          scale: 1
          growthFactor: 10
    - name: tcp_sent_bytes_total
      instance_name: tcpbytesent.metric.istio-system
      kind: COUNTER
      label_names:
      - reporter
      - source_app
      - source_principal
      - source_workload
      - source_workload_namespace
      - source_version
      - destination_app
      - destination_principal
      - destination_workload
      - destination_workload_namespace
      - destination_version
      - destination_service
      - destination_service_name
      - destination_service_namespace
      - connection_security_policy
      - response_flags
    - name: tcp_received_bytes_total
      instance_name: tcpbytereceived.metric.istio-system
      kind: COUNTER
      label_names:
      - reporter
      - source_app
      - source_principal
      - source_workload
      - source_workload_namespace
      - source_version
      - destination_app
      - destination_principal
      - destination_workload
      - destination_workload_namespace
      - destination_version
      - destination_service
      - destination_service_name
      - destination_service_namespace
      - connection_security_policy
      - response_flags
    - name: tcp_connections_opened_total
      instance_name: tcpconnectionsopened.metric.istio-system
      kind: COUNTER
      label_names:
      - reporter
      - source_app
      - source_principal
      - source_workload
      - source_workload_namespace
      - source_version
      - destination_app
      - destination_principal
      - destination_workload
      - destination_workload_namespace
      - destination_version
      - destination_service
      - destination_service_name
      - destination_service_namespace
      - connection_security_policy
      - response_flags
    - name: tcp_connections_closed_total
      instance_name: tcpconnectionsclosed.metric.istio-system
      kind: COUNTER
      label_names:
      - reporter
      - source_app
      - source_principal
      - source_workload
      - source_workload_namespace
      - source_version
      - destination_app
      - destination_principal
      - destination_workload
      - destination_workload_namespace
      - destination_version
      - destination_service
      - destination_service_name
      - destination_service_namespace
      - connection_security_policy
      - response_flags
---
apiVersion: "config.istio.io/v1alpha2"
kind: rule
metadata:
  name: promhttp
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  match: (context.protocol == "http" || context.protocol == "grpc") && (match((request.useragent | "-"), "kube-probe*") == false)
  actions:
  - handler: prometheus
    instances:
    - requestcount.metric
    - requestduration.metric
    - requestsize.metric
    - responsesize.metric
---
apiVersion: "config.istio.io/v1alpha2"
kind: rule
metadata:
  name: promtcp
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  match: context.protocol == "tcp"
  actions:
  - handler: prometheus
    instances:
    - tcpbytesent.metric
    - tcpbytereceived.metric
---
apiVersion: "config.istio.io/v1alpha2"
kind: rule
metadata:
  name: promtcpconnectionopen
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  match: context.protocol == "tcp" && ((connection.event | "na") == "open")
  actions:
  - handler: prometheus
    instances:
    - tcpconnectionsopened.metric
---
apiVersion: "config.istio.io/v1alpha2"
kind: rule
metadata:
  name: promtcpconnectionclosed
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  match: context.protocol == "tcp" && ((connection.event | "na") == "close")
  actions:
  - handler: prometheus
    instances:
    - tcpconnectionsclosed.metric
---
apiVersion: "config.istio.io/v1alpha2"
kind: handler
metadata:
  name: kubernetesenv
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  compiledAdapter: kubernetesenv
  params:
    # when running from mixer root, use the following config after adding a
    # symbolic link to a kubernetes config file via:
    #
    # $ ln -s ~/.kube/config mixer/adapter/kubernetes/kubeconfig
    #
    # kubeconfig_path: "mixer/adapter/kubernetes/kubeconfig"

---
apiVersion: "config.istio.io/v1alpha2"
kind: rule
metadata:
  name: kubeattrgenrulerule
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  actions:
  - handler: kubernetesenv
    instances:
    - attributes.kubernetes
---
apiVersion: "config.istio.io/v1alpha2"
kind: rule
metadata:
  name: tcpkubeattrgenrulerule
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  match: context.protocol == "tcp"
  actions:
  - handler: kubernetesenv
    instances:
    - attributes.kubernetes
---
apiVersion: "config.istio.io/v1alpha2"
kind: kubernetes
metadata:
  name: attributes
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  # Pass the required attribute data to the adapter
  source_uid: source.uid | ""
  source_ip: source.ip | ip("0.0.0.0") # default to unspecified ip addr
  destination_uid: destination.uid | ""
  destination_port: destination.port | 0
  attribute_bindings:
    # Fill the new attributes from the adapter produced output.
    # $out refers to an instance of OutputTemplate message
    source.ip: $out.source_pod_ip | ip("0.0.0.0")
    source.uid: $out.source_pod_uid | "unknown"
    source.labels: $out.source_labels | emptyStringMap()
    source.name: $out.source_pod_name | "unknown"
    source.namespace: $out.source_namespace | "default"
    source.owner: $out.source_owner | "unknown"
    source.serviceAccount: $out.source_service_account_name | "unknown"
    source.workload.uid: $out.source_workload_uid | "unknown"
    source.workload.name: $out.source_workload_name | "unknown"
    source.workload.namespace: $out.source_workload_namespace | "unknown"
    destination.ip: $out.destination_pod_ip | ip("0.0.0.0")
    destination.uid: $out.destination_pod_uid | "unknown"
    destination.labels: $out.destination_labels | emptyStringMap()
    destination.name: $out.destination_pod_name | "unknown"
    destination.container.name: $out.destination_container_name | "unknown"
    destination.namespace: $out.destination_namespace | "default"
    destination.owner: $out.destination_owner | "unknown"
    destination.serviceAccount: $out.destination_service_account_name | "unknown"
    destination.workload.uid: $out.destination_workload_uid | "unknown"
    destination.workload.name: $out.destination_workload_name | "unknown"
    destination.workload.namespace: $out.destination_workload_namespace | "unknown"
---
# Configuration needed by Mixer.
# Mixer cluster is delivered via CDS
# Specify mixer cluster settings
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: istio-policy
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  host: istio-policy.istio-system.svc.cluster.local
  trafficPolicy:
    connectionPool:
      http:
        http2MaxRequests: 10000
        maxRequestsPerConnection: 10000
---
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: istio-telemetry
  labels:
    app: mixer
    chart: mixer
    heritage: Tiller
    release: istio
spec:
  host: istio-telemetry.istio-system.svc.cluster.local
  trafficPolicy:
    connectionPool:
      http:
        http2MaxRequests: 10000
        maxRequestsPerConnection: 10000
---
`)
  th.writeK("/manifests/istio/istio-install/base", `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- istio-noauth.yaml
namespace: kubeflow
images:
- name: docker.io/istio/kubectl
  newName: docker.io/istio/kubectl
  newTag: 1.1.6
- name: docker.io/istio/galley
  newName: docker.io/istio/galley
  newTag: 1.1.6
- name: docker.io/istio/proxyv2
  newName: docker.io/istio/proxyv2
  newTag: 1.1.6
- name: grafana/grafana
  newName: grafana/grafana
  newTag: 6.0.2
- name: docker.io/kiali/kiali
  newName: docker.io/kiali/kiali
  newTag: v0.16
- name: docker.io/istio/mixer
  newName: docker.io/istio/mixer
  newTag: 1.1.6
- name: docker.io/istio/pilot
  newName: docker.io/istio/pilot
  newTag: 1.1.6
- name: docker.io/prom/prometheus
  newName: docker.io/prom/prometheus
  newTag: v2.3.1
- name: docker.io/istio/citadel
  newName: docker.io/istio/citadel
  newTag: 1.1.6
- name: docker.io/istio/sidecar_injector
  newName: docker.io/istio/sidecar_injector
  newTag: 1.1.6
- name: docker.io/jaegertracing/all-in-one
  newName: docker.io/jaegertracing/all-in-one
  newTag: '1.9'
`)
}

func TestIstioInstallBase(t *testing.T) {
  th := NewKustTestHarness(t, "/manifests/istio/istio-install/base")
  writeIstioInstallBase(th)
  m, err := th.makeKustTarget().MakeCustomizedResMap()
  if err != nil {
    t.Fatalf("Err: %v", err)
  }
  expected, err := m.AsYaml()
  if err != nil {
    t.Fatalf("Err: %v", err)
  }
  targetPath := "../istio/istio-install/base"
  fsys := fs.MakeRealFS()
  lrc := loader.RestrictionRootOnly
  _loader, loaderErr := loader.NewLoader(lrc, validators.MakeFakeValidator(), targetPath, fsys)
  if loaderErr != nil {
    t.Fatalf("could not load kustomize loader: %v", loaderErr)
  }
  rf := resmap.NewFactory(resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()), transformer.NewFactoryImpl())
  pc := plugins.DefaultPluginConfig()
  kt, err := target.NewKustTarget(_loader, rf, transformer.NewFactoryImpl(), plugins.NewLoader(pc, rf))
  if err != nil {
    th.t.Fatalf("Unexpected construction error %v", err)
  }
  actual, err := kt.MakeCustomizedResMap()
  if err != nil {
    t.Fatalf("Err: %v", err)
  }
  th.assertActualEqualsExpected(actual, string(expected))
}
