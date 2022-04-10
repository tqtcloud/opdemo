package resources

import (
	appv1beta1 "github.com/tqtcloud/opdemo/api/v1beta1"
	appv1 "k8s.io/api/apps/v1"
	core1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func MutateDeployment(app *appv1beta1.AppService, deploy *appv1.Deployment) {
	labels := map[string]string{"app": app.Name}
	selector := &metav1.LabelSelector{MatchLabels: labels}
	deploy.Spec = appv1.DeploymentSpec{
		Replicas: app.Spec.Size,
		Template: core1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Labels: labels},
			Spec: core1.PodSpec{
				Containers: NewContainers(app),
			},
		},
		Selector: selector,
	}
}
func MutateService(app *appv1beta1.AppService, svc *core1.Service) {
	svc.Spec = core1.ServiceSpec{
		ClusterIP: svc.Spec.ClusterIP,
		Type:      core1.ServiceTypeNodePort,
		Ports:     app.Spec.Ports,
		Selector: map[string]string{
			"app": app.Name,
		},
	}
}

func NewDeploy(app *appv1beta1.AppService) *appv1.Deployment {
	labels := map[string]string{"app": app.Name}
	selector := &metav1.LabelSelector{MatchLabels: labels}
	return &appv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(app, schema.GroupVersionKind{
					Group:   appv1beta1.GroupVersion.Group,
					Version: appv1beta1.GroupVersion.Version,
					Kind:    appv1beta1.Kind,
				}),
			},
		},
		Spec: appv1.DeploymentSpec{
			Replicas: app.Spec.Size,
			Template: core1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: core1.PodSpec{
					Containers: NewContainers(app),
				},
			},
			Selector: selector,
		},
	}
}

func NewContainers(app *appv1beta1.AppService) []core1.Container {
	containerPort := []core1.ContainerPort{}
	for _, svcPort := range app.Spec.Ports {
		cport := core1.ContainerPort{}
		cport.ContainerPort = svcPort.TargetPort.IntVal
		containerPort = append(containerPort, cport)
	}
	return []core1.Container{
		{
			Name:            app.Name,
			Image:           app.Spec.Image,
			Resources:       app.Spec.Resources,
			Ports:           containerPort,
			ImagePullPolicy: core1.PullIfNotPresent,
			Env:             app.Spec.Envs,
		},
	}
}

func NewService(app *appv1beta1.AppService) *core1.Service {
	return &core1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(app, schema.GroupVersionKind{
					Group:   appv1beta1.GroupVersion.Group,
					Version: appv1beta1.GroupVersion.Version,
					Kind:    appv1beta1.Kind,
				}),
			},
		},
		Spec: core1.ServiceSpec{
			Type:  core1.ServiceTypeNodePort,
			Ports: app.Spec.Ports,
			Selector: map[string]string{
				"app": app.Name,
			},
		},
	}
}
