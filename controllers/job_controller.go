/*
Copyright 2023.

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
	"time"

	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nadavbm/zlog"

	cronjobsv1 "example.com/job/api/v1"
)

// JobReconciler reconciles a Job object
type JobReconciler struct {
	Logger *zlog.Logger
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cronjobs.example.com,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cronjobs.example.com,resources=jobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cronjobs.example.com,resources=jobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Job object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *JobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := zlog.New()
	r.Logger = logger

	var job cronjobsv1.Job
	if err := r.Client.Get(context.Background(), req.NamespacedName, &job); err != nil {
		if errors.IsNotFound(err) {
			r.Logger.Info("job is not found, probably deleted. skipping..", zap.String("namespace", req.Namespace))
			return ctrl.Result{Requeue: false, RequeueAfter: 0}, nil
		}
		r.Logger.Error("could not fetch resource")
		return ctrl.Result{Requeue: true, RequeueAfter: time.Minute}, err
	}

	var cronJob batchv1.CronJob
	if err := r.Get(ctx, req.NamespacedName, &cronJob); err != nil {
		// r.Logger.Error("unable to fetch CronJob", zap.Error(err))
		if errors.IsNotFound(err) {
			r.Logger.Info("create cronjob", zap.String("namespace", req.Namespace))
			cj := buildCronJob(req.Namespace, &job)
			if err := r.Create(ctx, cj); err != nil {
				r.Logger.Error("could not create cronjob")
				return ctrl.Result{}, err
			}
			if err := r.Update(ctx, cj); err != nil {
				if errors.IsInvalid(err) {
					r.Logger.Error("invalid update", zap.String("object", cj.GetName()))
				} else {
					r.Logger.Error("unable to update", zap.String("object", cj.GetName()))
				}
			}
			return ctrl.Result{}, nil
		}

		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return ctrl.Result{}, nil
}

func buildCronJob(ns string, jobSpecs *cronjobsv1.Job) *batchv1.CronJob {
	controller := true
	return &batchv1.CronJob{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CronJob",
			APIVersion: "batch/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "jobop",
			Namespace: ns,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: jobSpecs.APIVersion,
					Kind:       jobSpecs.Kind,
					Name:       jobSpecs.Name,
					UID:        jobSpecs.UID,
					Controller: &controller,
				},
			},
		},
		Spec: batchv1.CronJobSpec{
			Schedule:          jobSpecs.Spec.Schedule,
			ConcurrencyPolicy: batchv1.ForbidConcurrent,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							RestartPolicy: "Never",
							Containers: []v1.Container{
								{
									Name:            "jobop",
									Image:           jobSpecs.Spec.Image,
									Command:         jobSpecs.Spec.Command,
									ImagePullPolicy: "Always",
								},
							},
						},
					},
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *JobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cronjobsv1.Job{}).
		Owns(&batchv1.CronJob{}).
		Complete(r)
}
