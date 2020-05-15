package apicurioregistry

import (
	"context"
	"errors"
	ar "github.com/Apicurio/apicurio-registry-operator/pkg/apis/apicur/v1alpha1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ reconcile.Reconciler = &ApicurioRegistryReconciler{}

type ApicurioRegistryReconciler struct {
	client           client.Client
	scheme           *runtime.Scheme
	controller       controller.Controller
	contexts         map[string]*Context
}

func NewApicurioRegistryReconciler(mgr manager.Manager) *ApicurioRegistryReconciler {

	return &ApicurioRegistryReconciler{
		client:           mgr.GetClient(),
		scheme:           mgr.GetScheme(),
		contexts: make(map[string]*Context),
	}
}

func (this *ApicurioRegistryReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("ApicurioRegistryReconciler executing.")

	app := request.Name

	// GetConfig the spec
	specList := &ar.ApicurioRegistryList{}
	listOps := client.ListOptions{Namespace: request.Namespace}
	err := this.client.List(context.TODO(), &listOps, specList)
	if err != nil {
		if api_errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	var spec *ar.ApicurioRegistry = nil

	for i,specItem := range specList.Items {

		key := specItem.Name

		c, ok := this.contexts[key]
		if !ok {
			reqLogger.Info("Creating new context")
			c = NewContext(this.controller, this.scheme, reqLogger.WithValues("app", key), this.client)

			var f ControlFunction
			f = NewDeploymentCF(c)
			c.AddControlFunction(f)

			f = NewServiceCF(c)
			c.AddControlFunction(f)

			f = NewIngressCF(c)
			c.AddControlFunction(f)

			f = NewImageConfigCF(c)
			c.AddControlFunction(f)

			f = NewConfReplicasCF(c)
			c.AddControlFunction(f)

			f = NewHostConfigCF(c)
			c.AddControlFunction(f)

			f = NewEnvCF(c)
			c.AddControlFunction(f)

			this.contexts[key] = c
		}

		if app == key {
			spec = &specList.Items[i] // Do not use spec = &specItem
		}
	}

	if spec == nil {
		_, ok := this.contexts[app];
		if ok {
			reqLogger.WithValues("app", app).Info("Deleting context")
			delete(this.contexts, app);
		}
		return reconcile.Result{}, nil
	}

	ctx := this.contexts[app]

	// Context update
	ctx.Update(spec)

	// GetConfig possible config errors
	if errs := ctx.GetConfiguration().GetErrors(); len(*errs) > 0 {
		for _, v := range *errs {
			err := errors.New(v)
			ctx.GetLog().Error(err, v)
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// The LOOP
	requeue := false
	for _, v := range ctx.GetControlFunctions() {
		err = v.Sense(spec, request)
		if err != nil {
			ctx.GetLog().Error(err, "Error during the SENSE phase of '"+v.Describe()+"' CF.")
			requeue = true
			continue
		}
		var discrepancy bool
		discrepancy, err = v.Compare(spec)
		if err != nil {
			ctx.GetLog().Error(err, "Error during the COMPARE phase of '"+v.Describe()+"' CF.")
			requeue = true
			continue
		}
		if !discrepancy {
			continue
		}
		var changed bool
		changed, err = v.Respond(spec)
		if changed {
			requeue = true
		}
		if err != nil {
			ctx.GetLog().Error(err, "Error during the RESPOND phase of '"+v.Describe()+"' CF.")
			requeue = true
			continue
		}
	}

	// Update the status
	spec = ctx.GetFactory().CreateSpec(spec)
	err = this.client.Status().Update(context.TODO(), spec)
	if err != nil {
		ctx.GetLog().Error(err, "Error updating status")
		return reconcile.Result{}, err
	}

	// Run patcher
	ctx.GetPatcher().Execute()

	return reconcile.Result{Requeue: requeue}, nil // err
}



func (this *ApicurioRegistryReconciler) setController(c controller.Controller) {
	this.controller = c
}