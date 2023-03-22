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
	"fmt"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/tools/record"
	"net/http"
	"net/url"
	opendatav1alpha1 "operator-learn/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"
)

// CryptoReconciler reconciles a Crypto object
type CryptoReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=opendata.example.com,resources=cryptoes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=opendata.example.com,resources=cryptoes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=opendata.example.com,resources=cryptoes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Crypto object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *CryptoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.FromContext(ctx)

	// TODO(user): your logic here
	fmt.Println("terry req", req.Name, req.NamespacedName, req.Namespace)

	instance := &opendatav1alpha1.Crypto{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		reqLogger.Error(err, err.Error())
		return ctrl.Result{}, err
	}

	urlPath, err := url.Parse("https://api.binance.com/api/v3/ticker/price")
	if err != nil {
		return ctrl.Result{}, err
	}
	q := urlPath.Query()
	q.Add("symbol", instance.Spec.Symbol)
	urlPath.RawQuery = q.Encode()

	get, err := http.DefaultClient.Get(urlPath.String())
	if err != nil {
		return ctrl.Result{}, err
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	body, err := io.ReadAll(get.Body)
	if err != nil {
		return ctrl.Result{}, err
	}

	data := struct {
		Price string `json:"price"`
	}{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return ctrl.Result{}, err
	}

	instance.Status = opendatav1alpha1.CryptoStatus{
		State:     "Completed",
		Price:     data.Price,
		UpdatedAt: metav1.Now(),
	}

	err = r.Status().Update(context.Background(), instance)
	if err != nil {
		return ctrl.Result{}, err
	}
	if err != nil {
		return ctrl.Result{}, err
	}
	if err != nil {
		reqLogger.Error(err, err.Error())
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true, RequeueAfter: time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CryptoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opendatav1alpha1.Crypto{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}
