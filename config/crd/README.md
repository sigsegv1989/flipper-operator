# RollingUpdate CRD Documentation

## Overview
The RollingUpdate CRD defines a custom resource used to manage rolling restart or update configurations for Kubernetes deployments.

## API Version
- **Group:** flipper.example.com
- **Version:** v1alpha1
- **Kind:** RollingUpdate

## Spec Fields

### matchLabels
- **Type:** object
- **Description:** Specifies a set of {key, value} pairs used to select specific resources based on labels. If specified, only resources matching all key-value pairs will be considered for rollout. If not specified, all resources in the namespace will be considered for rollout.
Each {key, value} pair in MatchLabels is equivalent to a label selector requirement using the "In" operator, where the requirement's key field matches the key, the operator is "In", and the values array contains only the value. The requirements are ANDed together.
- **Optional:** Yes
- **Example:**
  ```yaml
  matchLabels:
    app: nginx
    tier: frontend
### interval
- **Type:** string
- **Description:** Specifies the time interval between rollouts. If not specified, defaults to "24h".
Must be a valid duration string (e.g., "12h", "30m").
- **Optional:** Yes
- **Example:** "12h"

## Status Fields

### lastRolloutTime
- **Type:** string (date-time format)
- **Description:** Indicates the timestamp of the last rolling restart or rollout operation performed by this RollingUpdate CR. If not set, indicates that no rolling restart or rollout has been performed yet.
- **Example:** "2024-06-18T12:00:00Z"

### deployments
- **Type:** array of strings
- **Description:** Stores the names of deployments that were restarted by this RollingUpdate CR. Allows for back tracing to identify which deployments were affected by a particular rolling restart or rollout operation.
- **Example:**
  ```yaml
  deployments:
    - nginx-deployment
    - mysql-deployment
## Sample YAML for Creating a RollingUpdate CR
```yaml
apiVersion: flipper.example.com/v1alpha1
kind: RollingUpdate
metadata:
  name: rollingupdate-sample
spec:
  interval: "12h"
  matchLabels:
    app: nginx
    tier: frontend