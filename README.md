# Kubernetes EBS Tagger

This will watch all Persistent Volumes created by the Kubernetes AWS EBS provider and apply the labels from the Persistent Volume and Persistent Volume Claim to the EBS volume.


## Install

Using helm 3

```
$  helm repo add sstarcher https://shanestarcher.com/helm-charts/
$ helm install kube-ebs-tagger sstarcher/kube-ebs-tagger
```

If using kube2iam you will want to specify podAnnotations
```yaml
podAnnotations:
    iam.amazonaws.com/role: IAM_ROLE
```

## Automated Docker Builds

This is handled by docker hub's automated [build process](https://hub.docker.com/repository/docker/sstarcher/kube-ebs-tagger/).
