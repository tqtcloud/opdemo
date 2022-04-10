/*
Copyright 2022 MyApp.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"github.com/tqtcloud/opdemo/resources"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1beta1 "github.com/tqtcloud/opdemo/api/v1beta1"
)

//var oldSpecAnnotation = "old/spec"

// AppServiceReconciler reconciles a AppService object
type AppServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=app.example.org,resources=appservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.example.org,resources=appservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.example.org,resources=appservices/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AppService object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *AppServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	// TODO(user): your logic here

	// 业务逻辑实现
	// 获取 AppService 实例
	var appService appv1beta1.AppService
	if err := r.Client.Get(ctx, req.NamespacedName, &appService); err != nil {
		// myapp 被删除忽略
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	//if err != nil {
	//	// Myapp 被删除的时候忽略
	//	if client.IgnoreNotFound(err) != nil {
	//		return ctrl.Result{}, err
	//	}
	//	return ctrl.Result{}, err
	//}

	// 得到 myapp 去创建对应的deployment 和 service  （核心就是观察当前状态和期望状态）

	klog.Info("fetch appservice objects", "appservice", appService)

	//通过协调的方式部署 Deployment
	var deploy appv1.Deployment
	deploy.Name = appService.Name
	deploy.Namespace = appService.Namespace
	or, err := ctrl.CreateOrUpdate(ctx, r.Client, &deploy, func() error {
		//协调必须在这个函数中实现
		resources.MutateDeployment(&appService, &deploy)
		return controllerutil.SetControllerReference(&appService, &deploy, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	klog.Info("CreateOrUpdate", "Deployment", or)

	//  CreateOrUpdate Service
	var svc corev1.Service
	svc.Namespace = appService.Namespace
	svc.Name = appService.Name
	or, err = ctrl.CreateOrUpdate(ctx, r.Client, &svc, func() error {
		//协调必须在这个函数中实现
		resources.MutateService(&appService, &svc)
		return controllerutil.SetControllerReference(&appService, &svc, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	klog.Info("CreateOrUpdate", "Service", or)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1beta1.AppService{}).
		Owns(&appv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
