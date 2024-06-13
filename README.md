# silentstorm

SilentStorm is a Kubernetes Operator that simplifies the management of silences within your Kubernetes cluster. It
provides Custom Resource Definitions (CRDs) for both Alertmanager and Silences, allowing you to declaratively configure
them and ensure all alerts on an Alertmanager are silenced.

## Installation

### Helm

```console
helm repo add silentstorm https://biggold1310.github.io/silentstorm
helm install silentstorm biggold1310/silentstorm
```

## Usage
### Getting Started
This example will guide you through the process of setting up a basic Alertmanager with a ClusterSilence and Silence.
Conceptually, the Alertmanager selects the ClusterSilences/Silences to be applied over the silenceSelector.

1. Create an Alertmanager instance by applying the following YAML manifest:
   ```yaml
   apiVersion: silentstorm.biggold1310.ch/v1alpha1
   kind: Alertmanager
   metadata:
     name: sample-alertmanager
   spec:
     address: "http://alertmanager-1.alertmanager.svc.cluster.local:9093"
     silenceSelector:
       matchLabels:
         aminstance: "alertmanager-1"
   ```

2. Create a Silence (namespaced)
   ```yaml
   apiVersion: silentstorm.biggold1310.ch/v1alpha1
   kind: Silence
   metadata:
      labels:
         aminstance: alertmanager-1
      name: sample-silence
      namespace: sample-namespace
   spec:
      comment: "Silence all KubePodRestart alerts in sample-namespace."
      creator: "namespaceuser@cluster.local"
      matchers:
         - isEqual: true
           isRegex: false
           name: "alertname"
           value: "KubePodRestart"
   ```

3. Create a ClusterSilence
   ```yaml
   apiVersion: silentstorm.biggold1310.ch/v1alpha1
   kind: ClusterSilence
   metadata:
      labels:
         aminstance: alertmanager-1
      name: sample-clustersilence
   spec:
      comment: "Silence all warning and critical alerts for test environments"
      creator: "clusteradmin@cluster.local"
      matchers:
         - isEqual: true
           isRegex: true
           name: "severity"
           value: "(warning|critical)"
         - isEqual: true
           isRegex: true
           name: "environment"
           value: "(test|engineering|demo)"
   ```

## Known limitations
The current SilenceOperator won't delete managed silences while deleting the Alertmanager resource.

## Contributing

// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

